package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/khatibomar/kaedama/api"
	"github.com/khatibomar/kaedama/proxy"
)

type MemorySize int64

var memorySizeRegex = regexp.MustCompile(`^([+-]?[0-9.]+([eE][+-]?[0-9]+)?)([A-Za-z]*)$`)

func (m *MemorySize) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return fmt.Errorf("empty memory size")
	}

	matches := memorySizeRegex.FindStringSubmatch(s)
	if matches == nil {
		return fmt.Errorf("invalid memory size format: %s", s)
	}

	numStr := matches[1]
	suffix := matches[3]

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return fmt.Errorf("invalid memory size value: %s", numStr)
	}

	var multiplier float64 = 1
	switch suffix {
	case "", "B", "b":
		multiplier = 1
	case "K", "k":
		multiplier = 1e3
	case "M", "m":
		multiplier = 1e6
	case "G", "g":
		multiplier = 1e9
	case "T", "t":
		multiplier = 1e12
	case "P", "p":
		multiplier = 1e15
	case "E", "e":
		multiplier = 1e18
	case "Ki", "ki":
		multiplier = 1024
	case "Mi", "mi":
		multiplier = 1024 * 1024
	case "Gi", "gi":
		multiplier = 1024 * 1024 * 1024
	case "Ti", "ti":
		multiplier = 1024 * 1024 * 1024 * 1024
	case "Pi", "pi":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
	case "Ei", "ei":
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return fmt.Errorf("unrecognized memory suffix: %s", suffix)
	}

	*m = MemorySize(val * multiplier)
	return nil
}

type Config struct {
	Port           int    `default:"4140"        envconfig:"PORT"`
	Host           string `default:"0.0.0.0"     envconfig:"HOST"`
	Env            string `default:"development" envconfig:"ENV"`
	LogLevel       string `default:"debug"       envconfig:"LOG_LEVEL"`
	CacheTTL       int    `default:"300"         envconfig:"CACHE_TTL"`
	MaxCacheSize   MemorySize `default:"104857600"   envconfig:"MAX_CACHE_SIZE"`
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
	handler := api.New(log, proxyService, cfg.CORSOrigins, cacheTTL, int64(cfg.MaxCacheSize))

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
