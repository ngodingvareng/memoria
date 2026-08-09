package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	Endpoint     string
	Region       string
	AccessKey    string
	SecretKey    string
	Bucket       string
	UsePathStyle bool
}

type s3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(ctx context.Context, cfg S3Config) (*s3Storage, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey, cfg.SecretKey, "",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("loading aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
		// PutObject bodies here are unseekable multipart upload streams.
		// The SDK's default (WhenSupported) tries to compute a trailing
		// checksum for every request, which it refuses to do for an
		// unseekable stream over plain HTTP (no TLS) — exactly RustFS's
		// local dev setup. WhenRequired restores the pre-v1.21 behavior:
		// only compute a checksum when the operation demands one.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})

	return &s3Storage{client: client, bucket: cfg.Bucket}, nil

}

func (s *s3Storage) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	// SigV4 signing needs to seek the body to compute the payload hash.
	// Callers here (thread/moment image uploads) pass an io.MultiReader
	// chained on top of the multipart file for content-type sniffing,
	// which isn't an io.Seeker even though the underlying file was —
	// buffer it once instead. size is already capped by the caller
	// (10MB), so this is bounded.
	seekable, ok := body.(io.ReadSeeker)
	if !ok {
		buf, err := io.ReadAll(body)
		if err != nil {
			return fmt.Errorf("buffering upload body for %q: %w", key, err)
		}
		seekable = bytes.NewReader(buf)
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          seekable,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

func (s *s3Storage) PresignGet(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiresIn))
	if err != nil {
		return "", fmt.Errorf("presign object %q: %w", key, err)
	}
	return req.URL, nil
}

func (s *s3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", key, err)
	}
	return nil
}
