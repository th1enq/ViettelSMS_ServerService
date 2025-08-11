package configs

import (
	"fmt"
	"os"

	"github.com/th1enq/ViettelSMS_ServerService/configs"
	"gopkg.in/yaml.v2"
)

type ConfigFilePath string

type Config struct {
	ServerService ServerService `yaml:"server_service"`
	GRPC          GRPC          `yaml:"grpc"`
	Log           Log           `yaml:"log"`
	Postgres      Postgres      `yaml:"postgres"`
}

func NewConfig(filePath ConfigFilePath) (Config, error) {
	var (
		configBytes = configs.DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(string(filePath))
		if err != nil {
			return Config{}, fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}

func Load() (Config, error) {
	return NewConfig("configs/config.yaml")
}
