package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment      string
	APIAddr          string
	MySQLDSN         string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
	JWTIssuer        string
	JWTSecret        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	AutoVerifyEmail  bool
	LoginMaxAttempts int
	LoginLockTTL     time.Duration
}

func Load() (Config, error) {
	redisDB, err := strconv.Atoi(env("REDIS_DB", "0"))
	if err != nil {
		return Config{}, fmt.Errorf("REDIS_DB must be an integer: %w", err)
	}

	accessTTL, err := time.ParseDuration(env("ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("ACCESS_TOKEN_TTL must be a duration: %w", err)
	}
	refreshTTL, err := time.ParseDuration(env("REFRESH_TOKEN_TTL", "720h"))
	if err != nil {
		return Config{}, fmt.Errorf("REFRESH_TOKEN_TTL must be a duration: %w", err)
	}
	lockTTL, err := time.ParseDuration(env("LOGIN_LOCK_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("LOGIN_LOCK_TTL must be a duration: %w", err)
	}
	maxAttempts, err := strconv.Atoi(env("LOGIN_MAX_ATTEMPTS", "5"))
	if err != nil || maxAttempts < 1 {
		return Config{}, fmt.Errorf("LOGIN_MAX_ATTEMPTS must be a positive integer")
	}
	autoVerify, err := strconv.ParseBool(env("AUTO_VERIFY_EMAIL", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("AUTO_VERIFY_EMAIL must be a boolean: %w", err)
	}

	environment := env("APP_ENV", "development")
	jwtSecret := env("JWT_SECRET", "development-only-change-this-secret-32chars")
	if environment != "development" && len(jwtSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters outside development")
	}

	return Config{
		Environment:      environment,
		APIAddr:          env("API_ADDR", ":8080"),
		MySQLDSN:         env("MYSQL_DSN", "blctekip:blctekip@tcp(localhost:3306)/blctekip?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"),
		RedisAddr:        env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		RedisDB:          redisDB,
		JWTIssuer:        env("JWT_ISSUER", "blctekip-api"),
		JWTSecret:        jwtSecret,
		AccessTokenTTL:   accessTTL,
		RefreshTokenTTL:  refreshTTL,
		AutoVerifyEmail:  autoVerify,
		LoginMaxAttempts: maxAttempts,
		LoginLockTTL:     lockTTL,
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
