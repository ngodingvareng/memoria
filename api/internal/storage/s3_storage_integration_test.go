//go:build integration

package storage_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	memoriastorage "github.com/ngodingvareng/memoria/internal/storage"
)

const (
	testBucket    = "memoria-test"
	testAccessKey = "rustfsadmin"
	testSecretKey = "rustfsadmin-test-only"
)

// setupTestStorage spins up a throwaway RustFS container directly via
// GenericContainer. There's no official testcontainers module for RustFS
// yet (unlike Postgres/MinIO), but GenericContainer is the same
// underlying API those official modules are built on — nothing here is
// second-class because of that.
//
// Wait strategy: ForListeningPort rather than an HTTP check against
// RustFS's own /health endpoint. RustFS's health endpoint has some
// reported reliability quirks (HEAD requests failing; readiness
// sometimes not reflecting real state — see rustfs/rustfs#935 and #2658),
// so a plain "is the port accepting connections yet" check is the more
// robust signal for a test that just needs the S3 API to be reachable.
func setupTestStorage(t *testing.T) memoriastorage.Storage {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "rustfs/rustfs:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"RUSTFS_VOLUMES":        "/data/rustfs0",
			"RUSTFS_ADDRESS":        "0.0.0.0:9000",
			"RUSTFS_CONSOLE_ENABLE": "false", // not needed for this test
			"RUSTFS_ACCESS_KEY":     testAccessKey,
			"RUSTFS_SECRET_KEY":     testSecretKey,
		},
		WaitingFor: wait.ForListeningPort("9000/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	host, err := container.Host(ctx)
	require.NoError(t, err)
	mappedPort, err := container.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)
	endpoint := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	createTestBucket(t, ctx, endpoint)

	s, err := memoriastorage.NewS3Storage(ctx, memoriastorage.S3Config{
		Endpoint:     endpoint,
		Region:       "us-east-1",
		AccessKey:    testAccessKey,
		SecretKey:    testSecretKey,
		Bucket:       testBucket,
		UsePathStyle: true,
	})
	require.NoError(t, err)

	return s
}

// createTestBucket uses a raw s3.Client directly — separate from the
// Storage implementation under test — since Storage itself has no
// CreateBucket method (buckets are provisioned out-of-band in real
// deployments, not by application code).
func createTestBucket(t *testing.T, ctx context.Context, endpoint string) {
	t.Helper()

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(testAccessKey, testSecretKey, "")),
	)
	require.NoError(t, err)

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	// RustFS's S3 API can take a brief moment to become fully queryable
	// even after the TCP port is accepting connections — retry a few
	// times instead of failing on the first attempt.
	var lastErr error
	for i := 0; i < 10; i++ {
		_, lastErr = client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(testBucket),
		})
		if lastErr == nil {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, lastErr, "creating test bucket after retries")
}

func TestS3Storage_PutAndPresignGet_RoundTrip(t *testing.T) {
	s := setupTestStorage(t)
	ctx := context.Background()

	key := "activities/test-activity/photo.jpg"
	content := []byte("fake jpeg bytes for testing")

	err := s.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "image/jpeg")
	require.NoError(t, err)

	url, err := s.PresignGet(ctx, key, 5*time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, url)

	// The whole point of a presigned URL is that a plain HTTP GET works
	// without any additional auth — verify that's actually true, not
	// just that PresignGet returned something non-empty.
	resp, err := http.Get(url) //nolint:gosec // test-controlled URL
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, content, body)
}

func TestS3Storage_Delete_ObjectNoLongerDownloadable(t *testing.T) {
	s := setupTestStorage(t)
	ctx := context.Background()

	key := "activities/test-activity/to-delete.jpg"
	err := s.Put(ctx, key, bytes.NewReader([]byte("bytes")), 5, "image/jpeg")
	require.NoError(t, err)

	err = s.Delete(ctx, key)
	require.NoError(t, err)

	url, err := s.PresignGet(ctx, key, 5*time.Minute)
	require.NoError(t, err) // presigning itself doesn't check existence

	resp, err := http.Get(url) //nolint:gosec // test-controlled URL
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestS3Storage_PresignedURL_ExpiresEventually(t *testing.T) {
	s := setupTestStorage(t)
	ctx := context.Background()

	key := "activities/test-activity/expiring.jpg"
	err := s.Put(ctx, key, bytes.NewReader([]byte("bytes")), 5, "image/jpeg")
	require.NoError(t, err)

	// A negative/near-zero TTL should still produce a URL, but one
	// that's already (or almost immediately) expired.
	url, err := s.PresignGet(ctx, key, 1*time.Nanosecond)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(url) //nolint:gosec // test-controlled URL
	require.NoError(t, err)
	defer resp.Body.Close()

	require.NotEqual(t, http.StatusOK, resp.StatusCode, "an expired presigned URL should not grant access")
}
