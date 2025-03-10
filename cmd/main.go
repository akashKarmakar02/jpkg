package main

import (
	"errors"
	"flag"
	"fmt"
	"jpkg/cache"
	"jpkg/config"
	"jpkg/downloader"
	"jpkg/jvm"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func watchForChanges(srcDir, binDir, libDir, cacheDir, mainClass string, javaCmd *exec.Cmd) {
	for {
		isUptoDate, err := cache.IsCacheUpToDate(srcDir, cacheDir)
		if err == nil && !isUptoDate {
			fmt.Println("\033[2;37mFile changes found. Reloading the app...\033[0m")

			// Stop the running process
			if javaCmd != nil && javaCmd.Process != nil {
				// Check if the process is still running
				err := javaCmd.Process.Signal(os.Interrupt)
				if err != nil {
					if err.Error() != "os: process already finished" {
						fmt.Println("Failed to kill running process:", err)
					}
				} else {
					// Attempt to kill the process only if it hasn't finished
					if err := javaCmd.Process.Kill(); err != nil {
						fmt.Println("Failed to kill running process:", err)
					}
				}
			}

			// Recompile and rerun the Java program
			cache.CopySrcToCache(srcDir, cacheDir)
			if err := jvm.CompileJava(srcDir, binDir, libDir); err != nil {
				fmt.Println("\033[2;37mFailed to compile:", err, "\033[0m")
				return
			}

			javaCmd = jvm.RunJava(mainClass, binDir, libDir)

			go javaCmd.Run()
			fmt.Print("\033[H\033[2J")
		}

		// Sleep for a while before checking again
		time.Sleep(time.Second)
	}
}

func main() {
	srcDir := "src"
	binDir := ".jpkg/bin"
	libDir := "lib"
	cacheDir := ".jpkg/cache"
	var javaCmd *exec.Cmd

	// Parse command-line arguments
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: jpkg [build|run|init]")
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
		if err := jvm.CompileJava(srcDir, binDir, libDir); err != nil {
			fmt.Println("Failed to compile:", err)
			return
		}
		if path, err := jvm.CreateJar(binDir, "app.jar", mainClass, "lib"); err != nil {
			fmt.Println("Failed to create JAR:", err)
		} else {
			fmt.Println("\nBuild Successfully.")
			fmt.Println("Saved file: ", path)
		}

	case "build-native":
		if err := jvm.CompileJava(srcDir, binDir, libDir); err != nil {
			fmt.Println("Failed to compile:", err)
			return
		}
		if path, err := jvm.CreateJar(binDir, "app.jar", mainClass, "lib"); err != nil {
			fmt.Println("Failed to create JAR:", err)
		} else {
			err := jvm.BuildNative(path, args[1:])
			if err != nil {
				fmt.Println("Failed to compile native exec: ", err)
			}
		}

	case "run":
		isUptoDate, err := cache.IsCacheUpToDate(srcDir, cacheDir)
		if err != nil {
			fmt.Println("Failed to cache files: ", err)
		}

		if len(args) > 1 && strings.HasSuffix(args[1], ".java") {
			mainClass = args[1]
		}

		if isUptoDate {
			javaCmd = jvm.RunJava(mainClass, binDir, libDir)

			if len(args) > 1 && slices.Contains(args, "--watch") {
				go javaCmd.Run()
				watchForChanges(srcDir, binDir, libDir, cacheDir, mainClass, javaCmd)
				return
			}
			javaCmd.Run()
			return
		}

		cache.CopySrcToCache(srcDir, cacheDir)
		if err := jvm.CompileJava(srcDir, binDir, libDir); err != nil {
			fmt.Println("Failed to compile:", err)
			return
		}

		javaCmd = jvm.RunJava(mainClass, binDir, libDir)

		if err != nil && !os.IsNotExist(err) {
			fmt.Println("Failed to run:", err)
		}

		if len(args) > 1 && args[1] == "--watch" {
			go javaCmd.Run()
			watchForChanges(srcDir, binDir, libDir, cacheDir, mainClass, javaCmd)
			return
		}
		javaCmd.Run()

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
					if err := downloader.HandleMavenURL(url, libDir); err != nil {
						fmt.Println("Failed to install from Maven:", err)
					}
				} else if origin == "github" {
					url := fmt.Sprintf("https://github.com/%s", dep)
					if err := downloader.HandleGitHubURL(url, libDir); err != nil {
						fmt.Println("Failed to install from GitHub:", err)
					}
				}
			}
			return
		}
		url := args[1]
		if strings.HasPrefix(url, "pkg:maven") {
			if err := downloader.HandleMavenURL(url, libDir); err != nil {
				fmt.Println("Failed to install from Maven:", err)
			}
		} else if strings.HasPrefix(url, "https://github.com") {
			if err := downloader.HandleGitHubURL(url, libDir); err != nil {
				fmt.Println("Failed to install from GitHub:", err)
			}
		} else {
			fmt.Println("Unsupported URL format. Use Maven Central or GitHub.")
		}
	default:
		fmt.Println("Invalid command. Use 'build', 'run', or 'install'.")
	}
}
