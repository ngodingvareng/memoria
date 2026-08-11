package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/ngodingvareng/memoria/internal/entity"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// CheckUsernameAvailability implements [UserUsecase].
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

// SetUsername implements [UserUsecase].
func (u *userUsecase) SetUsername(ctx context.Context, userID uuid.UUID, username string) (*entity.User, error) {
	user, err := u.users.SetUsername(ctx, userID, username)
	if err != nil {
		return nil, fmt.Errorf("setting username: %w", err)
	}
	return user, nil
}
