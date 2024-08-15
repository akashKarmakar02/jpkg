package config

import (
	"fmt"
	"os"

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
