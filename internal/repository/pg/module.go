package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config is filled via envconfig under app.AppName and nested key DATABASE
// (e.g. ORDER_PACK_CALCULATOR_DATABASE_HOST).
type Config struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     string `envconfig:"PORT" default:"5432"`
	User     string `envconfig:"USER" default:"postgres"`
	Password string `envconfig:"PASSWORD" default:"postgres"`
	DBName   string `envconfig:"NAME" default:"packs"`
	SSLMode  string `envconfig:"SSLMODE" default:"disable"`
}

// DatabaseURL builds a PostgreSQL connection URI for pgx.
func (c Config) DatabaseURL() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%s", c.Host, c.Port),
		Path:   "/" + c.DBName,
	}
	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

// NewPool opens a connection pool from a DSN (e.g. cfg.DatabaseURL()).
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
