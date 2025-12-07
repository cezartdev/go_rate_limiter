package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
	env := os.Getenv("APP_ENV")

	if env == "" {
		env = "development"
	}

	_ = godotenv.Load()

	envFile := fmt.Sprintf(".env.%s", env)

	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Overload(envFile); err != nil {
			return err
		}
	}

	upstream := os.Getenv("UPSTREAM_URL")
	if upstream == "" {
		log.Fatal("UPSTREAM_URL is required")
	}

	return nil
}
