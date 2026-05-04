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
	AI       AIConfig
	Report   ReportConfig
	Outbound OutboundCallConfig
}

type AIConfig struct {
	Provider string
	APIKey   string
	Model    string
	BaseURL  string
}

type ReportConfig struct {
	ScheduleRetentionDays int
}

type AuthConfig struct {
	JWTSecret       string
	JWTExpiresIn    int
	StudentAccount  string
	StudentPassword string
	DoctorAccount   string
	DoctorPassword  string
	AdminAccount    string
	AdminPassword   string
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type OutboundCallConfig struct {
	Provider               string
	AliyunAccessKeyID      string
	AliyunAccessKeySecret  string
	AliyunRegionID         string
	AliyunCalledShowNumber string
	AliyunTTSCode          string
	AliyunPlayTimes        int
	AliyunTemplateCode     string
	AliyunCallbackSecret   string
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
			JWTSecret:       getEnv("AUTH_JWT_SECRET", ""),
			JWTExpiresIn:    getEnvAsInt("AUTH_JWT_EXPIRES_IN", 7200),
			StudentAccount:  getEnv("AUTH_STUDENT_ACCOUNT", "student"),
			StudentPassword: getEnv("AUTH_STUDENT_PASSWORD", ""),
			DoctorAccount:   getEnv("AUTH_DOCTOR_ACCOUNT", "doctor"),
			DoctorPassword:  getEnv("AUTH_DOCTOR_PASSWORD", ""),
			AdminAccount:    getEnv("AUTH_ADMIN_ACCOUNT", "admin"),
			AdminPassword:   getEnv("AUTH_ADMIN_PASSWORD", ""),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		AI: AIConfig{
			Provider: getEnv("AI_PROVIDER", "rule"),
			APIKey:   getEnv("AI_API_KEY", ""),
			Model:    getEnv("AI_MODEL", "qwen3.6-plus"),
			BaseURL:  getEnv("AI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
		},
		Report: ReportConfig{
			ScheduleRetentionDays: getEnvAsInt("REPORT_SCHEDULE_RETENTION_DAYS", 30),
		},
		Outbound: OutboundCallConfig{
			Provider:               strings.ToLower(getEnv("OUTBOUND_CALL_PROVIDER", "mock")),
			AliyunAccessKeyID:      getEnv("ALIYUN_CALL_ACCESS_KEY_ID", ""),
			AliyunAccessKeySecret:  getEnv("ALIYUN_CALL_ACCESS_KEY_SECRET", ""),
			AliyunRegionID:         getEnv("ALIYUN_CALL_REGION_ID", "cn-hangzhou"),
			AliyunCalledShowNumber: getEnv("ALIYUN_CALL_CALLED_SHOW_NUMBER", ""),
			AliyunTTSCode:          getEnv("ALIYUN_CALL_TTS_CODE", ""),
			AliyunPlayTimes:        getEnvAsInt("ALIYUN_CALL_PLAY_TIMES", 2),
			AliyunTemplateCode:     getEnv("ALIYUN_CALL_TEMPLATE_CODE", "external_medical_followup"),
			AliyunCallbackSecret:   getEnv("ALIYUN_CALL_CALLBACK_SECRET", ""),
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

	studentAccount := strings.TrimSpace(c.StudentAccount)
	doctorAccount := strings.TrimSpace(c.DoctorAccount)
	adminAccount := strings.TrimSpace(c.AdminAccount)
	if studentAccount == "" || doctorAccount == "" || adminAccount == "" {
		return errors.New("AUTH_STUDENT_ACCOUNT、AUTH_DOCTOR_ACCOUNT 和 AUTH_ADMIN_ACCOUNT 不能为空")
	}
	if hasDuplicateValues(studentAccount, doctorAccount, adminAccount) {
		return errors.New("AUTH_STUDENT_ACCOUNT、AUTH_DOCTOR_ACCOUNT 和 AUTH_ADMIN_ACCOUNT 不能相同")
	}

	if isUnsafePassword(c.StudentPassword, "replace-with-student-password") {
		return errors.New("AUTH_STUDENT_PASSWORD 必须设置为非默认值")
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

func hasDuplicateValues(values ...string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
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
