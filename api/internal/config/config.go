package config

import (
	"os"
	"strconv"
)

type Config struct {
	Host                string
	Port                string
	DBPath              string
	RateLimitPerMinute  int
	RateLimitPerHour    int
	MockTelecomDelayMs  int
	MockAntiFraudEnabled bool
	DefaultLanguage     string
}

func Load() *Config {
	return &Config{
		Host:                getEnv("API_HOST", "0.0.0.0"),
		Port:                getEnv("API_PORT", "8080"),
		DBPath:              getEnv("API_DB_PATH", "./data/promo.db"),
		RateLimitPerMinute:  getEnvInt("RATE_LIMIT_PER_MINUTE", 5),
		RateLimitPerHour:    getEnvInt("RATE_LIMIT_PER_HOUR", 20),
		MockTelecomDelayMs:  getEnvInt("MOCK_TELECOM_DELAY_MS", 500),
		MockAntiFraudEnabled: getEnvBool("MOCK_ANTIFRAUD_ENABLED", true),
		DefaultLanguage:     getEnv("DEFAULT_LANGUAGE", "fr"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
