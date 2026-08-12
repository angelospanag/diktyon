package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/angelospanag/diktyon/internal/api"
	"github.com/angelospanag/diktyon/internal/cache"
	"github.com/angelospanag/diktyon/internal/config"
	"github.com/angelospanag/diktyon/internal/providers/gr"
	"github.com/angelospanag/diktyon/internal/providers/uk"
	"github.com/angelospanag/diktyon/internal/registry"
	"github.com/danielgtaylor/huma/v2/humacli"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	cli := humacli.New(func(hooks humacli.Hooks, _ *struct{}) {
		// Load .env if present; silently ignored in production where env vars are injected directly.
		_ = godotenv.Load()

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config error: %v\n", err)
			os.Exit(1)
		}

		logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: cfg.LogLevel,
		}))
		slog.SetDefault(logger)

		var cacheClient cache.Cache
		if cfg.RedisURL != "" {
			cacheClient, err = cache.NewRedis(cfg.RedisURL, cfg.CacheTTL)
			if err != nil {
				slog.Warn("Redis unavailable — falling back to in-memory cache", "error", err)
				cacheClient = cache.NewMemory()
			} else {
				slog.Info("cache: Redis", "url", cfg.RedisURL)
			}
		} else {
			slog.Info("cache: in-memory (set REDIS_URL to use Redis)")
			cacheClient = cache.NewMemory()
		}

		chClient := uk.NewClient(
			cfg.CompaniesHouseAPIKey,
			cfg.RateLimitRPS,
			cfg.RateLimitBurst,
		)

		reg := registry.New()
		reg.Register("uk", uk.New(chClient))

		if cfg.GEMIAPIKey != "" {
			reg.Register("gr", gr.New(gr.NewClient(cfg.GEMIAPIKey)))
			slog.Info("provider registered", "country", "gr")
		}

		router, _ := api.NewRouter(reg, cacheClient)

		// chi ships gzip and deflate only; brotli is added here and takes precedence.
		compressor := chimw.NewCompressor(6)
		compressor.SetEncoder("br", func(w io.Writer, level int) io.Writer {
			return brotli.NewWriterLevel(w, level)
		})

		// Chain applies outermost first, so Recoverer also catches panics raised inside the
		// compressor — otherwise a panic mid-write leaves a truncated compressed body.
		handler := chi.Chain(chimw.Recoverer, compressor.Handler).Handler(router)

		srv := &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		hooks.OnStart(func() {
			slog.Info("server listening", "addr", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				slog.Error("server error", "error", err)
			}
		})

		hooks.OnStop(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				slog.Error("graceful shutdown failed", "error", err)
			}
			slog.Info("server stopped")
		})
	})

	// `openapi` prints the generated spec to stdout without starting the server
	// or touching the database — the codegen/CI source for api/openapi.yaml.
	cli.Root().AddCommand(&cobra.Command{
		Use:   "openapi",
		Short: "Print the OpenAPI spec (YAML) to stdout",
		Run: func(_ *cobra.Command, _ []string) {
			reg := registry.New()
			reg.Register("uk", uk.New(uk.NewClient("__schema__", 1, 1)))
			_, humaAPI := api.NewRouter(reg, cache.NewNoop())
			b, err := humaAPI.OpenAPI().YAML()
			if err != nil {
				fmt.Fprintf(os.Stderr, "schema error: %v\n", err)
				os.Exit(1)
			}
			// Print (not Println): YAML() already ends with a newline; an extra
			// one adds a trailing blank line that would defeat a CI drift check.
			fmt.Print(string(b))
		},
	})

	cli.Run()
}
