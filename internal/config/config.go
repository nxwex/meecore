package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	DatabaseDSN string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error load env: %v", err)
	}

	return Config{
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
		DatabaseDSN: os.Getenv("DATABASE_DSN"),
	}
}
