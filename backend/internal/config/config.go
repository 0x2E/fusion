package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBPath      string
	DatabaseURL string // PostgreSQL connection string (mutually exclusive with DBPath)
	Password    string // Plaintext password from env
	Port        int

	CORSAllowedOrigins []string // Allowed Origins for CORS. Empty means allow all.
	TrustedProxies     []string // Trusted reverse proxies for client IP resolution. Empty disables proxy trust.
	AllowPrivateFeeds  bool     // Allow pulling private/localhost feed URLs.

	PullInterval    int // Pull interval in seconds (default: 1800 = 30 min)
	PullTimeout     int // Request timeout in seconds (default: 30)
	PullConcurrency int // Max concurrent pulls (default: 10)
	PullMaxBackoff  int // Global max scheduling delay in seconds (default: 172800 = 48 hours)

	LoginRateLimit int // Max failed login attempts per window (default: 10)
	LoginWindow    int // Login rate limit window in seconds (default: 60)
	LoginBlock     int // Login block duration in seconds (default: 300)

	LogLevel  string // Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
	LogFormat string // Log format: text, json, auto (default: auto)

	// OIDC Configuration (optional, enabled when OIDCIssuer is set)
	OIDCIssuer       string // OIDC provider URL
	OIDCClientID     string // OAuth2 client ID
	OIDCClientSecret string // OAuth2 client secret
	OIDCAllowedUser  string // Optional: restrict to specific user identity (email or sub)
}

func Load() (*Config, error) {
	// Backward compatible env vars:
	// - DB (legacy) -> REEDME_DB_PATH
	// - PASSWORD (legacy) -> REEDME_PASSWORD
	// - PORT (legacy) -> REEDME_PORT
	dbPath := os.Getenv("REEDME_DB_PATH")
	if dbPath == "" {
		dbPath = os.Getenv("DB")
	}
	if dbPath == "" {
		dbPath = "reedme.db"
	}

	password := os.Getenv("REEDME_PASSWORD")
	if password == "" {
		password = os.Getenv("PASSWORD")
	}

	allowEmptyPassword, err := getEnvBool("REEDME_ALLOW_EMPTY_PASSWORD", false)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(password) == "" && !allowEmptyPassword {
		return nil, fmt.Errorf("REEDME_PASSWORD is required (or set REEDME_ALLOW_EMPTY_PASSWORD=true)")
	}

	port := os.Getenv("REEDME_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid REEDME_PORT: %w", err)
	}
	if parsedPort <= 0 || parsedPort > 65535 {
		return nil, fmt.Errorf("invalid REEDME_PORT: must be in range 1-65535")
	}

	pullInterval, err := getEnvInt("REEDME_PULL_INTERVAL", 1800, 1)
	if err != nil {
		return nil, err
	}
	pullTimeout, err := getEnvInt("REEDME_PULL_TIMEOUT", 30, 1)
	if err != nil {
		return nil, err
	}
	pullConcurrency, err := getEnvInt("REEDME_PULL_CONCURRENCY", 10, 1)
	if err != nil {
		return nil, err
	}
	pullMaxBackoff, err := getEnvInt("REEDME_PULL_MAX_BACKOFF", 172800, 1)
	if err != nil {
		return nil, err
	}

	loginRateLimit, err := getEnvInt("REEDME_LOGIN_RATE_LIMIT", 10, 1)
	if err != nil {
		return nil, err
	}
	loginWindow, err := getEnvInt("REEDME_LOGIN_WINDOW", 60, 1)
	if err != nil {
		return nil, err
	}
	loginBlock, err := getEnvInt("REEDME_LOGIN_BLOCK", 300, 1)
	if err != nil {
		return nil, err
	}

	corsAllowedOrigins := parseCSVEnv(os.Getenv("REEDME_CORS_ALLOWED_ORIGINS"))
	trustedProxies := parseCSVEnv(os.Getenv("REEDME_TRUSTED_PROXIES"))

	allowPrivateFeeds, err := getEnvBool("REEDME_ALLOW_PRIVATE_FEEDS", false)
	if err != nil {
		return nil, err
	}

	logLevel := os.Getenv("REEDME_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	logFormat := os.Getenv("REEDME_LOG_FORMAT")
	if logFormat == "" {
		logFormat = "auto"
	}

	return &Config{
		DBPath:             dbPath,
		DatabaseURL:        os.Getenv("REEDME_DATABASE_URL"),
		Password:           password,
		Port:               parsedPort,
		CORSAllowedOrigins: corsAllowedOrigins,
		TrustedProxies:     trustedProxies,
		AllowPrivateFeeds:  allowPrivateFeeds,
		PullInterval:       pullInterval,
		PullTimeout:        pullTimeout,
		PullConcurrency:    pullConcurrency,
		PullMaxBackoff:     pullMaxBackoff,
		LoginRateLimit:     loginRateLimit,
		LoginWindow:        loginWindow,
		LoginBlock:         loginBlock,
		LogLevel:           logLevel,
		LogFormat:          logFormat,

		OIDCIssuer:       os.Getenv("REEDME_OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("REEDME_OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("REEDME_OIDC_CLIENT_SECRET"),
		OIDCAllowedUser:  os.Getenv("REEDME_OIDC_ALLOWED_USER"),
	}, nil
}

func getEnvInt(key string, defaultVal, minVal int) (int, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", key, err)
	}
	if parsed < minVal {
		return 0, fmt.Errorf("invalid %s: must be >= %d", key, minVal)
	}
	return parsed, nil
}

func getEnvBool(key string, defaultVal bool) (bool, error) {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal, nil
	}
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false, fmt.Errorf("invalid %s: %w", key, err)
	}
	return parsed, nil
}

func parseCSVEnv(val string) []string {
	if strings.TrimSpace(val) == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		values = append(values, part)
	}

	if len(values) == 0 {
		return nil
	}

	return values
}
