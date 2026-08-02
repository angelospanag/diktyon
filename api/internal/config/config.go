package config

import (
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port                 int           `env:"PORT"                             envDefault:"8080"`
	CompaniesHouseAPIKey string        `env:"COMPANIES_HOUSE_API_KEY,required"`
	GEMIAPIKey           string        `env:"GEMI_API_KEY"` // empty → GEMI provider disabled
	RedisURL             string        `env:"REDIS_URL"`    // empty → in-memory cache
	CacheTTL             time.Duration `env:"CACHE_TTL"                        envDefault:"24h"`
	LogLevel             slog.Level    `env:"LOG_LEVEL"                        envDefault:"INFO"`
	RateLimitRPS         float64       `env:"RATE_LIMIT_RPS"                   envDefault:"2"`
	RateLimitBurst       int           `env:"RATE_LIMIT_BURST"                 envDefault:"10"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
