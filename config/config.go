package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	MainClass    string `toml:"main_class"`
	Dependencies map[string]Dependency
}

type Dependency struct {
	Origin  string
	Version string
}

func GetConfig() (*Config, error) {
	var config Config
	if _, err := os.Stat("amber.toml"); os.IsNotExist(err) {
		return nil, fmt.Errorf("amber.toml file not found")
	}
	if _, err := toml.DecodeFile("amber.toml", &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func CreateInitialFiles() error {
	// Create amber.toml
	config := Config{MainClass: "Main", Dependencies: map[string]Dependency{}}
	f, err := os.Create("amber.toml")
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(config); err != nil {
		return err
	}

	// Create src directory
	if err := os.MkdirAll("src", os.ModePerm); err != nil {
		return err
	}

	// Create src/Main.java
	mainJava := `public class Main {
    public static void main(String[] args) {
        System.out.println("Hello World");
    }
}
`
	mainFilePath := filepath.Join("src", "Main.java")
	if err := os.WriteFile(mainFilePath, []byte(mainJava), 0644); err != nil {
		return err
	}

	fmt.Println("Project initialized successfully.✨")
	return nil
}

func saveConfig(filename string, dependencies string) error {
	// Read the existing file content
	content, err := os.ReadFile("amber.toml")
	if err != nil {
		return err
	}

	// Replace the dependencies section with the new one
	re := regexp.MustCompile(`(?s)\[dependencies\].*?(\n\[|$)`)
	newContent := re.ReplaceAllString(string(content), dependencies+"$1")

	return os.WriteFile(filename, []byte(newContent), 0644)
}

func generateDependenciesSection(dependencies map[string]Dependency) string {
	var sb strings.Builder
	sb.WriteString("[dependencies]\n")
	for name, dep := range dependencies {
		sb.WriteString(fmt.Sprintf(`"%s" = { origin = "%s", version = "%s" }`+"\n", name, dep.Origin, dep.Version))
	}
	return sb.String()
}

func SaveDependency(name, origin, version string) error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	if config.Dependencies == nil {
		config.Dependencies = make(map[string]Dependency)
	}

	config.Dependencies[name] = Dependency{
		Origin:  origin,
		Version: version,
	}

	// Generate the new dependencies section
	dependenciesSection := generateDependenciesSection(config.Dependencies)

	return saveConfig("amber.toml", dependenciesSection)
}
