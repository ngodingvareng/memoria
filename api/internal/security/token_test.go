//go:build unit

package security_test

import (
	"testing"

	"github.com/ngodingvareng/memoria/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionTokenGenerator_GenerateSessionToken_Unique(t *testing.T) {
	g := security.NewSessionTokenGenerator()

	token1, err := g.GenerateSessionToken()
	require.NoError(t, err)
	token2, err := g.GenerateSessionToken()
	require.NoError(t, err)

	assert.NotEmpty(t, token1)
	assert.NotEqual(t, token1, token2, "two generated tokens should never collide")
}

func TestSessionTokenGenerator_HashToken_Deterministic(t *testing.T) {
	g := security.NewSessionTokenGenerator()

	hash1 := g.HashToken("some-raw-token")
	hash2 := g.HashToken("some-raw-token")

	assert.Equal(t, hash1, hash2, "hashing the same raw token twice must produce the same hash — that's what lookup-by-hash depends on")
}

func TestSessionTokenGenerator_HashToken_DifferentInputsDifferentHashes(t *testing.T) {
	g := security.NewSessionTokenGenerator()

	assert.NotEqual(t, g.HashToken("token-a"), g.HashToken("token-b"))
}

func TestSessionTokenGenerator_HashToken_NeverEqualsRawToken(t *testing.T) {
	g := security.NewSessionTokenGenerator()

	raw, err := g.GenerateSessionToken()
	require.NoError(t, err)

	// A sanity check against accidentally "hashing" as a no-op — the
	// whole point is the raw token and its hash must never be the same
	// value (see sessions.token_hash's design note in schema.sql).
	assert.NotEqual(t, raw, g.HashToken(raw))
}
