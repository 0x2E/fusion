package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	DBPath        string
	Password      string // Plaintext password from env
	Port          int
	FeverUsername string // Username used to derive Fever API key.

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
	OIDCRedirectURI  string // Callback URL (required when OIDC is enabled)
	OIDCAllowedUser  string // Optional: restrict to specific user identity (email or sub)
}

func Load() (*Config, error) {
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

	allowEmptyPassword, err := getEnvBool("FUSION_ALLOW_EMPTY_PASSWORD", false)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(password) == "" && !allowEmptyPassword {
		return nil, fmt.Errorf("FUSION_PASSWORD is required (or set FUSION_ALLOW_EMPTY_PASSWORD=true)")
	}

	port := os.Getenv("FUSION_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	parsedPort, err := parsePort(port)
	if err != nil {
		return nil, fmt.Errorf("invalid FUSION_PORT: %w", err)
	}
	if parsedPort <= 0 || parsedPort > 65535 {
		return nil, fmt.Errorf("invalid FUSION_PORT: must be in range 1-65535")
	}

	pullInterval, err := getEnvInt("FUSION_PULL_INTERVAL", 1800, 1)
	if err != nil {
		return nil, err
	}
	pullTimeout, err := getEnvInt("FUSION_PULL_TIMEOUT", 30, 1)
	if err != nil {
		return nil, err
	}
	pullConcurrency, err := getEnvInt("FUSION_PULL_CONCURRENCY", 10, 1)
	if err != nil {
		return nil, err
	}
	pullMaxBackoff, err := getEnvInt("FUSION_PULL_MAX_BACKOFF", 172800, 1)
	if err != nil {
		return nil, err
	}

	loginRateLimit, err := getEnvInt("FUSION_LOGIN_RATE_LIMIT", 10, 1)
	if err != nil {
		return nil, err
	}
	loginWindow, err := getEnvInt("FUSION_LOGIN_WINDOW", 60, 1)
	if err != nil {
		return nil, err
	}
	loginBlock, err := getEnvInt("FUSION_LOGIN_BLOCK", 300, 1)
	if err != nil {
		return nil, err
	}

	corsAllowedOrigins := parseCSVEnv(os.Getenv("FUSION_CORS_ALLOWED_ORIGINS"))
	trustedProxies := parseCSVEnv(os.Getenv("FUSION_TRUSTED_PROXIES"))

	allowPrivateFeeds, err := getEnvBool("FUSION_ALLOW_PRIVATE_FEEDS", false)
	if err != nil {
		return nil, err
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
		DBPath:             dbPath,
		Password:           password,
		Port:               parsedPort,
		FeverUsername:      getEnvString("FUSION_FEVER_USERNAME", "fusion"),
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

		OIDCIssuer:       os.Getenv("FUSION_OIDC_ISSUER"),
		OIDCClientID:     os.Getenv("FUSION_OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("FUSION_OIDC_CLIENT_SECRET"),
		OIDCRedirectURI:  os.Getenv("FUSION_OIDC_REDIRECT_URI"),
		OIDCAllowedUser:  os.Getenv("FUSION_OIDC_ALLOWED_USER"),
	}, nil
}

func getEnvString(key, defaultVal string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultVal
	}

	return val
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

// parsePort accepts plain numeric ports and Kubernetes service-link URL values
// such as tcp://10.43.157.55:8080.
func parsePort(val string) (int, error) {
	trimmed := strings.TrimSpace(val)
	parsed, err := strconv.Atoi(trimmed)
	if err == nil {
		return parsed, nil
	}

	if !strings.Contains(trimmed, "://") {
		return 0, err
	}

	parsedURL, err := url.Parse(trimmed)
	if err != nil {
		return 0, err
	}

	port := parsedURL.Port()
	if port == "" {
		return 0, fmt.Errorf("missing port")
	}

	return strconv.Atoi(port)
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
