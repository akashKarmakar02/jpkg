package main

import (
	"flag"
	"fmt"
	"jpkg/jvm"
	"jpkg/pkg/config"
)

func buildJar() {
	appConfig := config.GetConfig()
	tomlConfig, err := config.GetTomlConfig()

	if err != nil {
		fmt.Println("Error while getting mainClass from toml")
	}

	if err := jvm.CompileJava(appConfig.SrcDir, appConfig.BinDir, appConfig.PackageDir); err != nil {
		fmt.Println("Failed to compile:", err)
		return
	}
	if path, err := jvm.CreateJar(appConfig.BinDir, "app.jar", tomlConfig.MainClass, "lib"); err != nil {
		fmt.Println("Failed to create JAR:", err)
	} else {
		fmt.Println("\nBuild Successfully.")
		fmt.Println("Saved file: ", path)
	}
}

func buildNative() {
	appConfig := config.GetConfig()
	tomlConfig, err := config.GetTomlConfig()
	args := flag.Args()

	if err != nil {
		fmt.Println("Error while getting mainClass from toml")
	}

	if err := jvm.CompileJava(appConfig.SrcDir, appConfig.BinDir, appConfig.PackageDir); err != nil {
		fmt.Println("Failed to compile:", err)
		return
	}
	if path, err := jvm.CreateJar(appConfig.BinDir, "app.jar", tomlConfig.MainClass, "lib"); err != nil {
		fmt.Println("Failed to create JAR:", err)
	} else {
		err := jvm.BuildNative(path, args[1:])
		if err != nil {
			fmt.Println("Failed to compile native exec: ", err)
		}
	}
}
