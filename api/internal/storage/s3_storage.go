package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
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
	// PublicPrefixes are key prefixes (e.g. "users/", "circles/") that
	// should be readable by anyone with the URL, no signing required —
	// profile photos, not private archive content. Applied once, here,
	// as a bucket policy: per-object ACLs (the more common approach)
	// turned out not to be honored by RustFS when this was verified
	// against the dev instance, so this is a deliberate, tested choice,
	// not the default pattern.
	PublicPrefixes []string
}

type s3Storage struct {
	client       *s3.Client
	bucket       string
	endpoint     string
	usePathStyle bool
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

	s := &s3Storage{
		client:       client,
		bucket:       cfg.Bucket,
		endpoint:     cfg.Endpoint,
		usePathStyle: cfg.UsePathStyle,
	}

	if len(cfg.PublicPrefixes) > 0 {
		if err := s.applyPublicPrefixPolicy(ctx, cfg.PublicPrefixes); err != nil {
			// Not fatal to startup: worst case, newly uploaded profile
			// photos 403 until this is fixed, which is loud and quick to
			// notice — better than the app refusing to boot over an
			// object-storage policy call.
			slog.WarnContext(
				ctx,
				"failed to apply public-read bucket policy",
				"error",
				err,
				"prefixes",
				cfg.PublicPrefixes,
			)
		}
	}

	return s, nil
}

// applyPublicPrefixPolicy grants anonymous s3:GetObject on the given key
// prefixes, leaving everything else in the bucket private. Idempotent —
// safe to call on every startup.
func (s *s3Storage) applyPublicPrefixPolicy(ctx context.Context, prefixes []string) error {
	resources := make([]string, len(prefixes))
	for i, prefix := range prefixes {
		resources[i] = fmt.Sprintf("arn:aws:s3:::%s/%s*", s.bucket, prefix)
	}

	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Sid":       "PublicReadProfileImages",
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    "s3:GetObject",
				"Resource":  resources,
			},
		},
	}

	body, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("marshaling bucket policy: %w", err)
	}

	if _, err := s.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(s.bucket),
		Policy: aws.String(string(body)),
	}); err != nil {
		return fmt.Errorf("put bucket policy: %w", err)
	}
	return nil
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

// PublicURL implements [Storage]. Pure string formatting — matches
// whatever style (path vs. virtual-hosted) the client was configured
// with, since that's what determines which form actually resolves.
// Only resolves for keys under one of S3Config.PublicPrefixes.
func (s *s3Storage) PublicURL(key string) string {
	if s.usePathStyle {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(s.endpoint, "/"), s.bucket, key)
	}

	u, err := url.Parse(s.endpoint)
	if err != nil {
		return fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(s.endpoint, "/"), s.bucket, key)
	}
	u.Host = s.bucket + "." + u.Host
	u.Path = "/" + key
	return u.String()
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
