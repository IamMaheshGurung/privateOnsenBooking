package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Email configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	// Application configuration
	AppName     string
	Environment string
	ServerPort  string

	// Database configuration
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// NewConfig creates a new Config struct with values from environment variables
func NewConfig() *Config {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		// Email configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.example.com"),
		SMTPPort:     smtpPort,
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@kwangdihotel.com"),

		// Application configuration
		AppName:     getEnv("APP_NAME", "Kwangdi Hotel Booking"),
		Environment: getEnv("ENVIRONMENT", "development"),
		ServerPort:  getEnv("PORT", "3000"),

		// Database configuration
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "kwangdihotel"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// Helper function to get environment variables with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
