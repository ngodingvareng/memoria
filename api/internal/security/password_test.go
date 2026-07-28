//go:build unit

package security_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ngodingvareng/memoria/internal/security"
)

func TestScryptHasher_HashAndCompare_RoundTrip(t *testing.T) {
	h := security.NewScryptHasher()

	hash, err := h.Hash("correct horse battery staple")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(hash, "$scrypt$"))

	err = h.Compare(hash, "correct horse battery staple")
	assert.NoError(t, err)
}

func TestScryptHasher_Compare_WrongPassword(t *testing.T) {
	h := security.NewScryptHasher()

	hash, err := h.Hash("the real password")
	require.NoError(t, err)

	err = h.Compare(hash, "a totally different password")

	assert.Error(t, err)
}

func TestScryptHasher_Compare_MalformedHash(t *testing.T) {
	h := security.NewScryptHasher()

	err := h.Compare("not-a-valid-hash-at-all", "whatever")

	assert.Error(t, err)
}

func TestScryptHasher_DifferentSaltsProduceDifferentHashes(t *testing.T) {
	h := security.NewScryptHasher()

	hash1, err := h.Hash("same password")
	require.NoError(t, err)
	hash2, err := h.Hash("same password")
	require.NoError(t, err)

	assert.NotEqual(t, hash1, hash2, "two hashes of the same password must differ (random salt)")

	// Both must still independently verify correctly despite differing.
	assert.NoError(t, h.Compare(hash1, "same password"))
	assert.NoError(t, h.Compare(hash2, "same password"))
}
