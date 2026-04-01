package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppEnv   string
	HTTPAddr string

	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string

	RedisHost string
	RedisPort string
}

func Load() Config {
	cfg := Config{
		AppEnv:           getEnv("APP_ENV", "development"),
		HTTPAddr:         getEnv("HTTP_ADDR", ":8080"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:       getEnv("POSTGRES_DB", "gamidoc"),
		PostgresUser:     getEnv("POSTGRES_USER", "gamidoc"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "gamidoc"),
		RedisHost:        getEnv("REDIS_HOST", "localhost"),
		RedisPort:        getEnv("REDIS_PORT", "6379"),
	}
	return cfg
}

func (c Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
		c.PostgresUser,
		c.PostgresPassword,
	)
}

func (c Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
