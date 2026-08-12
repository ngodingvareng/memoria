package usecase

import (
	"context"
	"fmt"

	"github.com/ngodingvareng/memoria/internal/entity"
)

// UpdatePrivacySettings implements [UserUsecase].
func (u *userUsecase) UpdatePrivacySettings(
	ctx context.Context,
	input UpdatePrivacySettingsInput,
) (*entity.User, error) {
	updated, err := u.users.UpdatePrivacySettings(ctx, &entity.User{
		ID:                     input.UserID,
		MentionPolicy:          input.MentionPolicy,
		CircleInvitePolicy:     input.CircleInvitePolicy,
		DiscoverableByUsername: input.DiscoverableByUsername,
		StripPhotoMetadata:     input.StripPhotoMetadata,
	})
	if err != nil {
		return nil, fmt.Errorf("updating privacy settings: %w", err)
	}
	return updated, nil
}
