package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBPath   string
	Password string // Plaintext password from env
	Host     string // TODO parse and use
	Port     int

	PullInterval    int // Pull interval in seconds (default: 1800 = 30 min)
	PullTimeout     int // Request timeout in seconds (default: 30)
	PullConcurrency int // Max concurrent pulls (default: 10)
	PullMaxBackoff  int // Max backoff time in seconds (default: 604800 = 7 days)

	LogLevel  string // Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
	LogFormat string // Log format: text, json, auto (default: auto)

	// OIDC Configuration (optional, enabled when OIDCIssuer is set)
	OIDCIssuer       string // OIDC provider URL
	OIDCClientID     string // OAuth2 client ID
	OIDCClientSecret string // OAuth2 client secret
	OIDCRedirectURI  string // Callback URL (default: auto-detect from Host header)
	OIDCAllowedUser  string // Optional: restrict to specific user identity (email or sub)
}

func Load() *Config {
	// Backward compatible env vars:
	// - DB (legacy) -> FUSION_DB_PATH
	// - PASSWORD (legacy) -> FUSION_PASSWORD
	// - PORT (legacy) -> FUSION_PORT
	dbPath := os.Getenv("FUSION_DB_PATH")
	if dbPath == "" {
		dbPath = os.Getenv("DB")
	}
	if dbPath == "" {
		dbPath = "fusion.db"
	}

	password := os.Getenv("FUSION_PASSWORD")
	if password == "" {
		password = os.Getenv("PASSWORD")
	}
	if password == "" {
		password = "admin" // TODO allow empty password
	}

	port := os.Getenv("FUSION_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	logLevel := os.Getenv("FUSION_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	logFormat := os.Getenv("FUSION_LOG_FORMAT")
	if logFormat == "" {
		logFormat = "auto"
	}

	return &Config{
		DBPath:          dbPath,
		Password:        password,
		Port:            parsedPort,
		PullInterval:    getEnvInt("FUSION_PULL_INTERVAL", 1800),
		PullTimeout:     getEnvInt("FUSION_PULL_TIMEOUT", 30),
		PullConcurrency: getEnvInt("FUSION_PULL_CONCURRENCY", 10),
		PullMaxBackoff:  getEnvInt("FUSION_PULL_MAX_BACKOFF", 604800),
		LogLevel:        logLevel,
		LogFormat:       logFormat,

		OIDCIssuer:       os.Getenv("FUSION_OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("FUSION_OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("FUSION_OIDC_CLIENT_SECRET"),
		OIDCRedirectURI:  os.Getenv("FUSION_OIDC_REDIRECT_URI"),
		OIDCAllowedUser:  os.Getenv("FUSION_OIDC_ALLOWED_USER"),
	}
}

func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
