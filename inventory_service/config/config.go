package config

import (
	"log"
	"os"
)

type Config struct {
	DBUrl string
	Port  string
}

func New() *Config {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL environment variable is required")
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		log.Fatal("GRPC_PORT environment variable is required")
	}

	return &Config{
		DBUrl: dbUrl,
		Port:  port,
	}
}
