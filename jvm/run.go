package jvm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func RunJava(mainClass, binDir, libDir string) error {
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
