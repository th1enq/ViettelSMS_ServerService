package config

import (
	"fmt"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

type (
	Server struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	Postgres struct {
		Host         string `mapstructure:"host"`
		Port         int    `mapstructure:"port"`
		User         string `mapstructure:"user"`
		Password     string `mapstructure:"password"`
		DBName       string `mapstructure:"dbname"`
		MaxIdleConns int    `mapstructure:"max_idle_conns"`
		MaxOpenConns int    `mapstructure:"max_open_conns"`
	}

	Logger struct {
		Level      string `mapstructure:"level"`
		FilePath   string `mapstructure:"file_path"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxAge     int    `mapstructure:"max_age"`
	}

	Broker struct {
		Address  []string `mapstructure:"address"`
		ClientID string   `mapstructure:"client_id"`
	}
)

type Config struct {
	Server   Server   `mapstructure:"server"`
	Postgres Postgres `mapstructure:"postgres"`
	Logger   Logger   `mapstructure:"logger"`
	Broker   Broker   `mapstructure:"broker"`
}

var ConfigWireSet = wire.NewSet(LoadConfig)

func LoadConfig() *Config {
	viper := viper.New()
	viper.SetConfigFile("config.dev.yaml")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("failed to read configuration %w \n", err))
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode configuration %v", err))
	}

	return &cfg
}
