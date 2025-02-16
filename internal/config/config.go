package config

import (
	"avito-merch/pkg/database"
	"os"
)

type Config struct {
	DBConfig   database.Config
	ServerPort string
}

func LoadConfig() *Config {
	return &Config{
		DBConfig: database.Config{
			DBHost:     getEnv("DATABASE_HOST", "localhost"),
			DBPort:     getEnv("DATABASE_PORT", "5432"),
			DBUser:     getEnv("DATABASE_USER", "postgres"),
			DBPassword: getEnv("DATABASE_PASSWORD", "password"),
			DBName:     getEnv("DATABASE_NAME", "shop"),
		},
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
