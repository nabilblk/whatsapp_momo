package config

import (
	"os"
)

type Config struct {
	APIUrl          string
	SessionPath     string
	QRTerminal      bool
	DefaultLanguage string
	LogLevel        string
}

func Load() *Config {
	return &Config{
		APIUrl:          getEnv("BRIDGE_API_URL", "http://localhost:8080"),
		SessionPath:     getEnv("BRIDGE_SESSION_PATH", "./data/whatsapp.db"),
		QRTerminal:      getEnvBool("BRIDGE_QR_TERMINAL", true),
		DefaultLanguage: getEnv("DEFAULT_LANGUAGE", "fr"),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
