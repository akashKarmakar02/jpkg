package main

import (
	"flag"
	"fmt"
	"jpkg/downloader"
	"jpkg/pkg/config"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

func installPackage() {
	args := flag.Args()

	appConfig := config.GetConfig()

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
				if err := downloader.HandleMavenURL(url, appConfig.PackageDir); err != nil {
					fmt.Println("Failed to install from Maven:", err)
				}
			} else if origin == "github" {
				url := fmt.Sprintf("https://github.com/%s", dep)
				if err := downloader.HandleGitHubURL(url, appConfig.PackageDir); err != nil {
					fmt.Println("Failed to install from GitHub:", err)
				}
			}
		}
		return
	}
	url := args[1]
	if strings.HasPrefix(url, "pkg:maven") {
		if err := downloader.HandleMavenURL(url, appConfig.PackageDir); err != nil {
			fmt.Println("Failed to install from Maven:", err)
		}
	} else if strings.HasPrefix(url, "https://github.com") {
		if err := downloader.HandleGitHubURL(url, appConfig.PackageDir); err != nil {
			fmt.Println("Failed to install from GitHub:", err)
		}
	} else {
		fmt.Println("Unsupported URL format. Use Maven Central or GitHub.")
	}
}
