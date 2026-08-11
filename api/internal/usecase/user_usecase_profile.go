package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// GetPublicProfile implements [UserUsecase].
func (u *userUsecase) GetPublicProfile(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("getting public profile: %w", err)
	}
	return user, nil
}

// GetPublicProfileByUsername implements [UserUsecase].
func (u *userUsecase) GetPublicProfileByUsername(ctx context.Context, username string) (*entity.User, error) {
	user, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("getting public profile by username: %w", err)
	}
	return user, nil
}

// GetOwnProfile implements [UserUsecase].
func (u *userUsecase) GetOwnProfile(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := u.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting own profile: %w", err)
	}
	return user, nil
}
