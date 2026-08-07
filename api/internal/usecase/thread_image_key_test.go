//go:build unit

package usecase

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBuildImageKey_IgnoresPathTraversalInFilename(t *testing.T) {
	threadID := uuid.New()

	tests := []struct {
		name           string
		clientFileName string
		wantSuffix     string
	}{
		{"normal filename", "cat.jpg", ".jpg"},
		{"path traversal attempt", "../../etc/passwd.jpg", ".jpg"},
		{"absolute path, no real extension", "/etc/passwd", ""},
		{"no extension", "noext", ""},
		{"multiple dots keeps only the last", "archive.tar.gz", ".gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := buildImageKey(threadID, tt.clientFileName)

			// The key must always live under this thread's own prefix,
			// and must never carry the raw client filename through —
			// only its extension (if any) is allowed to leak into it.
			assert.True(t, strings.HasPrefix(key, "threads/"+threadID.String()+"/"),
				"key %q should be scoped under the thread's own prefix", key)
			assert.NotContains(t, key, "..")
			assert.NotContains(t, key, "etc/passwd")

			if tt.wantSuffix != "" {
				assert.True(t, strings.HasSuffix(key, tt.wantSuffix))
			}
		})
	}
}
