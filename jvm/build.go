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

	// Path to the fat JAR file in the .jpkg/build directory
	fatJarFilePath := filepath.Join(buildDir, jarFileName)

	// Create a temporary manifest file
	manifestFile := filepath.Join(buildDir, "MANIFEST.MF")
	manifestContent := fmt.Sprintf("Main-Class: %s\n", mainClass)
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		return err
	}
	defer os.Remove(manifestFile)

	// Collect all JAR files from the libDir
	jarFiles, err := getJarFiles(libDir)
	if err != nil {
		return err
	}
	fmt.Println(jarFiles)

	// Create the fat JAR
	cmd := exec.Command("jar", "cfm", fatJarFilePath, manifestFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Add classes from binDir
	cmd.Args = append(cmd.Args, "-C", binDir, ".")

	// Add classes from all JAR files in libDir
	for _, jarFile := range strings.Split(jarFiles, string(os.PathListSeparator)) {
		cmd.Args = append(cmd.Args, "-C", jarFile)
	}

	// Execute the jar command to create the fat JAR
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create fat JAR: %w", err)
	}

	fmt.Println("Fat JAR file created successfully:", fatJarFilePath)
	return nil
}
