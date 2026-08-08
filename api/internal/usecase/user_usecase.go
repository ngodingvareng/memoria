package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// UserUsecase backs the post-register onboarding step, where a newly
// created account (no username yet) claims one.
type UserUsecase interface {
	CheckUsernameAvailability(ctx context.Context, username string) (bool, error)
	// SetUsername returns errs.ErrUsernameAlreadyExists if the username
	// was taken between the last availability check and this call.
	SetUsername(ctx context.Context, userID uuid.UUID, username string) (*entity.User, error)
}

type userUsecase struct {
	users UserRepository
}

func NewUserUsecase(users UserRepository) UserUsecase {
	return &userUsecase{users: users}
}

func (u *userUsecase) CheckUsernameAvailability(ctx context.Context, username string) (bool, error) {
	_, err := u.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return true, nil
		}
		return false, fmt.Errorf("checking username availability: %w", err)
	}
	return false, nil
}

func (u *userUsecase) SetUsername(ctx context.Context, userID uuid.UUID, username string) (*entity.User, error) {
	user, err := u.users.SetUsername(ctx, userID, username)
	if err != nil {
		return nil, fmt.Errorf("setting username: %w", err)
	}
	return user, nil
}
