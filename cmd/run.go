package main

import (
	"flag"
	"fmt"
	"jpkg/jvm"
	"jpkg/pkg/cache"
	"jpkg/pkg/config"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
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

func runApp() {
	args := flag.Args()
	appConfig := config.GetConfig()
	tomlConfig, err := config.GetTomlConfig()
	var javaCmd *exec.Cmd

	mainClass := tomlConfig.MainClass

	isUptoDate, err := cache.IsCacheUpToDate(appConfig.SrcDir, appConfig.CacheDir)

	if err != nil {
		fmt.Println("Failed to cache files: ", err)
	}

	if len(args) > 1 && strings.HasSuffix(args[1], ".java") {
		mainClass = args[1]
	}

	if isUptoDate {
		javaCmd = jvm.RunJava(mainClass, appConfig.BinDir, appConfig.BinDir)

		if len(args) > 1 && slices.Contains(args, "--watch") {
			go javaCmd.Run()
			watchForChanges(appConfig.SrcDir, appConfig.BinDir, appConfig.PackageDir, appConfig.CacheDir, mainClass, javaCmd)
			return
		}
		javaCmd.Run()
		return
	}

	cache.CopySrcToCache(appConfig.SrcDir, appConfig.CacheDir)
	if err := jvm.CompileJava(appConfig.SrcDir, appConfig.BinDir, appConfig.PackageDir); err != nil {
		fmt.Println("Failed to compile:", err)
		return
	}

	javaCmd = jvm.RunJava(mainClass, appConfig.BinDir, appConfig.PackageDir)

	if err != nil && !os.IsNotExist(err) {
		fmt.Println("Failed to run:", err)
	}

	if len(args) > 1 && args[1] == "--watch" {
		go javaCmd.Run()
		watchForChanges(appConfig.SrcDir, appConfig.BinDir, appConfig.PackageDir, appConfig.CacheDir, mainClass, javaCmd)
		return
	}
	javaCmd.Run()
}
