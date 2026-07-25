package storage

import (
	"context"
	"io"
	"time"
)

// Storage abstracts the object storage backend. It's implemented by
// s3Storage (s3_storage.go), which works against RustFS today but is
// equally valid against MinIO, AWS S3, Cloudflare R2, Backblaze B2, or
// anything else speaking the S3 API — only the endpoint/credentials
// change, this interface and its implementation don't.
type Storage interface {
	// Put uploads an object under key. key is what gets persisted as
	// activity_images.image_path — store the key, not a full URL, so
	// switching backends later never requires a data migration.
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error

	// PresignGet returns a temporary, signed URL the client can download
	// the object from directly. The app's own server never streams file
	// bytes itself.
	PresignGet(ctx context.Context, key string, expiresIn time.Duration) (string, error)

	Delete(ctx context.Context, key string) error
}
