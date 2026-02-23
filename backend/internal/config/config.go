package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName  string
	AppEnv   string
	AppPort  int
	DataMode string
	DB       DBConfig
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (db DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host,
		db.Port,
		db.User,
		db.Password,
		db.Name,
		db.SSLMode,
	)
}

func (db DBConfig) IsConfigured() bool {
	return db.Host != "" && db.User != "" && db.Name != ""
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppName:  getEnv("APP_NAME", "medical-assistant-backend"),
		AppEnv:   getEnv("APP_ENV", "development"),
		AppPort:  getEnvAsInt("APP_PORT", 8080),
		DataMode: getEnv("APP_DATA_MODE", "db"),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func (c Config) ResolveDataMode(hasDB bool) string {
	if c.DataMode == "mock" || !hasDB {
		return "mock"
	}

	return "db"
}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return fallback
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback
	}

	return value
}
