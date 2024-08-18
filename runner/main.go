package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

type DependencyLock struct {
	Dependencies map[string]string `json:"dependencies"`
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Usage: pass a java executable name [amber-cli|server createproject]")
		os.Exit(1)
	}

	jarName, err := getJarFileName(args[0])
	if err != nil {
		fmt.Println("Failed to get JAR file name:", err)
		os.Exit(1)
	}

	// Prepare the arguments for exec.Command
	javaArgs := []string{"-jar", fmt.Sprintf("lib/%s", jarName)}
	if len(args) > 1 {
		javaArgs = append(javaArgs, args[1:]...)
	}

	// Use variadic argument list for exec.Command
	cmd := exec.Command("java", javaArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// GetJarFileName retrieves the JAR file name for a given dependency
func getJarFileName(name string) (string, error) {
	lockFile := "dependencies-lock.json"
	var lock DependencyLock

	file, err := os.Open(lockFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&lock); err != nil {
		return "", err
	}

	jarFileName, exists := lock.Dependencies[name]
	if !exists {
		return "", fmt.Errorf("dependency %s not found in lock file", name)
	}
	return jarFileName, nil
}
