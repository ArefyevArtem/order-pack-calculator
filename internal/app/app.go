package app

import (
	"context"
	"fmt"
	"log/slog"

	aphttp "order-pack-calculator/internal/api/http"
	calcctrl "order-pack-calculator/internal/api/http/controller/calculator"
	healthctrl "order-pack-calculator/internal/api/http/controller/health"
	postgres "order-pack-calculator/internal/repository/pg"
	"order-pack-calculator/internal/usecase/calculator"
)

type App struct {
	cfg Config
}

func New(cfg Config) *App {
	return &App{cfg: cfg}
}

// Run opens PostgreSQL, applies migrations, wires the calculator use case and HTTP server,
// then listens until ctx is canceled (graceful shutdown).
func (a *App) Run(ctx context.Context) error {
	log := slog.Default()

	pool, err := postgres.NewPool(ctx, a.cfg.Database.DatabaseURL())
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	packsRepo := postgres.NewPackRepository(pool)
	uc := calculator.NewService(packsRepo)

	srv := aphttp.NewServer(a.cfg.Server)
	srv.AddController(
		healthctrl.New(pool),
		calcctrl.New(uc, log),
	)

	log.Info("listening", "addr", a.cfg.Server.Addr())
	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
