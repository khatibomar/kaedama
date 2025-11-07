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
	Port           int    `default:"3000"        envconfig:"PORT"`
	Host           string `default:"0.0.0.0"     envconfig:"HOST"`
	Env            string `default:"development" envconfig:"ENV"`
	LogLevel       string `default:"debug"       envconfig:"LOG_LEVEL"`
	CacheTTL       int    `default:"300"         envconfig:"CACHE_TTL"`
	RequestTimeout int    `default:"30000"       envconfig:"REQUEST_TIMEOUT"`
}

func main() {
	var cfg Config
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelError,
	}))
	if err := envconfig.Process("", &cfg); err != nil {
		log.Error("Failed to parse configs", slog.Any("error", err))
		os.Exit(1)
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		log.Error("Failed to parse log level", slog.Any("error", err))
		os.Exit(1)
	}
	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))

	if err := realMain(log, cfg); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("Failed to startup server", slog.Any("error", err))
		os.Exit(1)
	}
}

func realMain(log *slog.Logger, cfg Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	proxyService := proxy.New()
	handler := api.New(log, proxyService)

	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
	}

	go func() {
		log.InfoContext(ctx, "starting server", "addr", server.Addr, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error(
				"failed to listen and serve",
				slog.String("address", address),
				slog.Any("error", err),
			)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	return nil
}
