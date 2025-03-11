package main

import (
	"errors"
	"flag"
	"fmt"
	"jpkg/pkg/config"
)

func main() {
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

	_, error := config.GetTomlConfig()
	if error != nil {
		err := errors.New("initialize the project. then try running [jpkg run|jpkg build]")
		fmt.Println(err)
		return
	}

	switch args[0] {
	case "build":
		buildJar()
	case "build-native":
		buildNative()
	case "run":
		runApp()
	case "install":
		installPackage()
	default:
		fmt.Println("Invalid command. Use 'build', 'run', or 'install'.")
	}
}
