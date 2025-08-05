package config

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

//go:embed config.dev.yaml
var DefaultConfigBytes []byte

type ConfigFilePath string

type Config struct {
	Server   Server   `yaml:"server"`
	Postgres Postgres `yaml:"postgres"`
	Logger   Logger   `yaml:"logger"`
}

func NewConfig(filePath ConfigFilePath) (Config, error) {
	var (
		configBytes = DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(string(filePath))
		if err != nil {
			return Config{}, fmt.Errorf("failed to read YAML file: %w", err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}
	return config, nil
}

func Load() (Config, error) {
	return NewConfig("configs/config.dev.yaml")
}
