//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	repopg "order-pack-calculator/internal/repository/pg"
)

// PostgreSQL via Testcontainers (postgres:16-alpine). Requires Docker.
// Not part of default `go test ./...` — use `make test-integration` or `-tags=integration`.

func startPostgresRepo(t *testing.T) *repopg.PackRepository {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)

	ctr, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("packs"),
		tcpostgres.WithUsername("app"),
		tcpostgres.WithPassword("app"),
		tcpostgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		termCtx, termCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer termCancel()
		require.NoError(t, ctr.Terminate(termCtx))
	})

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := repopg.NewPool(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	require.NoError(t, repopg.Migrate(ctx, pool))
	return repopg.NewPackRepository(pool)
}

func TestPackRepository_ReplaceAll_List_ordering(t *testing.T) {
	repo := startPostgresRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	require.NoError(t, repo.ReplaceAll(ctx, []int{53, 23, 31}))
	got, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, []int{23, 31, 53}, got)

	require.NoError(t, repo.ReplaceAll(ctx, []int{100, 20}))
	got, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Equal(t, []int{20, 100}, got)
}

func TestPackRepository_ReplaceAll_duplicate_rejected(t *testing.T) {
	repo := startPostgresRepo(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := repo.ReplaceAll(ctx, []int{7, 7})
	require.Error(t, err)
}
