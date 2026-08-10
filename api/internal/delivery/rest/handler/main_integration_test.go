//go:build integration

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/go-openapi/testify/v2/require"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ngodingvareng/memoria/internal/delivery/rest"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/security"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// testPool is shared by every handler integration test in this package,
// same reasoning as internal/repository/main_integration_test.go: one
// Postgres container per test binary run, not one per test.
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

	migrator, err := migrate.New("file://../../../../db/migrations", dsn)
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

// setupTestDB truncates every table before each test so they stay
// isolated despite sharing one container.
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

func testUsername() string {
	return "user_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:20]
}

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

// fakeStorage satisfies usecase.Storage without touching any real
// object storage backend. The real S3-backed implementation
// (internal/storage/s3_storage.go) already has its own integration
// test against a real RustFS container
// (internal/storage/s3_storage_integration_test.go) — handler tests
// only need to verify routing/auth/DTO/usecase wiring, not the S3 SDK,
// so re-exercising a real backend here would just add container
// startup cost without covering anything new.
type fakeStorage struct{}

func (fakeStorage) Put(_ context.Context, _ string, body io.Reader, _ int64, _ string) error {
	_, err := io.Copy(io.Discard, body)
	return err
}

func (fakeStorage) PresignGet(_ context.Context, key string, _ time.Duration) (string, error) {
	return "https://fake.storage.test/" + key, nil
}

func (fakeStorage) PublicURL(key string) string {
	return "https://fake.storage.test/" + key
}

func (fakeStorage) Delete(_ context.Context, _ string) error {
	return nil
}

// fakeMailer satisfies usecase.Mailer without sending any real email —
// handler tests only need to verify routing/auth/DTO/usecase wiring, not
// SMTP delivery.
type fakeMailer struct{}

func (fakeMailer) Send(_ context.Context, _, _, _ string) error {
	return nil
}

// testApp is the full route tree wired against the shared test
// Postgres, mirroring internal/app.NewContainer's wiring exactly except
// for object storage (fakeStorage above) and config (hardcoded test
// values instead of .env/viper, since GetDSN's "localhost" assumption
// doesn't match the testcontainers-assigned port anyway).
type testApp struct {
	app    *fiber.App
	pool   *pgxpool.Pool
	issuer usecase.AccessTokenIssuer
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()
	pool := setupTestDB(t)

	accessTokenIssuer := security.NewJWTAccessTokenIssuer([]byte("test-secret"), time.Hour, "memoria-test")
	refreshTokenGenerator := security.NewRefreshTokenGenerator()
	hasher := security.NewScryptHasher()
	storage := fakeStorage{}
	mailer := fakeMailer{}

	userRepo := repository.NewUserRepository(pool)
	userAccountRepo := repository.NewUserAccountRepository(pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(pool)
	userVerificationRepo := repository.NewUserVerificationRepository(pool)
	authUoW := repository.NewAuthUnitOfWork(pool)
	authUsecase := usecase.NewAuthUsecase(authUoW, userRepo, userAccountRepo, refreshTokenRepo, userVerificationRepo, hasher,
		accessTokenIssuer, refreshTokenGenerator, mailer, "https://memoria-test.example", 30*24*time.Hour, 5, 15*time.Minute)
	authHandler := handler.NewAuthHandler(authUsecase, false)

	userPrivacyRepo := repository.NewUserPrivacyRepository(pool)

	userUsecase := usecase.NewUserUsecase(userRepo, userPrivacyRepo, storage)
	userHandler := handler.NewUserHandler(userUsecase)

	circleRepo := repository.NewCircleRepository(pool)

	threadRepo := repository.NewThreadRepository(pool)
	threadUsecase := usecase.NewThreadUsecase(threadRepo, circleRepo)
	threadHandler := handler.NewThreadHandler(threadUsecase)

	threadImageRepo := repository.NewThreadImageRepository(pool)
	threadImageUsecase := usecase.NewThreadImageUsecase(threadImageRepo, storage, threadRepo)
	threadImageHandler := handler.NewThreadImageHandler(threadImageUsecase)

	momentRepo := repository.NewMomentRepository(pool)
	momentUsecase := usecase.NewMomentUsecase(momentRepo, threadRepo, circleRepo)
	momentHandler := handler.NewMomentHandler(momentUsecase)

	momentImageRepo := repository.NewMomentImageRepository(pool)
	momentImageUsecase := usecase.NewMomentImageUsecase(momentImageRepo, storage, momentRepo)
	momentImageHandler := handler.NewMomentImageHandler(momentImageUsecase)

	circleUsecase := usecase.NewCircleUsecase(circleRepo, storage)
	circleHandler := handler.NewCircleHandler(circleUsecase)

	circleInviteRepo := repository.NewCircleInviteRepository(pool)
	circleInviteUsecase := usecase.NewCircleInviteUsecase(circleInviteRepo, circleRepo, userRepo, userPrivacyRepo, refreshTokenGenerator)
	circleInviteHandler := handler.NewCircleInviteHandler(circleInviteUsecase)

	circleJoinRequestRepo := repository.NewCircleJoinRequestRepository(pool)
	circleJoinRequestUsecase := usecase.NewCircleJoinRequestUsecase(circleJoinRequestRepo, circleRepo, circleInviteRepo, refreshTokenGenerator)
	circleJoinRequestHandler := handler.NewCircleJoinRequestHandler(circleJoinRequestUsecase)

	mentionRepo := repository.NewMentionRepository(pool)
	mentionUsecase := usecase.NewMentionUsecase(mentionRepo, momentRepo, circleRepo, circleRepo, userRepo, userPrivacyRepo, userPrivacyRepo, userRepo)
	mentionHandler := handler.NewMentionHandler(mentionUsecase)

	responseEventRepo := repository.NewResponseEventRepository(pool)
	responseAccessChecker := usecase.NewResponseAccessChecker(momentRepo, userPrivacyRepo)

	commentRepo := repository.NewCommentRepository(pool)
	commentUsecase := usecase.NewCommentUsecase(commentRepo, responseAccessChecker, responseEventRepo)
	commentHandler := handler.NewCommentHandler(commentUsecase)

	reactionRepo := repository.NewReactionRepository(pool)
	reactionUsecase := usecase.NewReactionUsecase(reactionRepo, responseAccessChecker, responseEventRepo)
	reactionHandler := handler.NewReactionHandler(reactionUsecase)

	fiberApp := fiber.New(fiber.Config{ErrorHandler: middleware.CustomErrorHandler})
	fiberApp.Use(recover.New())

	// No rate limiting in tests — CORS/access-log/limiter are
	// transport-level concerns wired in internal/app.NewContainer, not
	// route-level behavior SetupRoutes itself is responsible for.
	noOpRateLimiter := func(c fiber.Ctx) error { return c.Next() }

	rest.SetupRoutes(fiberApp, accessTokenIssuer, noOpRateLimiter, rest.Handlers{
		Auth:              authHandler,
		User:              userHandler,
		Thread:            threadHandler,
		ThreadImage:       threadImageHandler,
		Moment:            momentHandler,
		MomentImage:       momentImageHandler,
		Circle:            circleHandler,
		CircleInvite:      circleInviteHandler,
		CircleJoinRequest: circleJoinRequestHandler,
		Mention:           mentionHandler,
		Comment:           commentHandler,
		Reaction:          reactionHandler,
	})

	return &testApp{app: fiberApp, pool: pool, issuer: accessTokenIssuer}
}

// authHeader mints a real access token for userID, bypassing
// POST /auth/login — tests that aren't about the auth flow itself just
// need a valid Bearer token, the same way they'd seed a user directly
// via SQL instead of going through POST /auth/register.
func (a *testApp) authHeader(t *testing.T, userID uuid.UUID) string {
	t.Helper()
	token, _, err := a.issuer.Generate(userID)
	require.NoError(t, err)
	return "Bearer " + token
}

// doRequest sends body (marshaled to JSON if non-nil) to method+path
// through the full route tree via Fiber's app.Test, which exercises
// routing/middleware/handlers without binding a real port.
func (a *testApp) doRequest(t *testing.T, method, path string, body any, headers map[string]string) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(raw)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := a.app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second, FailOnTimeout: true})
	require.NoError(t, err)
	return resp
}

// decodeBody reads and JSON-decodes resp.Body into a value of type T,
// closing the body afterwards.
func decodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()

	var out T
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}
