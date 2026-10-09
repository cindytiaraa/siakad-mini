package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTExpires    int // detik
	MigrationsDir string
}

func Load() *Config {
	_ = godotenv.Load()

	expires, err := strconv.Atoi(getEnv("JWT_EXPIRES_IN", "3600"))
	if err != nil || expires <= 0 {
		expires = 3600
	}

	return &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		Port:          getEnv("APP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpires:    expires,
		MigrationsDir: getEnv("MIGRATIONS_DIR", "migrations"),
	}
}

func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

func (c *Config) RequireDatabase() {
	if c.DatabaseURL == "" {
		log.Fatal("DATABASE_URL wajib diisi (lihat .env.example)")
	}
}

func (c *Config) RequireJWT() {
	if c.JWTSecret == "" {
		log.Fatal("JWT_SECRET wajib diisi (lihat .env.example)")
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
