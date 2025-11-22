package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/khatibomar/kaedama/api"
	"github.com/khatibomar/kaedama/proxy"
)

type Config struct {
	Port           int    `default:"4140"        envconfig:"PORT"`
	Host           string `default:"0.0.0.0"     envconfig:"HOST"`
	Env            string `default:"development" envconfig:"ENV"`
	LogLevel       string `default:"debug"       envconfig:"LOG_LEVEL"`
	CacheTTL       int    `default:"300"         envconfig:"CACHE_TTL"`
	RequestTimeout int    `default:"30000"       envconfig:"REQUEST_TIMEOUT"`
	CORSOrigins    string `default:"*"           envconfig:"CORS_ORIGINS"`
}

func main() {
	var (
		cfg   Config
		level slog.LevelVar
	)
	level.Set(slog.LevelError)
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &level,
	}))
	if err := envconfig.Process("", &cfg); err != nil {
		log.Error("Failed to parse configs", slog.Any("error", err))
		os.Exit(1)
	}
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		log.Error("Failed to parse log level", slog.Any("error", err))
		os.Exit(1)
	}

	if err := realMain(log, cfg); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Failed to startup server", slog.Any("error", err))
		os.Exit(1)
	}
}

func realMain(log *slog.Logger, cfg Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cacheTTL := time.Duration(cfg.CacheTTL) * time.Second
	proxyService := proxy.New()
	handler := api.New(log, proxyService, cfg.CORSOrigins, cacheTTL)

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.InfoContext(ctx, "starting server", "addr", server.Addr, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		// graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)

	case err := <-errCh:
		// startup or runtime failure
		return fmt.Errorf("server failed: %w", err)
	}
}
