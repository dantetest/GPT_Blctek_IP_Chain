package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/config"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/platform/cache"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/platform/database"
	"github.com/dantetest/GPT_Blctek_IP_Chain/apps/api/internal/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil { logger.Error("load configuration", "error", err); os.Exit(1) }
	db, err := database.Open(cfg.MySQLDSN)
	if err != nil { logger.Error("connect mysql", "error", err); os.Exit(1) }
	defer db.Close()
	redisClient, err := cache.Open(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil { logger.Error("connect redis", "error", err); os.Exit(1) }
	defer redisClient.Close()

	httpServer := &http.Server{Addr: cfg.APIAddr, Handler: server.New(logger, db, redisClient), ReadHeaderTimeout: 5*time.Second, ReadTimeout: 15*time.Second, WriteTimeout: 30*time.Second, IdleTimeout: 60*time.Second}
	go func() {
		logger.Info("api server started", "address", cfg.APIAddr, "environment", cfg.Environment)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { logger.Error("api server stopped unexpectedly", "error", err); os.Exit(1) }
	}()
	stop := make(chan os.Signal, 1); signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM); <-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil { logger.Error("graceful shutdown", "error", err); os.Exit(1) }
}
