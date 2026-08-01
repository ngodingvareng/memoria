package app

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/config"
	"github.com/ngodingvareng/memoria/internal/delivery/rest"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/repository"
	"github.com/ngodingvareng/memoria/internal/security"
	"github.com/ngodingvareng/memoria/internal/storage"
	"github.com/ngodingvareng/memoria/internal/usecase"
	slogfiber "github.com/samber/slog-fiber"
)

type Container struct {
	App *fiber.App
	DB  *pgxpool.Pool
}

func NewContainer(cfg *config.Config) (*Container, error) {
	ctx := context.Background()

	// 1. DB
	conn, err := pgxpool.New(ctx, cfg.GetDSN())
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	// 2. Object storage (RustFS in dev; any S3-compatible backend works
	// the same way — only these config values change).
	objectStorage, err := storage.NewS3Storage(ctx, storage.S3Config{
		Endpoint:     cfg.StorageEndpoint,
		Region:       cfg.StorageRegion,
		AccessKey:    cfg.StorageAccessKey,
		SecretKey:    cfg.StorageSecretKey,
		Bucket:       cfg.StorageBucket,
		UsePathStyle: cfg.StorageUsePathStyle,
	})
	if err != nil {
		return nil, err
	}

	// 3. Wiring: repository -> usecase -> handler

	// 3a. Auth
	userRepo := repository.NewUserRepository(conn)
	userAccountRepo := repository.NewUserAccountRepository(conn)
	refreshTokenRepo := repository.NewRefreshTokenRepository(conn)
	authUoW := repository.NewAuthUnitOfWork(conn)

	accessTokenIssuer := security.NewJWTAccessTokenIssuer([]byte(cfg.JWTSecret), cfg.JWTAccessTokenTTL, cfg.JWTIssuer)
	refreshTokenGenerator := security.NewRefreshTokenGenerator()
	hasher := security.NewScryptHasher()

	authUsecase := usecase.NewAuthUsecase(authUoW, userRepo, userAccountRepo, refreshTokenRepo, hasher, accessTokenIssuer, refreshTokenGenerator, cfg.JWTRefreshTokenTTL)
	authHandler := handler.NewAuthHandler(authUsecase, cfg.SecureCookies)

	// 3b. Activities
	activityRepo := repository.NewActivityRepository(conn)
	activityUsecase := usecase.NewActivityUsecase(activityRepo)
	activityHandler := handler.NewActivityHandler(activityUsecase)

	activityImageRepo := repository.NewActivityImageRepository(conn)
	activityImageUsecase := usecase.NewActivityImageUsecase(activityImageRepo, objectStorage)
	activityImageHandler := handler.NewActivityImageHandler(activityImageUsecase)

	// 4. Fiber app + global middleware.
	// Renamed the local var from "app" to "fiberApp" — this file's own
	// package is named "app", and reusing that name for a variable here
	// reads confusingly once there are several fiberApp.Use(...) calls.
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler,
		// Default fasthttp body limit is 4MB — too small for image
		// uploads. ActivityImageHandler also enforces its own
		// maxImageUploadSize as a second, more specific check.
		BodyLimit: 15 * 1024 * 1024, // 15 MB

	})

	// Access log for every request, structured via slog (see main.go for
	// the JSON handler setup). recover.New() should sit right after it so
	// a panic anywhere downstream still gets logged and turned into a
	// normal 500 instead of crashing the whole process.
	fiberApp.Use(slogfiber.New(slog.Default()))
	fiberApp.Use(recover.New())

	// 5. Router
	rest.SetupRoutes(fiberApp, rest.Handlers{
		Auth:          authHandler,
		Activity:      activityHandler,
		ActivityImage: activityImageHandler,
	})

	return &Container{
		App: fiberApp,
		DB:  conn,
	}, nil

}
