//go:build integration

package repository_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testPool is shared by every integration test in this package. A single
// Postgres container backs the whole test binary run instead of one per
// test function — spinning up 50+ containers sequentially blew past go
// test's default 10m timeout. setupTestDB truncates all tables before each
// test to keep them isolated from one another.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
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
	if err != nil {
		fmt.Fprintln(os.Stderr, "setup test postgres container:", err)
		os.Exit(1)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "resolve test postgres dsn:", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	migrator, err := migrate.New("file://../../db/migrations", dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load migrations:", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}
	if err := migrator.Up(); err != nil {
		fmt.Fprintln(os.Stderr, "run migrations:", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect test pool:", err)
		_ = pgContainer.Terminate(ctx)
		os.Exit(1)
	}
	testPool = pool

	code := m.Run()

	pool.Close()
	_ = pgContainer.Terminate(ctx)

	os.Exit(code)
}

// setupTestDB returns the package-wide pool set up by TestMain above,
// truncating every table first so each test starts from a clean, isolated
// database.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	rows, err := testPool.Query(ctx,
		`SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations'`)
	require.NoError(t, err)

	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	rows.Close()
	require.NoError(t, rows.Err())

	if len(tables) > 0 {
		query := `TRUNCATE TABLE ` + strings.Join(tables, ", ") + ` RESTART IDENTITY CASCADE`
		_, err = testPool.Exec(ctx, query)
		require.NoError(t, err)
	}

	return testPool
}

// testUsername generates a unique value matching
// chk_users_username_format (^[a-z0-9_.]{3,30}$) — uuid.NewString()'s
// dashes aren't allowed there, so they're stripped.
func testUsername() string {
	return "user_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:20]
}

// strPtr is a small literal-to-pointer helper for entity.User.Username
// (nil until claimed via UserRepository.SetUsername).
func strPtr(s string) *string { return &s }

func seedTestUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (name, username, email) VALUES ($1, $2, $3) RETURNING id`,
		"Test User", testUsername(), uuid.NewString()+"@example.com",
	).Scan(&id)
	require.NoError(t, err)
	return id
}
