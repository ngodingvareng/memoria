package usecase

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/entity"
)

type ActivityImageRepository interface {
	Create(ctx context.Context, image *entity.ActivityImage) (*entity.ActivityImage, error)
	ListByActivityID(ctx context.Context, activityID uuid.UUID) ([]*entity.ActivityImage, error)
	Delete(ctx context.Context, activityID, imageID uuid.UUID) error
}

// ActivityAccessChecker is the minimal capability this usecase needs
// from the activity domain: confirming an activity actually belongs to
// the requesting user before letting them touch its images. Deliberately
// separate from the full ActivityRepository interface in
// activity_usecase.go — this usecase has no business with
// Create/WithTransaction. activityRepository satisfies both interfaces
// structurally without any extra wiring.
type ActivityAccessChecker interface {
	GetActivityByID(ctx context.Context, id, userID uuid.UUID) (*entity.Activity, error)
}

type Storage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	PresignGet(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

const presignedURLTTL = 15 * time.Minute

type UploadActivityImageInput struct {
	ActivityID  uuid.UUID
	UserID      uuid.UUID
	FileName    string // client-supplied, used only for its extension — see usecase, never used as the storage key directly
	ContentType string
	Size        int64
	Body        io.Reader
}

// ActivityImageWithURL pairs the stored record with a freshly presigned,
// time-limited download URL — the URL itself is never persisted, it's
// generated on read.
type ActivityImageWithURL struct {
	*entity.ActivityImage
	URL string
}

type ActivityImageUsecase interface {
	UploadActivityImage(ctx context.Context, input UploadActivityImageInput) (*ActivityImageWithURL, error)
	ListActivityImages(ctx context.Context, activityID, userID uuid.UUID) ([]ActivityImageWithURL, error)
	DeleteActivityImage(ctx context.Context, activityID, imageID, userID uuid.UUID) error
}

type activityImageUsecase struct {
	repo       ActivityImageRepository
	storage    Storage
	activities ActivityAccessChecker
}

func NewActivityImageUsecase(repo ActivityImageRepository, storage Storage, activities ActivityAccessChecker) ActivityImageUsecase {
	return &activityImageUsecase{repo: repo, storage: storage, activities: activities}
}

// UploadActivityImage implements [ActivityImageUsecase].
func (u *activityImageUsecase) UploadActivityImage(ctx context.Context, input UploadActivityImageInput) (*ActivityImageWithURL, error) {
	if _, err := u.activities.GetActivityByID(ctx, input.ActivityID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking activity ownership: %w", err)
	}

	key := buildImageKey(input.ActivityID, input.FileName)

	image, err := u.repo.Create(ctx, &entity.ActivityImage{
		ActivityID: input.ActivityID,
		ImagePath:  key,
	})
	if err != nil {
		// The upload to storage already succeeded before this DB insert
		// failed — without this, the object would be orphaned (uploaded,
		// but nothing in the DB ever points to it). Best-effort cleanup:
		// if THIS delete also fails, the object still ends up orphaned,
		// which is exactly the open issue tracked as "orphaned image
		// files" in SCHEMA_REVIEW.md / TODO.md — a periodic reconciliation
		// job (diffing storage against image_path values still
		// referenced in the DB) is the real fix for that, not this.
		if delErr := u.storage.Delete(context.WithoutCancel(ctx), key); delErr != nil {
			return nil, fmt.Errorf("saving image record: %w (cleanup also failed: %v)", err, delErr)
		}
		return nil, fmt.Errorf("saving image record: %w", err)
	}

	url, err := u.storage.PresignGet(ctx, image.ImagePath, presignedURLTTL)
	if err != nil {
		return nil, fmt.Errorf("presigning uploaded image %s: %w", image.ID, err)
	}

	return &ActivityImageWithURL{ActivityImage: image, URL: url}, nil
}

// ListActivityImages implements [ActivityImageUsecase].
func (u *activityImageUsecase) ListActivityImages(ctx context.Context, activityID, userID uuid.UUID) ([]ActivityImageWithURL, error) {
	if _, err := u.activities.GetActivityByID(ctx, activityID, userID); err != nil {
		return nil, fmt.Errorf("checking activity ownership: %w", err)
	}

	images, err := u.repo.ListByActivityID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("listing activity images: %w", err)
	}

	out := make([]ActivityImageWithURL, 0, len(images))
	for _, img := range images {
		url, err := u.storage.PresignGet(ctx, img.ImagePath, presignedURLTTL)
		if err != nil {
			return nil, fmt.Errorf("presigning image %s: %w", img.ID, err)
		}
		out = append(out, ActivityImageWithURL{ActivityImage: img, URL: url})
	}
	return out, nil

}

// DeleteActivityImage implements [ActivityImageUsecase].
func (u *activityImageUsecase) DeleteActivityImage(ctx context.Context, activityID, imageID, userID uuid.UUID) error {
	if _, err := u.activities.GetActivityByID(ctx, activityID, userID); err != nil {
		return fmt.Errorf("checking activity ownership: %w", err)
	}

	// Deliberately delete the DB record first, storage object second.
	// If storage deletion then fails, the result is an orphaned object
	// (same open issue as above) rather than a dangling DB record
	// pointing at nothing — an orphan in storage is a cleanup-job
	// problem; a dangling DB reference is a correctness problem for
	// every other read path.
	if err := u.repo.Delete(ctx, activityID, imageID); err != nil {
		return fmt.Errorf("deleting activity image record: %w", err)
	}
	return nil

}

// buildImageKey deliberately ignores the client-supplied filename except
// for its extension — using it directly would be a path-traversal risk
// (e.g. a filename of "../../etc/passwd").
func buildImageKey(activityID uuid.UUID, clientFileName string) string {
	return fmt.Sprintf("activities/%s/%s%s", activityID, uuid.NewString(), path.Ext(clientFileName))
}
