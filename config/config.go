package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	OAuth    OAuthConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
	SSLMode  string
	TimeZone string
}

type JWTConfig struct {
	Secret           string
	RefreshSecret    string
	AccessExpiry     time.Duration
	RefreshExpiry    time.Duration
}

type OAuthConfig struct {
	Google    OAuthProvider
	GitHub    OAuthProvider
	Microsoft OAuthProvider
	Facebook  OAuthProvider
	RedirectBaseURL string
}

type OAuthProvider struct {
	ClientID     string
	ClientSecret string
	Enabled      bool
}

type ServerConfig struct {
	Port string
}

var AppConfig *Config

// Load loads configuration from environment variables
func Load() {
	// Load .env file if it exists (optional in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	accessExpiry, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_HOURS", "24"))
	refreshExpiry, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_HOURS", "168")) // 7 days

	AppConfig = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", "gorm"),
			Password: getEnv("DB_PASSWORD", "gorm"),
			DBName:   getEnv("DB_NAME", "gorm"),
			Port:     getEnv("DB_PORT", "9920"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Shanghai"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-secret-key-change-this"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key-change-this"),
			AccessExpiry:  time.Duration(accessExpiry) * time.Hour,
			RefreshExpiry: time.Duration(refreshExpiry) * time.Hour,
		},
		OAuth: OAuthConfig{
			RedirectBaseURL: getEnv("OAUTH_REDIRECT_BASE_URL", "http://localhost:8080"),
			Google: OAuthProvider{
				ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
				ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
				Enabled:      getEnv("GOOGLE_OAUTH_ENABLED", "false") == "true",
			},
			GitHub: OAuthProvider{
				ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
				ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
				Enabled:      getEnv("GITHUB_OAUTH_ENABLED", "false") == "true",
			},
			Microsoft: OAuthProvider{
				ClientID:     getEnv("MICROSOFT_CLIENT_ID", ""),
				ClientSecret: getEnv("MICROSOFT_CLIENT_SECRET", ""),
				Enabled:      getEnv("MICROSOFT_OAUTH_ENABLED", "false") == "true",
			},
			Facebook: OAuthProvider{
				ClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
				ClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
				Enabled:      getEnv("FACEBOOK_OAUTH_ENABLED", "false") == "true",
			},
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
