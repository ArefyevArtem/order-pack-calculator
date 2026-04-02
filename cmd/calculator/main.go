// Command calculator starts Order Pack Calculator HTTP API.
package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"order-pack-calculator/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cfg, err := app.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	return app.New(cfg).Run(ctx)
}
