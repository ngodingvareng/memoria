//go:build unit

package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/security"
)

func TestRefreshTokenGenerator_Generate_Unique(t *testing.T) {
	g := security.NewRefreshTokenGenerator()

	token1, err := g.Generate()
	require.NoError(t, err)
	token2, err := g.Generate()
	require.NoError(t, err)

	assert.NotEmpty(t, token1)
	assert.NotEqual(t, token1, token2, "two generated tokens should never collide")
}

func TestRefreshTokenGenerator_Hash_Deterministic(t *testing.T) {
	g := security.NewRefreshTokenGenerator()

	hash1 := g.Hash("some-raw-token")
	hash2 := g.Hash("some-raw-token")

	assert.Equal(t, hash1, hash2, "hashing the same raw token twice must produce the same hash — that's what lookup-by-hash depends on")
}

func TestRefreshTokenGenerator_Hash_DifferentInputsDifferentHashes(t *testing.T) {
	g := security.NewRefreshTokenGenerator()

	assert.NotEqual(t, g.Hash("token-a"), g.Hash("token-b"))
}

func TestRefreshTokenGenerator_Hash_NeverEqualsRawToken(t *testing.T) {
	g := security.NewRefreshTokenGenerator()

	raw, err := g.Generate()
	require.NoError(t, err)

	// A sanity check against accidentally "hashing" as a no-op — the
	// whole point is the raw token and its hash must never be the same
	// value (see refresh_tokens.token_hash's design note in schema.sql).
	assert.NotEqual(t, raw, g.Hash(raw))
}
