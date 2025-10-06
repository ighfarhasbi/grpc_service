package config

import (
	"log"
	"os"
)

type Config struct {
	DBUrl                string
	Port                 string
	InventoryServiceAddr string
}

func New() *Config {
	dbUrl := os.Getenv("DB_ORDER_URL")
	if dbUrl == "" {
		log.Fatal("DB_ORDER_URL environment variable is required")
	}

	port := os.Getenv("GRPC_ORDER_PORT")
	if port == "" {
		log.Fatal("GRPC_ORDER_PORT environment variable is required")
	}

	inventoryServiceAddr := os.Getenv("INVENTORY_SERVICE_ADDR")
	if inventoryServiceAddr == "" {
		log.Fatal("INVENTORY_SERVICE_ADDR environment variable is required")
	}

	return &Config{
		DBUrl:                dbUrl,
		Port:                 port,
		InventoryServiceAddr: inventoryServiceAddr,
	}
}
