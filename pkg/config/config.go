package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort      string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBName          string
	MigrationsDir   string
	JWTSecret       string
	JWTExpiresHours int
}

func Load() Config {
	return Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "127.0.0.1"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBUser:          getEnv("DB_USER", "shortlink"),
		DBPassword:      getEnv("DB_PASSWORD", "shortlink_password"),
		DBName:          getEnv("DB_NAME", "shortlink"),
		MigrationsDir:   getEnv("MIGRATIONS_DIR", "/app/migrations"),
		JWTSecret:       getEnv("JWT_SECRET", "change-this-secret-in-production"),
		JWTExpiresHours: getEnvInt("JWT_EXPIRES_HOURS", 24),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
