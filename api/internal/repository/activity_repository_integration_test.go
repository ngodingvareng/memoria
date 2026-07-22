package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moez-rd/memoria/internal/entity"
	"github.com/moez-rd/memoria/internal/repository"
	"github.com/moez-rd/memoria/internal/usecase"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("memoria_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(wait.
			ForLog("database system is ready to accept connections").
			WithOccurrence(2),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgContainer.Terminate(ctx) })

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	m, err := migrate.New("file://../../db/migrations", dsn)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}

func seedTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id`,
		"Test User", uuid.NewString()+"@example.com",
	).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestActivityRepository_WithTransaction_CommitsOnSuccess(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	var created *entity.Activity
	err := repo.WithTransaction(context.Background(), func(tx usecase.ActivityRepository) error {
		var err error
		created, err = tx.Create(context.Background(), &entity.Activity{
			UserID:          userID,
			Name:            "Morning cook",
			IsFixedSchedule: true,
		})
		return err
	})
	require.NoError(t, err)
	require.NotNil(t, created)

	// Confirm it's actually visible outside the transaction that created
	// it — i.e. it was really committed, not just returned in-memory.
	var count int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM activities WHERE id = $1`, created.ID,
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestActivityRepository_WithTransaction_RollsBackOnError(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	sentinelErr := errors.New("boom")

	err := repo.WithTransaction(context.Background(), func(tx usecase.ActivityRepository) error {
		_, err := tx.Create(context.Background(), &entity.Activity{
			UserID:          userID,
			Name:            "Should not persist",
			IsFixedSchedule: true,
		})
		if err != nil {
			return err
		}
		// Force the whole transaction to fail after the insert succeeded.
		return sentinelErr
	})
	require.ErrorIs(t, err, sentinelErr)

	var count int
	err = pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM activities WHERE name = $1`, "Should not persist",
	).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count, "insert should have been rolled back, not committed")
}

func TestActivityRepository_WithTransaction_GuardsAgainstNestedTx(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	err := repo.WithTransaction(context.Background(), func(tx usecase.ActivityRepository) error {
		// tx here is bound to a transaction already (its internal pool
		// field is nil). Calling WithTransaction on it again must be
		// rejected instead of silently opening a second, independent
		// transaction — see the guard in activity_repository.go.
		return tx.WithTransaction(context.Background(), func(usecase.ActivityRepository) error {
			return nil
		})
	})
	require.Error(t, err)

	_ = userID // seeded for consistency with the other tests; unused here
}
