package usecase

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

type MomentImageRepository interface {
	Create(ctx context.Context, image *entity.MomentImage) (*entity.MomentImage, error)
	ListByMomentID(ctx context.Context, momentID uuid.UUID) ([]*entity.MomentImage, error)
	// Delete returns the deleted image (nil if id/momentID didn't match
	// an existing row) so the caller can remove the matching storage
	// object.
	Delete(ctx context.Context, momentID, imageID uuid.UUID) (*entity.MomentImage, error)
	// ListImagePathsByOwnerID is account deletion's storage-cleanup
	// gather step — every image path belonging to a Moment ownerUserID
	// owns, resolved before those Moments are soft-deleted (see
	// UserUsecase.DeleteAccount).
	ListImagePathsByOwnerID(ctx context.Context, ownerUserID uuid.UUID) ([]string, error)
}

type UploadMomentImageInput struct {
	MomentID    uuid.UUID
	UserID      uuid.UUID
	FileName    string // client-supplied, used only for its extension — see buildMomentImageKey
	ContentType string
	Size        int64
	Body        io.Reader
}

// MomentImageWithURL pairs the stored record with a freshly presigned,
// time-limited download URL — the URL itself is never persisted, it's
// generated on read. Mirrors ThreadImageWithURL.
type MomentImageWithURL struct {
	*entity.MomentImage

	URL string
}

type MomentImageUsecase interface {
	UploadMomentImage(ctx context.Context, input UploadMomentImageInput) (*MomentImageWithURL, error)
	ListMomentImages(ctx context.Context, momentID, userID uuid.UUID) ([]MomentImageWithURL, error)
	DeleteMomentImage(ctx context.Context, momentID, imageID, userID uuid.UUID) error
}

type momentImageUsecase struct {
	repo    MomentImageRepository
	storage Storage
	moments MomentAccessChecker
}

func NewMomentImageUsecase(
	repo MomentImageRepository,
	storage Storage,
	moments MomentAccessChecker,
) MomentImageUsecase {
	return &momentImageUsecase{repo: repo, storage: storage, moments: moments}
}

// UploadMomentImage implements [MomentImageUsecase]. Photos are what
// feed the Album (FEATURES.md, Looking Back), so every image here is
// tied to a Moment the caller actually owns.
func (u *momentImageUsecase) UploadMomentImage(
	ctx context.Context,
	input UploadMomentImageInput,
) (*MomentImageWithURL, error) {
	if _, err := u.moments.GetByID(ctx, input.MomentID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}

	key := buildMomentImageKey(input.MomentID, input.FileName)

	if err := u.storage.Put(ctx, key, input.Body, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("uploading image %s: %w", key, err)
	}

	contentType := input.ContentType
	size := input.Size
	image, err := u.repo.Create(ctx, &entity.MomentImage{
		MomentID:    input.MomentID,
		ImagePath:   key,
		ContentType: &contentType,
		ByteSize:    &size,
	})
	if err != nil {
		// Same orphan-on-failure tradeoff as ThreadImageUsecase.UploadThreadImage
		// — see its comment and SCHEMA_REVIEW.md / TODO.md for the real fix.
		if delErr := u.storage.Delete(context.WithoutCancel(ctx), key); delErr != nil {
			return nil, fmt.Errorf("saving image record: %w (cleanup also failed: %w)", err, delErr)
		}
		return nil, fmt.Errorf("saving image record: %w", err)
	}

	url, err := u.storage.PresignGet(ctx, image.ImagePath, presignedURLTTL)
	if err != nil {
		return nil, fmt.Errorf("presigning uploaded image %s: %w", image.ID, err)
	}

	return &MomentImageWithURL{MomentImage: image, URL: url}, nil
}

// ListMomentImages implements [MomentImageUsecase].
func (u *momentImageUsecase) ListMomentImages(
	ctx context.Context,
	momentID, userID uuid.UUID,
) ([]MomentImageWithURL, error) {
	if _, err := u.moments.GetByID(ctx, momentID, userID); err != nil {
		return nil, fmt.Errorf("checking moment ownership: %w", err)
	}

	images, err := u.repo.ListByMomentID(ctx, momentID)
	if err != nil {
		return nil, fmt.Errorf("listing moment images: %w", err)
	}

	out := make([]MomentImageWithURL, 0, len(images))
	for _, img := range images {
		url, err := u.storage.PresignGet(ctx, img.ImagePath, presignedURLTTL)
		if err != nil {
			return nil, fmt.Errorf("presigning image %s: %w", img.ID, err)
		}
		out = append(out, MomentImageWithURL{MomentImage: img, URL: url})
	}
	return out, nil
}

// DeleteMomentImage implements [MomentImageUsecase].
func (u *momentImageUsecase) DeleteMomentImage(ctx context.Context, momentID, imageID, userID uuid.UUID) error {
	if _, err := u.moments.GetByID(ctx, momentID, userID); err != nil {
		return fmt.Errorf("checking moment ownership: %w", err)
	}

	// DB record deleted before the storage object, same ordering
	// rationale as ThreadImageUsecase.DeleteThreadImage.
	deleted, err := u.repo.Delete(ctx, momentID, imageID)
	if err != nil {
		return fmt.Errorf("deleting moment image record: %w", err)
	}
	if deleted == nil {
		return nil
	}
	if err := u.storage.Delete(ctx, deleted.ImagePath); err != nil {
		return fmt.Errorf("deleting moment image object: %w", err)
	}
	return nil
}

// buildMomentImageKey deliberately ignores the client-supplied filename
// except for its extension — using it directly would be a path-traversal
// risk (e.g. a filename of "../../etc/passwd").
func buildMomentImageKey(momentID uuid.UUID, clientFileName string) string {
	return fmt.Sprintf("moments/%s/%s%s", momentID, uuid.NewString(), path.Ext(clientFileName))
}
