package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName  string
	AppEnv   string
	AppPort  int
	DataMode string
	Auth     AuthConfig
	DB       DBConfig
}

type AuthConfig struct {
	JWTSecret      string
	JWTExpiresIn   int
	DoctorAccount  string
	DoctorPassword string
	AdminAccount   string
	AdminPassword  string
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
		Auth: AuthConfig{
			JWTSecret:      getEnv("AUTH_JWT_SECRET", ""),
			JWTExpiresIn:   getEnvAsInt("AUTH_JWT_EXPIRES_IN", 7200),
			DoctorAccount:  getEnv("AUTH_DOCTOR_ACCOUNT", "doctor"),
			DoctorPassword: getEnv("AUTH_DOCTOR_PASSWORD", ""),
			AdminAccount:   getEnv("AUTH_ADMIN_ACCOUNT", "admin"),
			AdminPassword:  getEnv("AUTH_ADMIN_PASSWORD", ""),
		},
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

func (c Config) Validate() error {
	return c.Auth.Validate()
}

func (c AuthConfig) Validate() error {
	jwtSecret := strings.TrimSpace(c.JWTSecret)
	if jwtSecret == "" {
		return errors.New("AUTH_JWT_SECRET 不能为空")
	}
	if jwtSecret == "replace-with-a-long-random-secret" {
		return errors.New("AUTH_JWT_SECRET 必须替换为自定义值")
	}

	doctorAccount := strings.TrimSpace(c.DoctorAccount)
	adminAccount := strings.TrimSpace(c.AdminAccount)
	if doctorAccount == "" || adminAccount == "" {
		return errors.New("AUTH_DOCTOR_ACCOUNT 和 AUTH_ADMIN_ACCOUNT 不能为空")
	}
	if doctorAccount == adminAccount {
		return errors.New("AUTH_DOCTOR_ACCOUNT 和 AUTH_ADMIN_ACCOUNT 不能相同")
	}

	if isUnsafePassword(c.DoctorPassword, "replace-with-doctor-password") {
		return errors.New("AUTH_DOCTOR_PASSWORD 必须设置为非默认值")
	}
	if isUnsafePassword(c.AdminPassword, "replace-with-admin-password") {
		return errors.New("AUTH_ADMIN_PASSWORD 必须设置为非默认值")
	}

	return nil
}

func isUnsafePassword(password string, placeholder string) bool {
	normalized := strings.TrimSpace(password)
	if normalized == "" {
		return true
	}

	switch normalized {
	case placeholder, "dev", "admin123", "changeme", "change-me", "password":
		return true
	default:
		return false
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
