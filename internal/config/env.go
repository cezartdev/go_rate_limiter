package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)


type Config struct {
	Env             string
	BindAddr        string
	UpstreamURL     string
	RateLimitRPS    float64
	Burst           float64
	BucketTTL       time.Duration
	CleanupInterval time.Duration
}


func Load() (*Config, error) {
	// Determine environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	
	_ = godotenv.Load()

	// Load environment-specific file (optional, overrides base)
	envFile := fmt.Sprintf(".env.%s", env)
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Overload(envFile); err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", envFile, err)
		}
	}

	cfg := &Config{
		Env: env,
	}

	// Required: UPSTREAM_URL
	cfg.UpstreamURL = os.Getenv("UPSTREAM_URL")
	if cfg.UpstreamURL == "" {
		return nil, fmt.Errorf("UPSTREAM_URL is required")
	}

	// Optional: BIND_ADDR (default :8080)
	cfg.BindAddr = getEnvOrDefault("BIND_ADDR", ":8080")

	// Optional: RATE_LIMIT_RPS (default 5.0)
	cfg.RateLimitRPS = getEnvAsFloat("RATE_LIMIT_RPS", 5.0)

	// Optional: BURST (default 10.0)
	cfg.Burst = getEnvAsFloat("BURST", 10.0)

	// Optional: BUCKET_TTL in seconds (default 300 = 5 minutes)
	ttlSecs := getEnvAsInt("BUCKET_TTL", 300)
	cfg.BucketTTL = time.Duration(ttlSecs) * time.Second

	// Optional: CLEANUP_INTERVAL in seconds (default 60 = 1 minute)
	cleanupSecs := getEnvAsInt("CLEANUP_INTERVAL", 60)
	cfg.CleanupInterval = time.Duration(cleanupSecs) * time.Second

	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
