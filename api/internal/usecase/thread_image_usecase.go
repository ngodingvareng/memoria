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

type ThreadImageRepository interface {
	Create(ctx context.Context, image *entity.ThreadImage) (*entity.ThreadImage, error)
	ListByThreadID(ctx context.Context, threadID uuid.UUID) ([]*entity.ThreadImage, error)
	Delete(ctx context.Context, threadID, imageID uuid.UUID) error
}

// ThreadAccessChecker is the minimal capability this usecase needs
// from the thread domain: confirming an thread actually belongs to
// the requesting user before letting them touch its images. Deliberately
// separate from the full ThreadRepository interface in
// thread_usecase.go — this usecase has no business with
// Create/WithTransaction. The method name is kept identical to
// ThreadRepository.GetByID on purpose: Go satisfies interfaces
// structurally by method name + signature, so threadRepository's single
// GetByID implementation satisfies both interfaces with no extra method
// or wrapper needed.
type ThreadAccessChecker interface {
	GetByID(ctx context.Context, id, userID uuid.UUID) (*entity.Thread, error)
}

type Storage interface {
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	PresignGet(ctx context.Context, key string, expiresIn time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}

const presignedURLTTL = 15 * time.Minute

type UploadThreadImageInput struct {
	ThreadID    uuid.UUID
	UserID      uuid.UUID
	FileName    string // client-supplied, used only for its extension — see usecase, never used as the storage key directly
	ContentType string
	Size        int64
	Body        io.Reader
}

// ThreadImageWithURL pairs the stored record with a freshly presigned,
// time-limited download URL — the URL itself is never persisted, it's
// generated on read.
type ThreadImageWithURL struct {
	*entity.ThreadImage
	URL string
}

type ThreadImageUsecase interface {
	UploadThreadImage(ctx context.Context, input UploadThreadImageInput) (*ThreadImageWithURL, error)
	ListThreadImages(ctx context.Context, threadID, userID uuid.UUID) ([]ThreadImageWithURL, error)
	DeleteThreadImage(ctx context.Context, threadID, imageID, userID uuid.UUID) error
}

type threadImageUsecase struct {
	repo    ThreadImageRepository
	storage Storage
	threads ThreadAccessChecker
}

func NewThreadImageUsecase(repo ThreadImageRepository, storage Storage, threads ThreadAccessChecker) ThreadImageUsecase {
	return &threadImageUsecase{repo: repo, storage: storage, threads: threads}
}

// UploadThreadImage implements [ThreadImageUsecase].
func (u *threadImageUsecase) UploadThreadImage(ctx context.Context, input UploadThreadImageInput) (*ThreadImageWithURL, error) {
	if _, err := u.threads.GetByID(ctx, input.ThreadID, input.UserID); err != nil {
		return nil, fmt.Errorf("checking thread ownership: %w", err)
	}

	key := buildImageKey(input.ThreadID, input.FileName)

	if err := u.storage.Put(ctx, key, input.Body, input.Size, input.ContentType); err != nil {
		return nil, fmt.Errorf("uploading image %s: %w", key, err)
	}

	image, err := u.repo.Create(ctx, &entity.ThreadImage{
		ThreadID:  input.ThreadID,
		ImagePath: key,
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

	return &ThreadImageWithURL{ThreadImage: image, URL: url}, nil
}

// ListThreadImages implements [ThreadImageUsecase].
func (u *threadImageUsecase) ListThreadImages(ctx context.Context, threadID, userID uuid.UUID) ([]ThreadImageWithURL, error) {
	if _, err := u.threads.GetByID(ctx, threadID, userID); err != nil {
		return nil, fmt.Errorf("checking thread ownership: %w", err)
	}

	images, err := u.repo.ListByThreadID(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("listing thread images: %w", err)
	}

	out := make([]ThreadImageWithURL, 0, len(images))
	for _, img := range images {
		url, err := u.storage.PresignGet(ctx, img.ImagePath, presignedURLTTL)
		if err != nil {
			return nil, fmt.Errorf("presigning image %s: %w", img.ID, err)
		}
		out = append(out, ThreadImageWithURL{ThreadImage: img, URL: url})
	}
	return out, nil

}

// DeleteThreadImage implements [ThreadImageUsecase].
func (u *threadImageUsecase) DeleteThreadImage(ctx context.Context, threadID, imageID, userID uuid.UUID) error {
	if _, err := u.threads.GetByID(ctx, threadID, userID); err != nil {
		return fmt.Errorf("checking thread ownership: %w", err)
	}

	// Deliberately delete the DB record first, storage object second.
	// If storage deletion then fails, the result is an orphaned object
	// (same open issue as above) rather than a dangling DB record
	// pointing at nothing — an orphan in storage is a cleanup-job
	// problem; a dangling DB reference is a correctness problem for
	// every other read path.
	if err := u.repo.Delete(ctx, threadID, imageID); err != nil {
		return fmt.Errorf("deleting thread image record: %w", err)
	}
	return nil

}

// buildImageKey deliberately ignores the client-supplied filename except
// for its extension — using it directly would be a path-traversal risk
// (e.g. a filename of "../../etc/passwd").
func buildImageKey(threadID uuid.UUID, clientFileName string) string {
	return fmt.Sprintf("threads/%s/%s%s", threadID, uuid.NewString(), path.Ext(clientFileName))
}
