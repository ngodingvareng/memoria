package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/ngodingvareng/memoria/internal/entity"
)

// UploadProfileImage implements [UserUsecase].
func (u *userUsecase) UploadProfileImage(ctx context.Context, input UploadProfileImageInput) (*entity.User, error) {
	current, err := u.users.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting current user: %w", err)
	}

	key := buildPublicImageKey("users", input.UserID, input.FileName)
	if err := u.storage.Put(ctx, key, input.Body, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("uploading profile image %s: %w", key, err)
	}

	publicURL := u.storage.PublicURL(key)
	updated, err := u.users.UpdateImagePath(ctx, input.UserID, &publicURL)
	if err != nil {
		// Mirrors thread_image_usecase.go's orphan-cleanup: the upload
		// already succeeded before this DB write failed, so best-effort
		// delete it rather than leave it dangling in storage.
		if delErr := u.storage.Delete(context.WithoutCancel(ctx), key); delErr != nil {
			return nil, fmt.Errorf("saving profile image: %w (cleanup also failed: %v)", err, delErr)
		}
		return nil, fmt.Errorf("saving profile image: %w", err)
	}

	if current.ImagePath != nil {
		if oldKey, ok := strings.CutPrefix(*current.ImagePath, u.storage.PublicURL("")); ok {
			_ = u.storage.Delete(context.WithoutCancel(ctx), oldKey)
		}
	}

	return updated, nil
}
