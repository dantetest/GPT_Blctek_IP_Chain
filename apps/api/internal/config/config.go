package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct { Environment, APIAddr, MySQLDSN, RedisAddr, RedisPassword string; RedisDB int }

func Load() (Config, error) {
	redisDB, err := strconv.Atoi(env("REDIS_DB", "0")); if err != nil { return Config{}, fmt.Errorf("REDIS_DB must be an integer: %w", err) }
	return Config{Environment: env("APP_ENV", "development"), APIAddr: env("API_ADDR", ":8080"), MySQLDSN: env("MYSQL_DSN", "blctekip:blctekip@tcp(localhost:3306)/blctekip?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"), RedisAddr: env("REDIS_ADDR", "localhost:6379"), RedisPassword: os.Getenv("REDIS_PASSWORD"), RedisDB: redisDB}, nil
}
func env(key, fallback string) string { if value := os.Getenv(key); value != "" { return value }; return fallback }
