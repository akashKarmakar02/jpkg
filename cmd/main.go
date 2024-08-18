package main

import (
	"errors"
	"flag"
	"fmt"
	"jpkg/cache"
	"jpkg/config"
	downloader "jpkg/download"
	"jpkg/jvm"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

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
		if err := jvm.CompileJava(srcDir, binDir, libDir); err != nil {
			fmt.Println("Failed to compile:", err)
			return
		}
		if err := jvm.CreateJar(binDir, "app.jar", mainClass, "lib"); err != nil {
			fmt.Println("Failed to create JAR:", err)
		}
	case "run":
		if isUptoDate, err := cache.IsCacheUpToDate(srcDir, cacheDir); err == nil && isUptoDate {
			if err := jvm.RunJava(mainClass, binDir, libDir); err != nil {
				fmt.Println("Failed to run:", err)
			}
		} else {
			cache.CopySrcToCache(srcDir, cacheDir)
			if err := jvm.CompileJava(srcDir, binDir, libDir); err != nil {
				fmt.Println("Failed to compile:", err)
				return
			}
			if err := jvm.RunJava(mainClass, binDir, libDir); err != nil {
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
