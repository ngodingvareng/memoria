//go:build unit

package security_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/security"
)

func TestJWTAccessTokenIssuer_GenerateAndParse_RoundTrip(t *testing.T) {
	issuer := security.NewJWTAccessTokenIssuer([]byte("test-secret"), 15*time.Minute, "memoria-test")
	userID := uuid.New()

	token, expiresAt, err := issuer.Generate(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), expiresAt, time.Second)

	parsedID, err := issuer.Parse(token)

	require.NoError(t, err)
	assert.Equal(t, userID, parsedID)
}

func TestJWTAccessTokenIssuer_Parse_Expired(t *testing.T) {
	// A negative TTL mints a token that's already expired the moment
	// it's generated — the simplest way to exercise the expiry branch
	// without sleeping in a test.
	issuer := security.NewJWTAccessTokenIssuer([]byte("test-secret"), -1*time.Minute, "memoria-test")

	token, _, err := issuer.Generate(uuid.New())
	require.NoError(t, err)

	_, err = issuer.Parse(token)

	assert.ErrorIs(t, err, errs.ErrTokenExpired)
}

func TestJWTAccessTokenIssuer_Parse_WrongSecret(t *testing.T) {
	signer := security.NewJWTAccessTokenIssuer([]byte("secret-a"), 15*time.Minute, "memoria-test")
	verifier := security.NewJWTAccessTokenIssuer([]byte("secret-b"), 15*time.Minute, "memoria-test")

	token, _, err := signer.Generate(uuid.New())
	require.NoError(t, err)

	_, err = verifier.Parse(token)

	assert.ErrorIs(t, err, errs.ErrUnauthorized,
		"a token signed with a different secret must fail verification, not be silently accepted")
}

func TestJWTAccessTokenIssuer_Parse_MalformedToken(t *testing.T) {
	issuer := security.NewJWTAccessTokenIssuer([]byte("test-secret"), 15*time.Minute, "memoria-test")

	_, err := issuer.Parse("not-a-real-jwt")

	assert.ErrorIs(t, err, errs.ErrUnauthorized)
}
