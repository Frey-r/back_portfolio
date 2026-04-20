package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                     string
	AllowedOrigins           []string
	SQLitePath               string
	LinkedInClientID         string
	LinkedInClientSecret     string
	LinkedInRedirectURI      string
	LinkedInAccessToken      string
	ContactNotificationEmail string
	RateLimitContact         int
	RateLimitGeneral         int
}

// Load reads the application configuration from environment variables and .env file.
func Load() *Config {
	// Try loading from .env, ignore if it doesn't exist
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := &Config{
		Port:                 getEnv("PORT", "8080"),
		AllowedOrigins:       strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:4321"), ","),
		SQLitePath:           getEnv("SQLITE_PATH", "./data/portfolio.db"),
		LinkedInClientID:     getEnv("LINKEDIN_CLIENT_ID", getEnv("LINKEDIN_ACCESS_ID", "")),
		LinkedInClientSecret: getEnv("LINKEDIN_CLIENT_SECRET", ""),
		LinkedInRedirectURI:  getEnv("LINKEDIN_REDIRECT_URI", "http://localhost/api/auth/linkedin/callback"),

		LinkedInAccessToken:  getEnv("LINKEDIN_ACCESS_TOKEN", ""),
		ContactNotificationEmail: getEnv("CONTACT_NOTIFICATION_EMAIL", "hello@eduardobachmann.dev"),
		RateLimitContact:     getEnvAsInt("RATE_LIMIT_CONTACT", 10),
		RateLimitGeneral:     getEnvAsInt("RATE_LIMIT_GENERAL", 60),
	}


	// Ensure port has a leading colon if not provided
	if !strings.HasPrefix(cfg.Port, ":") {
		cfg.Port = ":" + cfg.Port
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(name string, fallback int) int {
	valStr := getEnv(name, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}
