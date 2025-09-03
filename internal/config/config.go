package config

import (
	"github.com/spf13/viper"
)

type (
	Server struct {
		Host string
		Port int
	}

	Postgres struct {
		Host         string
		Port         int
		User         string
		Password     string
		DBName       string
		MaxIdleConns int
		MaxOpenConns int
	}

	Logger struct {
		Level      string
		FilePath   string
		MaxSize    int
		MaxBackups int
		MaxAge     int
	}

	Kafka struct {
		Address  []string
		ClientID string
	}

	JWT struct {
		Secret string
	}

	Consumer struct {
		StatusConsumer string
	}
)

type Config struct {
	Server   Server
	Postgres Postgres
	Logger   Logger
	Kafka    Kafka
	JWT      JWT
	Consumer Consumer
}

func LoadConfig() *Config {
	viper := viper.New()
	viper.AutomaticEnv()

	// server service env
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("SERVER_PORT", 8080)

	serverEnv := Server{
		Host: viper.GetString("SERVER_HOST"),
		Port: viper.GetInt("SERVER_PORT"),
	}

	// postgres env
	viper.SetDefault("POSTGRES_HOST", "postgres")
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "password")
	viper.SetDefault("POSTGRES_DBNAME", "vcs_sms")
	viper.SetDefault("POSTGRES_MAX_CONNECTIONS", 100)
	viper.SetDefault("POSTGRES_IDLE_CONNECTIONS", 10)

	postgresEnv := Postgres{
		Host:         viper.GetString("POSTGRES_HOST"),
		Port:         viper.GetInt("POSTGRES_PORT"),
		User:         viper.GetString("POSTGRES_USER"),
		Password:     viper.GetString("POSTGRES_PASSWORD"),
		DBName:       viper.GetString("POSTGRES_DBNAME"),
		MaxIdleConns: viper.GetInt("POSTGRES_MAX_CONNECTIONS"),
		MaxOpenConns: viper.GetInt("POSTGRES_IDLE_CONNECTIONS"),
	}

	// logger env
	viper.SetDefault("LOGGER_LEVEL", "info")
	viper.SetDefault("LOGGER_FILE_PATH", "./logs/app.log")
	viper.SetDefault("LOGGER_MAX_SIZE", 100)
	viper.SetDefault("LOGGER_MAX_BACKUPS", 10)
	viper.SetDefault("LOGGER_MAX_AGE", 30)

	loggerEnv := Logger{
		Level:      viper.GetString("LOGGER_LEVEL"),
		FilePath:   viper.GetString("LOGGER_FILE_PATH"),
		MaxSize:    viper.GetInt("LOGGER_MAX_SIZE"),
		MaxBackups: viper.GetInt("LOGGER_MAX_BACKUPS"),
		MaxAge:     viper.GetInt("LOGGER_MAX_AGE"),
	}

	// kafka env
	viper.SetDefault("KAFKA_ADDRESS", []string{"kafka:9092"})
	viper.SetDefault("KAFKA_CLIENT_ID", "vcs_sms")
	kafkaEnv := Kafka{
		Address:  viper.GetStringSlice("KAFKA_ADDRESS"),
		ClientID: viper.GetString("KAFKA_CLIENT_ID"),
	}

	viper.SetDefault("JWT_SECRET", "mysecret")
	jwtEnv := JWT{
		Secret: viper.GetString("JWT_SECRET"),
	}

	// consumer env
	viper.SetDefault("STATUS_CONSUMER_GROUP", "status-consumer-group")
	consumerEnv := Consumer{
		StatusConsumer: viper.GetString("STATUS_CONSUMER_GROUP"),
	}

	return &Config{
		Server:   serverEnv,
		Postgres: postgresEnv,
		Logger:   loggerEnv,
		Kafka:    kafkaEnv,
		JWT:      jwtEnv,
		Consumer: consumerEnv,
	}
}
