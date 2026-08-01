package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const refreshTokenBytes = 32 // 256 bits of entropy

// RefreshTokenGenerator generates an opaque random token that is sent to
// the client, and hashes it for storage. Just like the old session
// token schema — the only thing that changes is its role: from "checked on every
// request" to "exchanged for a new access token".
type RefreshTokenGenerator struct{}

func NewRefreshTokenGenerator() *RefreshTokenGenerator {
	return &RefreshTokenGenerator{}
}

func (g *RefreshTokenGenerator) Generate() (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating refresh token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func (g *RefreshTokenGenerator) Hash(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
