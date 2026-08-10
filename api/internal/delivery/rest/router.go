package rest

import (
	"github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	_ "github.com/ngodingvareng/memoria/docs"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/handler"
	"github.com/ngodingvareng/memoria/internal/delivery/rest/middleware"
	"github.com/ngodingvareng/memoria/internal/usecase"
)

// Handlers groups every handler the router needs. Add a field here each
// time a new domain's handler is wired up in app.go.
type Handlers struct {
	Health            *handler.HealthHandler
	Auth              *handler.AuthHandler
	User              *handler.UserHandler
	Thread            *handler.ThreadHandler
	ThreadImage       *handler.ThreadImageHandler
	Moment            *handler.MomentHandler
	MomentImage       *handler.MomentImageHandler
	Circle            *handler.CircleHandler
	CircleInvite      *handler.CircleInviteHandler
	CircleJoinRequest *handler.CircleJoinRequestHandler
	Mention           *handler.MentionHandler
	Comment           *handler.CommentHandler
	Reaction          *handler.ReactionHandler
	Notification      *handler.NotificationHandler
	Search            *handler.SearchHandler
	Album             *handler.AlbumHandler
}

// SetupRoutes wires every route. authRateLimiter is built in app.go
// (where the LOGIN_RATE_LIMIT_* config lives) and applied only to the
// unauthenticated auth endpoints that are actually brute-forceable —
// /refresh and /logout require a valid cookie/session already, so they
// aren't password-guessing targets the way /login and /register are.
func SetupRoutes(app *fiber.App, issuer usecase.AccessTokenIssuer, authRateLimiter fiber.Handler, requireUsername fiber.Handler, h Handlers) {
	app.Get("/healthz", h.Health.Health)

	app.Get("/docs/*", swaggo.New(swaggo.Config{Title: "Book API"}))

	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})

	app.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("./docs/swagger.json")
	})

	auth := app.Group("/auth")
	auth.Post("/register", authRateLimiter, h.Auth.Register)
	auth.Post("/login", authRateLimiter, h.Auth.Login)
	auth.Post("/google", authRateLimiter, h.Auth.GoogleLogin)
	auth.Post("/refresh", h.Auth.Refresh)
	auth.Post("/logout", h.Auth.Logout)
	// Same rate-limited group as register/login: both unauthenticated,
	// guessable-input endpoints.
	auth.Post("/forgot-password", authRateLimiter, h.Auth.ForgotPassword)
	auth.Post("/reset-password", authRateLimiter, h.Auth.ResetPassword)

	users := app.Group("/users", middleware.RequireAuth(issuer))
	// These two are exempt from requireUsername on purpose — they're the
	// only endpoints a not-yet-onboarded account is allowed to call, since
	// they're how it finishes onboarding in the first place.
	users.Get("/username-availability", h.User.CheckUsernameAvailability)
	users.Patch("/me/username", h.User.SetUsername)
	users.Get("/me", requireUsername, h.User.GetMe)
	users.Put("/me/privacy", requireUsername, h.User.UpdatePrivacySettings)
	users.Get("/me/blocks", requireUsername, h.User.ListBlockedUsers)
	users.Post("/me/blocks", requireUsername, h.User.BlockUser)
	users.Delete("/me/blocks/:username", requireUsername, h.User.UnblockUser)
	users.Get("/me/mutes", requireUsername, h.User.ListMutedUsers)
	users.Post("/me/mutes", requireUsername, h.User.MuteUser)
	users.Delete("/me/mutes/:username", requireUsername, h.User.UnmuteUser)
	users.Get("/:id", requireUsername, h.User.GetUserByID)
	users.Get("/username/:username", requireUsername, h.User.GetUserByUsername)
	users.Post("/username/:username/known", requireUsername, h.User.MarkUserKnown)
	users.Post("/me/image", requireUsername, h.User.UploadProfileImage)

	threads := app.Group("/threads", middleware.RequireAuth(issuer), requireUsername)
	threads.Post("/", h.Thread.CreateThread)
	threads.Put("/:id", h.Thread.UpdateThread)
	threads.Delete("/:id", h.Thread.DeleteThread)
	threads.Get("/", h.Thread.SearchThreads)
	threads.Get("/:id", h.Thread.GetThread)

	threads.Post("/:id/images", h.ThreadImage.UploadThreadImage)
	threads.Get("/:id/images", h.ThreadImage.ListThreadImages)
	threads.Delete("/:id/images/:imageId", h.ThreadImage.DeleteThreadImage)

	threads.Get("/:id/moments", h.Moment.ListThreadMoments)

	moments := app.Group("/moments", middleware.RequireAuth(issuer), requireUsername)
	moments.Post("/", h.Moment.CreateMoment)
	moments.Get("/", h.Moment.ListMoments)
	moments.Get("/search", h.Moment.SearchMoments)
	moments.Put("/:id", h.Moment.UpdateMoment)
	moments.Delete("/:id", h.Moment.DeleteMoment)
	moments.Get("/:id", h.Moment.GetMoment)

	moments.Post("/:id/images", h.MomentImage.UploadMomentImage)
	moments.Get("/:id/images", h.MomentImage.ListMomentImages)
	moments.Delete("/:id/images/:imageId", h.MomentImage.DeleteMomentImage)

	moments.Get("/:id/mentions", h.Mention.ListMentions)
	moments.Post("/:id/mentions", h.Mention.CreateMention)
	moments.Delete("/:id/mentions/:mentionId", h.Mention.DeleteMention)
	moments.Post("/:id/mentions/leave", h.Mention.LeaveMention)
	moments.Post("/:id/share", h.Mention.ShareToCircle)
	moments.Delete("/:id/share/:circleId", h.Mention.UnshareFromCircle)
	moments.Get("/:id/shares", h.Mention.ListShares)

	moments.Get("/:id/audience", h.Comment.GetAudience)
	moments.Get("/:id/comments", h.Comment.ListComments)
	moments.Post("/:id/comments", h.Comment.CreateComment)
	moments.Put("/:id/comments/:commentId", h.Comment.UpdateComment)
	moments.Delete("/:id/comments/:commentId", h.Comment.DeleteComment)

	moments.Get("/:id/reactions", h.Reaction.ListReactions)
	moments.Put("/:id/reactions", h.Reaction.SetReaction)
	moments.Delete("/:id/reactions", h.Reaction.RemoveReaction)

	mentioned := app.Group("/mentions", middleware.RequireAuth(issuer), requireUsername)
	mentioned.Get("/", h.Mention.ListMentionedMoments)
	mentioned.Get("/users", h.Mention.SearchMentionableUsers)

	circles := app.Group("/circles", middleware.RequireAuth(issuer), requireUsername)
	circles.Post("/", h.Circle.CreateCircle)
	circles.Get("/", h.Circle.ListMyCircles)
	circles.Get("/:id", h.Circle.GetCircle)
	circles.Put("/:id", h.Circle.UpdateCircle)
	circles.Post("/:id/image", h.Circle.UploadCircleImage)
	circles.Delete("/:id", h.Circle.DissolveCircle)
	circles.Get("/:id/members", h.Circle.ListMembers)
	circles.Delete("/:id/members/:userId", h.Circle.RemoveMember)
	circles.Patch("/:id/members/:userId", h.Circle.UpdateMemberPermissions)
	circles.Patch("/:id/members/:userId/role", h.Circle.UpdateMemberRole)

	circles.Post("/:id/members/direct", h.CircleInvite.InviteDirect)
	circles.Get("/:id/invite-link", h.CircleInvite.GetInviteLink)
	circles.Post("/:id/invite-link", h.CircleInvite.CreateOrRotateInviteLink)
	circles.Patch("/:id/invite-link/approval", h.CircleInvite.SetInviteLinkRequiresApproval)
	circles.Post("/:id/invites/:inviteId/revoke", h.CircleInvite.RevokeInvite)

	circles.Get("/:id/threads", h.Thread.ListCircleThreads)
	circles.Get("/:id/moments", h.Moment.ListCircleMoments)
	circles.Get("/:id/album", h.Album.ListCircleAlbum)

	circles.Get("/:id/join-requests", h.CircleJoinRequest.ListPending)
	circles.Post("/:id/join-requests/:requestId/approve", h.CircleJoinRequest.Approve)
	circles.Post("/:id/join-requests/:requestId/reject", h.CircleJoinRequest.Reject)

	circleInvites := app.Group("/circle-invites", middleware.RequireAuth(issuer), requireUsername)
	circleInvites.Get("/", h.CircleInvite.ListMyPendingInvites)
	circleInvites.Post("/:id/accept", h.CircleInvite.AcceptInvite)
	circleInvites.Post("/:id/decline", h.CircleInvite.DeclineInvite)

	circleJoinRequests := app.Group("/circle-join-requests", middleware.RequireAuth(issuer), requireUsername)
	circleJoinRequests.Post("/", h.CircleJoinRequest.FollowInviteLink)
	circleJoinRequests.Post("/:id/cancel", h.CircleJoinRequest.Cancel)

	notifications := app.Group("/notifications", middleware.RequireAuth(issuer), requireUsername)
	notifications.Get("/", h.Notification.ListNotifications)
	notifications.Get("/preferences", h.Notification.GetNotificationPreferences)
	notifications.Put("/preferences", h.Notification.UpdateNotificationPreferences)
	notifications.Post("/read-all", h.Notification.MarkAllNotificationsRead)
	notifications.Patch("/:id/read", h.Notification.MarkNotificationRead)

	search := app.Group("/search", middleware.RequireAuth(issuer), requireUsername)
	search.Get("/suggestions", h.Search.Suggestions)

	album := app.Group("/album", middleware.RequireAuth(issuer), requireUsername)
	album.Get("/", h.Album.ListAlbum)
}
