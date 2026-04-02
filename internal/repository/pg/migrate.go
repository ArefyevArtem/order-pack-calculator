package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const migrateSQL = `
CREATE TABLE IF NOT EXISTS pack_sizes (
	id   BIGSERIAL PRIMARY KEY,
	size INT NOT NULL UNIQUE,
	CONSTRAINT pack_sizes_size_positive CHECK (size > 0)
);
`

// Migrate applies schema required for this service. Safe to run on every startup.
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, migrateSQL)
	return err
}
