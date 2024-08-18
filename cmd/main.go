package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"jpkg/cache"
	"jpkg/config"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/schollz/progressbar/v3"
)

type Asset struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	DonwloadUrl string `json:"browser_download_url"`
}

type GithubRes struct {
	Assets []Asset `json:"assets"`
}

func getJavaFiles(srcDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".java") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func getJarFiles(libDir string) (string, error) {
	var jars []string
	err := filepath.Walk(libDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jar") {
			jars = append(jars, path)
		}
		return nil
	})
	return strings.Join(jars, string(os.PathListSeparator)), err
}

func compileJava(srcDir, binDir, libDir string) error {
	javaFiles, err := getJavaFiles(srcDir)
	if err != nil {
		return err
	}

	var jarFiles string

	// Check if the lib directory exists
	if _, err := os.Stat(libDir); err == nil {
		jarFiles, err = getJarFiles(libDir)
		if err != nil {
			return err
		}
	}

	// Ensure the binDir and cacheDir directories exist
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		err := os.MkdirAll(binDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Ensure the cacheDir exists (if required)
	cacheDir := ".jpkg/cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err := os.MkdirAll(cacheDir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Construct javac arguments
	var args []string
	if jarFiles != "" {
		args = append(args, "-cp", jarFiles)
	}
	args = append(args, "-d", binDir)
	args = append(args, javaFiles...)

	cmd := exec.Command("javac", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createJar(binDir, jarFileName, mainClass, libDir string) error {
	// Ensure the .jpkg/build directory exists
	buildDir := filepath.Join(".jpkg", "build")
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		if err := os.MkdirAll(buildDir, os.ModePerm); err != nil {
			return err
		}
	}

	// Get the JAR files for the classpath
	jarFiles, err := getJarFiles(libDir)
	if err != nil {
		return err
	}

	// Convert JAR file paths to relative paths for the manifest
	relJarFiles := strings.ReplaceAll(jarFiles, string(os.PathSeparator), "/")
	relJarFilesList := strings.Split(relJarFiles, string(os.PathListSeparator))

	// Create a temporary manifest file
	manifestFile := filepath.Join(binDir, "MANIFEST.MF")
	manifestContent := fmt.Sprintf("Main-Class: %s\nClass-Path: %s\n", mainClass, strings.Join(relJarFilesList, " "))
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		return err
	}
	defer os.Remove(manifestFile)

	// Path to the JAR file in the .jpkg/build directory
	jarFilePath := filepath.Join(buildDir, jarFileName)

	// Create the JAR file using the `jar` command with the manifest
	cmd := exec.Command("jar", "cmf", manifestFile, jarFilePath, "-C", binDir, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runJava(mainClass, binDir, libDir string) error {
	var classpath string

	// Check if the lib directory exists
	if _, err := os.Stat(libDir); err == nil {
		jarFiles, err := getJarFiles(libDir)
		if err != nil {
			return err
		}
		classpath = fmt.Sprintf("%s%s%s", binDir, string(os.PathListSeparator), jarFiles)
	} else {
		// If libDir does not exist, use only binDir
		classpath = binDir
	}

	cmd := exec.Command("java", "-cp", classpath, mainClass)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func downloadFile(name, url, dest string) error {
	resp, err := http.Head(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	size := resp.ContentLength

	// Create a progress bar
	bar := progressbar.DefaultBytes(size,
		name,
	)

	// Start downloading the file
	resp, err = http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create a writer to update the progress bar
	writer := io.MultiWriter(out, bar)

	// Copy the content to the file while updating the progress bar
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("\nDownload complete")
	return nil
}

func handleMavenURL(url string, libDir string) error {
	// Remove the "pkg:maven/" prefix
	trimmedURL := strings.TrimPrefix(url, "pkg:maven/")
	parts := strings.Split(trimmedURL, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid Maven URL format")
	}

	groupID := strings.ReplaceAll(parts[0], ".", "/")
	artifactVersionStr := strings.Split(parts[1], "@")
	artifactID := artifactVersionStr[0]
	version := artifactVersionStr[1]
	jarFileName := fmt.Sprintf("%s-%s.jar", artifactID, version)
	downloadURL := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s", groupID, artifactID, version, jarFileName)

	// Download the JAR file
	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		os.Mkdir(libDir, os.ModePerm)
	}
	dest := filepath.Join(libDir, jarFileName)

	if err := config.SaveDependency(fmt.Sprintf("%s/%s", groupID, artifactID), "maven", version); err != nil {
		return err
	}

	return downloadFile(artifactID, downloadURL, dest)
}

// Function to handle GitHub URL
func handleGitHubURL(url, libDir string) error {
	// Example: https://github.com/user/repo/releases/latest/download/file.jar
	parts := strings.Split(url, "/")
	if len(parts) < 5 || !strings.HasPrefix(url, "https://github.com") {
		return fmt.Errorf("invalid GitHub URL format")
	}

	// Construct the latest release API URL
	user := parts[3]
	repo := parts[4]
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", user, repo)

	// Fetch the latest release info
	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var data GithubRes
	downloadUrl := ""

	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	for _, asset := range data.Assets {
		if asset.ContentType == "application/java-archive" {
			downloadUrl = asset.DonwloadUrl
		}
	}

	if downloadUrl == "" {
		return errors.New("no jar file available in repo")
	}

	jarDownloadURL := downloadUrl
	jarFileName := filepath.Base(jarDownloadURL)

	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		os.Mkdir(libDir, os.ModePerm)
	}
	dest := filepath.Join(libDir, jarFileName)

	if err := config.SaveDependency(fmt.Sprintf("%s/%s", user, repo), "github", ""); err != nil {
		return err
	}
	return downloadFile(repo, jarDownloadURL, dest)
}

func main() {
	srcDir := "src"
	binDir := ".jpkg/bin"
	libDir := "lib"
	cacheDir := ".jpkg/cache"

	// Parse command-line arguments
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: go run main.go [build|run]")
		return
	}

	if args[0] == "init" {
		if err := config.CreateInitialFiles(); err != nil {
			fmt.Println("Failed to initialize project:", err)
		}
		return
	}

	configs, error := config.GetConfig()
	if error != nil {
		err := errors.New("initialize the project. then try running [jpkg run|jpkg build]")
		fmt.Println(err)
		return
	}

	mainClass := configs.MainClass

	switch args[0] {
	case "build":
		if err := compileJava(srcDir, binDir, libDir); err != nil {
			fmt.Println("Failed to compile:", err)
			return
		}
		if err := createJar(binDir, "app.jar", mainClass, "lib"); err != nil {
			fmt.Println("Failed to create JAR:", err)
		}
	case "run":
		if isUptoDate, err := cache.IsCacheUpToDate(srcDir, cacheDir); err == nil && isUptoDate {
			if err := runJava(mainClass, binDir, libDir); err != nil {
				fmt.Println("Failed to run:", err)
			}
		} else {
			cache.CopySrcToCache(srcDir, cacheDir)
			if err := compileJava(srcDir, binDir, libDir); err != nil {
				fmt.Println("Failed to compile:", err)
				return
			}
			if err := runJava(mainClass, binDir, libDir); err != nil {
				fmt.Println("Failed to run:", err)
			}
		}
	case "install":
		if len(args) < 2 {
			if _, err := os.Stat("amber.toml"); os.IsNotExist(err) {
				fmt.Println("amber.toml file not found")
				return
			}

			var config map[string]map[string]map[string]string
			if _, err := toml.DecodeFile("amber.toml", &config); err != nil {
				fmt.Println("Failed to load config:", err)
				return
			}

			for dep, info := range config["dependencies"] {
				origin := info["origin"]
				version := info["version"]

				if origin == "maven" {
					url := fmt.Sprintf("pkg:maven/%s@%s", dep, version)
					if err := handleMavenURL(url, libDir); err != nil {
						fmt.Println("Failed to install from Maven:", err)
					}
				} else if origin == "github" {
					url := fmt.Sprintf("https://github.com/%s", dep)
					if err := handleGitHubURL(url, libDir); err != nil {
						fmt.Println("Failed to install from GitHub:", err)
					}
				}
			}
			return
		}
		url := args[1]
		if strings.HasPrefix(url, "pkg:maven") {
			if err := handleMavenURL(url, libDir); err != nil {
				fmt.Println("Failed to install from Maven:", err)
			}
		} else if strings.HasPrefix(url, "https://github.com") {
			if err := handleGitHubURL(url, libDir); err != nil {
				fmt.Println("Failed to install from GitHub:", err)
			}
		} else {
			fmt.Println("Unsupported URL format. Use Maven Central or GitHub.")
		}
	default:
		fmt.Println("Invalid command. Use 'build', 'run', or 'install'.")
	}
}
