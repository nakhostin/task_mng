package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Host string
	Port string
}

func LoadConfigFromEnv() (Config, error) {
	if err := godotenv.Load(); err != nil {
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	config := Config{
		Host: host,
		Port: port,
	}

	return config, nil
}

func MustLoadConfigFromEnv() Config {
	config, err := LoadConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return config
}

func ValidateConfig(config Config) error {
	if config.Host == "" {
		return fmt.Errorf("host is required")
	}

	if config.Port == "" {
		return fmt.Errorf("port is required")
	}

	return nil
}
