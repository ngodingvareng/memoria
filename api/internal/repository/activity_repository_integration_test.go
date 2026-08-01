//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/require"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/usecase"
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

func TestActivityRepository_Update_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Activity{
		UserID: userID, Name: "Original name", IsFixedSchedule: true,
	})
	require.NoError(t, err)

	newColor := "#123456"
	updated, err := repo.Update(context.Background(), &entity.Activity{
		ID: created.ID, UserID: userID, Name: "Updated name", ColorHex: &newColor,
	})

	require.NoError(t, err)
	require.Equal(t, "Updated name", updated.Name)
	require.Equal(t, "#123456", *updated.ColorHex)
	// is_fixed_schedule must be untouched by Update — it's not part of
	// UpdateActivityParams.
	require.True(t, updated.IsFixedSchedule)
}

func TestActivityRepository_Update_WrongUserID_NotFound(t *testing.T) {
	// UserID in the WHERE clause doubles as the ownership check — a
	// mismatched userID must behave identically to a nonexistent id.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Activity{
		UserID: userID, Name: "Mine", IsFixedSchedule: true,
	})
	require.NoError(t, err)

	_, err = repo.Update(context.Background(), &entity.Activity{
		ID: created.ID, UserID: otherUserID, Name: "Hijacked",
	})

	require.ErrorIs(t, err, errs.ErrNotFound)
}

func TestActivityRepository_Delete_Success(t *testing.T) {
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Activity{
		UserID: userID, Name: "To delete", IsFixedSchedule: true,
	})
	require.NoError(t, err)

	err = repo.Delete(context.Background(), created.ID, userID)
	require.NoError(t, err)

	var deletedAt *time.Time
	err = pool.QueryRow(context.Background(),
		`SELECT deleted_at FROM activities WHERE id = $1`, created.ID,
	).Scan(&deletedAt)
	require.NoError(t, err)
	require.NotNil(t, deletedAt, "deleted_at should be set after Delete")
}

func TestActivityRepository_Delete_WrongUserID_NoOp(t *testing.T) {
	// Same silent-no-op contract as ActivityImageRepository.Delete: the
	// underlying query is :exec, so a non-matching WHERE (wrong owner
	// here) doesn't surface as an error — the activity must still exist
	// and be unaffected afterwards.
	pool := setupTestDB(t)
	userID := seedTestUser(t, pool)
	otherUserID := seedTestUser(t, pool)
	repo := repository.NewActivityRepository(pool)

	created, err := repo.Create(context.Background(), &entity.Activity{
		UserID: userID, Name: "Not yours", IsFixedSchedule: true,
	})
	require.NoError(t, err)

	err = repo.Delete(context.Background(), created.ID, otherUserID)
	require.NoError(t, err)

	var name string
	var deletedAt *time.Time
	err = pool.QueryRow(context.Background(),
		`SELECT name, deleted_at FROM activities WHERE id = $1`, created.ID,
	).Scan(&name, &deletedAt)
	require.NoError(t, err)
	require.Equal(t, "Not yours", name)
	require.Nil(t, deletedAt, "delete scoped to the wrong user must not soft-delete the row")
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
