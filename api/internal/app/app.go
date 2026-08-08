package app

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ngodingvareng/memoria/internal/config"
	"github.com/ngodingvareng/memoria/internal/delivery/rest"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/errs"
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

	authUsecase := usecase.NewAuthUsecase(authUoW, userRepo, userAccountRepo, refreshTokenRepo, hasher, accessTokenIssuer, refreshTokenGenerator, cfg.JWTRefreshTokenTTL, cfg.LoginMaxFailedAttempts, cfg.LoginLockoutDuration)
	authHandler := handler.NewAuthHandler(authUsecase, cfg.SecureCookies)

	// 3b. Threads
	threadRepo := repository.NewThreadRepository(conn)
	threadUsecase := usecase.NewThreadUsecase(threadRepo)
	threadHandler := handler.NewThreadHandler(threadUsecase)

	threadImageRepo := repository.NewThreadImageRepository(conn)
	threadImageUsecase := usecase.NewThreadImageUsecase(threadImageRepo, objectStorage, threadRepo)
	threadImageHandler := handler.NewThreadImageHandler(threadImageUsecase)

	// 3c. Moments
	momentRepo := repository.NewMomentRepository(conn)
	momentUsecase := usecase.NewMomentUsecase(momentRepo, threadRepo)
	momentHandler := handler.NewMomentHandler(momentUsecase)

	momentImageRepo := repository.NewMomentImageRepository(conn)
	momentImageUsecase := usecase.NewMomentImageUsecase(momentImageRepo, objectStorage, momentRepo)
	momentImageHandler := handler.NewMomentImageHandler(momentImageUsecase)

	// 3d. Circle, Mention, Response (Comment/Reaction). userPrivacyRepo
	// has no usecase of its own — it's passed directly wherever a
	// UserBlockChecker/UserKnownChecker/UserMuteChecker is needed.
	userPrivacyRepo := repository.NewUserPrivacyRepository(conn)

	circleRepo := repository.NewCircleRepository(conn)
	circleUsecase := usecase.NewCircleUsecase(circleRepo)
	circleHandler := handler.NewCircleHandler(circleUsecase)

	circleInviteRepo := repository.NewCircleInviteRepository(conn)
	circleInviteUsecase := usecase.NewCircleInviteUsecase(circleInviteRepo, circleRepo, userRepo, userPrivacyRepo, refreshTokenGenerator)
	circleInviteHandler := handler.NewCircleInviteHandler(circleInviteUsecase)

	circleJoinRequestRepo := repository.NewCircleJoinRequestRepository(conn)
	circleJoinRequestUsecase := usecase.NewCircleJoinRequestUsecase(circleJoinRequestRepo, circleRepo, circleInviteRepo, refreshTokenGenerator)
	circleJoinRequestHandler := handler.NewCircleJoinRequestHandler(circleJoinRequestUsecase)

	mentionRepo := repository.NewMentionRepository(conn)
	mentionUsecase := usecase.NewMentionUsecase(mentionRepo, momentRepo, circleRepo, userRepo, userPrivacyRepo, userPrivacyRepo)
	mentionHandler := handler.NewMentionHandler(mentionUsecase)

	responseEventRepo := repository.NewResponseEventRepository(conn)
	responseAccessChecker := usecase.NewResponseAccessChecker(momentRepo, userPrivacyRepo)

	commentRepo := repository.NewCommentRepository(conn)
	commentUsecase := usecase.NewCommentUsecase(commentRepo, responseAccessChecker, responseEventRepo)
	commentHandler := handler.NewCommentHandler(commentUsecase)

	reactionRepo := repository.NewReactionRepository(conn)
	reactionUsecase := usecase.NewReactionUsecase(reactionRepo, responseAccessChecker, responseEventRepo)
	reactionHandler := handler.NewReactionHandler(reactionUsecase)

	// 4. Fiber app + global middleware.
	// Renamed the local var from "app" to "fiberApp" — this file's own
	// package is named "app", and reusing that name for a variable here
	// reads confusingly once there are several fiberApp.Use(...) calls.
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: middleware.CustomErrorHandler,
		// Default fasthttp body limit is 4MB — too small for image
		// uploads. ThreadImageHandler also enforces its own
		// maxImageUploadSize as a second, more specific check.
		BodyLimit: 15 * 1024 * 1024, // 15 MB

	})

	// Access log for every request, structured via slog (see main.go for
	// the JSON handler setup). recover.New() should sit right after it so
	// a panic anywhere downstream still gets logged and turned into a
	// normal 500 instead of crashing the whole process.
	fiberApp.Use(slogfiber.New(slog.Default()))
	fiberApp.Use(recover.New())

	// AllowCredentials is required for the browser to send/receive the
	// httpOnly refresh_token cookie cross-origin (see auth_handler.go's
	// setRefreshCookie) — that in turn means AllowOrigins can't contain
	// "*", only the explicit origins listed in CORS_ALLOWED_ORIGINS.
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSAllowedOrigins,
		AllowCredentials: true,
		AllowHeaders:     []string{fiber.HeaderAuthorization, fiber.HeaderContentType},
	}))

	// Per-IP throttle for the brute-forceable, unauthenticated auth
	// endpoints (login/register) — see router.go for which routes this
	// gets attached to. LimitReached returns errs.ErrTooManyRequests so
	// the response goes through the same CustomErrorHandler JSON shape
	// as every other error.
	authRateLimiter := limiter.New(limiter.Config{
		Max:        cfg.LoginRateLimitMax,
		Expiration: cfg.LoginRateLimitWindow,
		LimitReached: func(c fiber.Ctx) error {
			return errs.ErrTooManyRequests
		},
	})

	// 5. Router
	rest.SetupRoutes(fiberApp, accessTokenIssuer, authRateLimiter, rest.Handlers{
		Auth:              authHandler,
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

	return &Container{
		App: fiberApp,
		DB:  conn,
	}, nil

}
