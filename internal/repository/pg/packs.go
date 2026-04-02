package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PackRepository persists the ordered list of pack sizes used for calculations.
type PackRepository struct {
	pool *pgxpool.Pool
}

// NewPackRepository constructs a repository backed by pool.
func NewPackRepository(pool *pgxpool.Pool) *PackRepository {
	return &PackRepository{pool: pool}
}

// List returns all pack sizes sorted ascending (ORDER BY size).
func (r *PackRepository) List(ctx context.Context) ([]int, error) {
	rows, err := r.pool.Query(ctx, `SELECT size FROM pack_sizes ORDER BY size ASC`)
	if err != nil {
		return nil, fmt.Errorf("list pack sizes: %w", err)
	}
	defer rows.Close()

	var out []int
	for rows.Next() {
		var s int
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("scan pack size: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ReplaceAll replaces the full set of pack sizes in a single transaction.
// Duplicate sizes in the slice are rejected (unique constraint).
func (r *PackRepository) ReplaceAll(ctx context.Context, sizes []int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM pack_sizes`); err != nil {
		return fmt.Errorf("clear pack sizes: %w", err)
	}

	for _, s := range sizes {
		if _, err := tx.Exec(ctx, `INSERT INTO pack_sizes (size) VALUES ($1)`, s); err != nil {
			return fmt.Errorf("insert pack size %d: %w", s, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Pool exposes the underlying pool for graceful shutdown (optional).
func (r *PackRepository) Pool() *pgxpool.Pool { return r.pool }
