package usecase

import "context"

// AccountDeletionRepositories bundles every repository
// UserUsecase.DeleteAccount needs, all bound to the same transaction —
// the multi-repository counterpart to the single-repository
// WithTransaction pattern used by MomentRepository/ThreadRepository/
// CircleRepository themselves (see CODING_STANDARDS.md; mirrors
// AuthRepositories/AuthUnitOfWork in auth_usecase.go, which is the same
// shape for a smaller set of repositories).
type AccountDeletionRepositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	Moment       MomentRepository
	MomentImage  MomentImageRepository
	Thread       ThreadRepository
	ThreadImage  ThreadImageRepository
	Comment      CommentRepository
	Reaction     ReactionRepository
	Circle       CircleRepository
}

// AccountDeletionUnitOfWork runs fn with repositories all bound to the
// SAME transaction. Account deletion touches users, moments, threads,
// their images, comments, reactions, refresh_tokens, and circle_members
// — no single existing repository's own WithTransaction can coordinate
// across all of them.
type AccountDeletionUnitOfWork interface {
	WithTransaction(ctx context.Context, fn func(AccountDeletionRepositories) error) error
}
