package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/ngodingvareng/memoria/internal/entity"
)

// UploadCircleImage implements [CircleUsecase]. Admin-only, enforced by
// CircleRepository.UpdateImagePath's own WHERE clause — a non-admin's
// attempt surfaces as errs.ErrNotFound, same as UpdateCircle.
func (u *circleUsecase) UploadCircleImage(ctx context.Context, input UploadCircleImageInput) (*entity.Circle, error) {
	current, err := u.repo.GetByID(ctx, input.CircleID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("getting current circle: %w", err)
	}

	key := buildPublicImageKey("circles", input.CircleID, input.FileName)
	if err := u.storage.Put(ctx, key, input.Body, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("uploading circle image %s: %w", key, err)
	}

	publicURL := u.storage.PublicURL(key)
	updated, err := u.repo.UpdateImagePath(ctx, input.CircleID, input.UserID, &publicURL)
	if err != nil {
		if delErr := u.storage.Delete(context.WithoutCancel(ctx), key); delErr != nil {
			return nil, fmt.Errorf("saving circle image: %w (cleanup also failed: %v)", err, delErr)
		}
		return nil, fmt.Errorf("saving circle image: %w", err)
	}

	if current.ImagePath != nil {
		if oldKey, ok := strings.CutPrefix(*current.ImagePath, u.storage.PublicURL("")); ok {
			_ = u.storage.Delete(context.WithoutCancel(ctx), oldKey)
		}
	}

	return updated, nil
}
