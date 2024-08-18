package jvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func CompileJava(srcDir, binDir, libDir string) error {
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

func CreateJar(binDir, jarFileName, mainClass, libDir string) error {
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

	for index, file := range relJarFilesList {
		absJarPath, _ := filepath.Abs(file)
		relJarFilesList[index] = absJarPath
	}

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
