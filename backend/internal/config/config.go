package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBPath   string
	Password string
	Host     string // TODO parse and use
	Port     int
}

func Load() *Config {
	dbPath := os.Getenv("FUSION_DB_PATH")
	if dbPath == "" {
		dbPath = "fusion.db"
	}

	password := os.Getenv("FUSION_PASSWORD")
	if password == "" {
		password = "admin" // TODO allow empty password
	}

	port := os.Getenv("FUSION_PORT")
	if port == "" {
		port = "8080"
	}
	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}

	return &Config{
		DBPath:   dbPath,
		Password: password,
		Port:     parsedPort,
	}
}
