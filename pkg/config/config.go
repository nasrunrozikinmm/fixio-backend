package config

import (
	"log"

	"github.com/spf13/viper"
)

// Environment holds all application configuration values loaded from .env file
type Environment struct {
	// Application
	Environment       string `mapstructure:"ENV"`
	AppHost           string `mapstructure:"APP_HOST"`
	AppPort           uint16 `mapstructure:"APP_PORT"`
	ServerReadTimeout int    `mapstructure:"SERVER_READ_TIMEOUT"`

	// PostgreSQL
	PgHost string `mapstructure:"PG_DB_HOST"`
	PgPort string `mapstructure:"PG_DB_PORT"`
	PgName string `mapstructure:"PG_DB_NAME"`
	PgUser string `mapstructure:"PG_DB_USER"`
	PgPass string `mapstructure:"PG_DB_PASS"`

	// JWT
	JwtSecret          string `mapstructure:"JWT_SECRET"`
	JwtTokenExpiryDays int    `mapstructure:"JWT_TOKEN_EXPIRY_DAYS"`

	// Redis
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`

	// OAuth - Google
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `mapstructure:"GOOGLE_REDIRECT_URL"`

	// Frontend
	FrontendURL string `mapstructure:"FRONTEND_URL"`

	// MinIO (S3-compatible object storage)
	MinioEndpoint  string `mapstructure:"MINIO_ENDPOINT"`
	MinioAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinioBucket    string `mapstructure:"MINIO_BUCKET"`
	MinioUseSSL    bool   `mapstructure:"MINIO_USE_SSL"`
}

// NewEnvironment loads environment variables from .env file and returns an Environment struct
func NewEnvironment() *Environment {
	env := &Environment{}

	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	if err := viper.Unmarshal(env); err != nil {
		log.Fatalf("Failed to unmarshal environment config: %v", err)
	}

	// Set defaults
	if env.AppHost == "" {
		env.AppHost = "localhost"
	}
	if env.AppPort == 0 {
		env.AppPort = 8080
	}
	if env.ServerReadTimeout == 0 {
		env.ServerReadTimeout = 10
	}
	if env.JwtTokenExpiryDays == 0 {
		env.JwtTokenExpiryDays = 7
	}
	if env.Environment == "" {
		env.Environment = "development"
	}
	if env.MinioEndpoint == "" {
		env.MinioEndpoint = "localhost:9000"
	}
	if env.MinioBucket == "" {
		env.MinioBucket = "fixio"
	}

	return env
}

// IsDevelopment returns true if running in development environment
func (e *Environment) IsDevelopment() bool {
	return e.Environment == "development"
}

// IsProduction returns true if running in production environment
func (e *Environment) IsProduction() bool {
	return e.Environment == "production"
}
