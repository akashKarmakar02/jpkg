package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	MainClass string `toml:"main_class"`
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
	config := Config{MainClass: "Main"}
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

	fmt.Println("Project initialized successfully.")
	return nil
}
