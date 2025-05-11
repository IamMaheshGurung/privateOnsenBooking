package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	configInstance *Config
	once           sync.Once
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	ServerPort  string
	ServerHost  string
	AppName     string
	AppURL      string
	Environment string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT Authentication
	JWTSecret        string
	JWTExpiryHours   int
	JWTRefreshSecret string
	JWTIssuer        string

	// Session configuration
	SessionSecret string

	// Email configuration
	SMTPHost      string
	SMTPPort      int
	SMTPUser      string
	SMTPPass      string
	SMTPFrom      string
	SMTPFromName  string
	SMTPTemplates string

	// Storage configuration
	StoragePath    string
	MaxUploadSize  int64
	AllowedFormats []string

	// Pagination defaults
	DefaultPageSize int
	MaxPageSize     int

	// Rate limiting
	RateLimit     int
	RateLimitTime time.Duration
}

// GetConfig returns the singleton config instance
func GetConfig() *Config {
	once.Do(func() {
		configInstance = &Config{
			// Server configuration
			ServerPort:  getEnv("PORT", "3000"),
			ServerHost:  getEnv("HOST", "localhost"),
			AppName:     getEnv("APP_NAME", "Kwangdi Onsen"),
			AppURL:      getEnv("APP_URL", "http://localhost:3000"),
			Environment: getEnv("ENVIRONMENT", "development"),

			// Database configuration
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBUser:     getEnv("DB_USER", "postgres"),
			DBPassword: getEnv("DB_PASSWORD", "postgres"),
			DBName:     getEnv("DB_NAME", "kwangdi_onsen"),
			DBSSLMode:  getEnv("DB_SSL_MODE", "disable"),

			// JWT Authentication
			JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			JWTExpiryHours:   getIntEnv("JWT_EXPIRY_HOURS", 24),
			JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-in-production"),
			JWTIssuer:        getEnv("JWT_ISSUER", "kwangdi-onsen"),

			// Session configuration
			SessionSecret: getEnv("SESSION_SECRET", "your-session-secret-change-in-production"),

			// Email configuration
			SMTPHost:      getEnv("SMTP_HOST", "smtp.example.com"),
			SMTPPort:      getIntEnv("SMTP_PORT", 587),
			SMTPUser:      getEnv("SMTP_USER", ""),
			SMTPPass:      getEnv("SMTP_PASS", ""),
			SMTPFrom:      getEnv("SMTP_FROM", "noreply@kwangdionsen.com"),
			SMTPFromName:  getEnv("SMTP_FROM_NAME", "Kwangdi Onsen"),
			SMTPTemplates: getEnv("SMTP_TEMPLATES", "./templates/emails"),

			// Storage configuration
			StoragePath:    getEnv("STORAGE_PATH", "./static/uploads"),
			MaxUploadSize:  getInt64Env("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB default
			AllowedFormats: getSliceEnv("ALLOWED_FORMATS", []string{"jpg", "jpeg", "png", "gif"}),

			// Pagination defaults
			DefaultPageSize: getIntEnv("DEFAULT_PAGE_SIZE", 20),
			MaxPageSize:     getIntEnv("MAX_PAGE_SIZE", 100),

			// Rate limiting
			RateLimit:     getIntEnv("RATE_LIMIT", 100),
			RateLimitTime: getDurationEnv("RATE_LIMIT_TIME", 1*time.Minute),
		}
	})

	return configInstance
}

// GetEmailConfig returns EmailConfig for email service
func (c *Config) GetEmailConfig() map[string]interface{} {
	return map[string]interface{}{
		"SMTPServer":   c.SMTPHost,
		"SMTPPort":     c.SMTPPort,
		"SMTPUsername": c.SMTPUser,
		"SMTPPassword": c.SMTPPass,
		"FromEmail":    c.SMTPFrom,
		"FromName":     c.SMTPFromName,
		"TemplatesDir": c.SMTPTemplates,
		"Environment":  c.Environment,
	}
}

// Helper functions to get environment variables with defaults

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getInt64Env(key string, fallback int64) int64 {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intVal
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func getSliceEnv(key string, fallback []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return stringSplit(value, ",")
	}
	return fallback
}

func stringSplit(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
