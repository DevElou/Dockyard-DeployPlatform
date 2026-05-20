package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DSN      string
	MaxConns int32
}

func LoadConfigFromEnv() (Config, error) {
	dsn := os.Getenv("DOCKYARD_DATABASE_URL")
	if dsn == "" {
		return Config{}, fmt.Errorf("DOCKYARD_DATABASE_URL is required")
	}
	return Config{DSN: dsn, MaxConns: 10}, nil
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return pool, nil
}
