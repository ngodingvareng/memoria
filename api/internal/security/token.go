package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const sessionTokenBytes = 32 // 256 bits of entropy

type SessionTokenGenerator struct{}

func NewSessionTokenGenerator() *SessionTokenGenerator {
	return &SessionTokenGenerator{}
}

// GenerateSessionToken returns a new random, hex-encoded raw token. This
// is the value that goes to the client via cookie — it is never stored
// anywhere; only HashToken's output (sessions.token_hash) is.
func (g *SessionTokenGenerator) GenerateSessionToken() (string, error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// HashToken returns the SHA-256 hex digest of a raw token.
func (g *SessionTokenGenerator) HashToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
