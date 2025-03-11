package jvm

import (
	"errors"
	"fmt"
	"jpkg/pkg/cache"
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

func copyDir(src string, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(path, src)
		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(destPath, data, info.Mode())
		}
	})
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
	cache.RemoveAll(binDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Copy resources to binDir
	resourcesDir := "resources"
	if _, err := os.Stat(resourcesDir); !os.IsNotExist(err) {
		err = copyDir(resourcesDir, binDir)
		if err != nil {
			return fmt.Errorf("failed to copy resources: %w", err)
		}
	}

	return nil
}

func CreateJar(binDir, jarFileName, mainClass, libDir string) (string, error) {
	buildDir := filepath.Join(".jpkg", "build", "jar")
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		if err := os.MkdirAll(buildDir, os.ModePerm); err != nil {
			return "", err
		}
	}

	jarFiles, err := getJarFiles(libDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	relJarFiles := strings.ReplaceAll(jarFiles, string(os.PathSeparator), "/")
	relJarFilesList := strings.Split(relJarFiles, string(os.PathListSeparator))

	for index, file := range relJarFilesList {
		absJarPath, _ := filepath.Abs(file)
		relJarFilesList[index] = absJarPath
	}

	manifestFile := filepath.Join(binDir, "MANIFEST.MF")
	manifestContent := fmt.Sprintf("Main-Class: %s\nClass-Path: %s\n", mainClass, strings.Join(relJarFilesList, " "))
	if err := os.WriteFile(manifestFile, []byte(manifestContent), 0644); err != nil {
		return "", err
	}
	defer os.Remove(manifestFile)

	jarFilePath := filepath.Join(buildDir, jarFileName)

	cmd := exec.Command("jar", "cmf", manifestFile, jarFilePath, "-C", binDir, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return jarFilePath, nil
}

func BuildNative(jarPath string, args []string) error {
	buildDir := filepath.Join(".jpkg", "build", "linux")
	if _, err := os.Stat(buildDir); os.IsNotExist(err) {
		if err := os.MkdirAll(buildDir, os.ModePerm); err != nil {
			return err
		}
	}
	command := []string{"native-image", "--no-fallback"}
	command = append(command, args...)

	cmd := exec.Command(command[0], append(command[1:], "-jar", jarPath, ".jpkg/build/linux/app")...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func DetectRequiredModules(jarFilePath string) (string, error) {
	cmd := exec.Command("jdeps",
		"--multi-release", "9",
		"--module-path", os.Getenv("JAVA_HOME")+"/jmods",
		"--print-module-deps",
		jarFilePath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to detect required modules: %w", err)
	}

	modules := strings.TrimSpace(string(output))
	return modules, nil
}

func CreateCustomRuntime(outputDir, jarFilePath, modules string) error {
	if _, err := os.Stat(outputDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to remove existing runtime directory: %w", err)
		}
	}

	// Construct the jlink command
	cmd := exec.Command("jlink",
		"--module-path", os.Getenv("JAVA_HOME")+"/jmods:"+filepath.Dir(jarFilePath),
		"--add-modules", modules,
		"--limit-modules", modules,
		"--output", outputDir,
		"--strip-debug",
		"--verbose",
		"--compress=2",
		"--no-header-files",
		"--no-man-pages",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create custom runtime: %w", err)
	}

	return nil
}
