package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptSaltBytes = 16
	scryptKeyLen    = 32
	// N must be a power of 2. 32768 (2^15) costs roughly 32 MB of memory
	// per hash operation (~128*N*r bytes) — fine for typical login
	// traffic, but worth knowing this scales with concurrent
	// logins/registrations, unlike bcrypt's much smaller fixed footprint.
	scryptN = 32768
	scryptR = 8 // block size
	scryptP = 1 // parallelization

)

// ScryptHasher implements usecase.PasswordHasher using
// golang.org/x/crypto/scrypt. usecase never imports that package
// directly — same reasoning as everywhere else infrastructure details
// are kept behind an interface.
//
// Stored hash format: "$scrypt$N$r$p$<base64 salt>$<base64 derived key>".
// Encoding N/r/p into the hash itself (rather than only reading them
// from current config) means the cost parameters can be raised later
// without breaking the ability to verify passwords hashed under the old,
// lighter parameters.
type ScryptHasher struct {
	n, r, p, keyLen int
}

func NewScryptHasher() *ScryptHasher {
	return &ScryptHasher{n: scryptN, r: scryptR, p: scryptP, keyLen: scryptKeyLen}
}

func (h *ScryptHasher) Hash(password string) (string, error) {
	salt := make([]byte, scryptSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	derived, err := scrypt.Key([]byte(password), salt, h.n, h.r, h.p, h.keyLen)
	if err != nil {
		return "", fmt.Errorf("deriving key: %w", err)
	}

	return fmt.Sprintf("$scrypt$%d$%d$%d$%s$%s",
		h.n, h.r, h.p,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(derived),
	), nil
}

func (h *ScryptHasher) Compare(hash, password string) error {
	n, r, p, salt, storedKey, err := parseScryptHash(hash)
	if err != nil {
		return err
	}

	candidateKey, err := scrypt.Key([]byte(password), salt, n, r, p, len(storedKey))
	if err != nil {
		return fmt.Errorf("deriving key: %w", err)
	}

	// Constant-time comparison: a plain == or bytes.Equal would let an
	// attacker learn how many leading bytes matched from response timing.
	if subtle.ConstantTimeCompare(candidateKey, storedKey) != 1 {
		return errors.New("scrypt: password does not match")
	}
	return nil
}

func parseScryptHash(encoded string) (n, r, p int, salt, key []byte, err error) {
	// Splitting "$scrypt$N$r$p$salt$key" by "$" yields a leading empty
	// string (from the first "$"), then 6 real fields.
	parts := strings.Split(encoded, "$")
	if len(parts) != 7 || parts[0] != "" || parts[1] != "scrypt" {
		return 0, 0, 0, nil, nil, errors.New("scrypt: invalid hash format")
	}

	n, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("scrypt: invalid N: %w", err)
	}
	r, err = strconv.Atoi(parts[3])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("scrypt: invalid r: %w", err)
	}
	p, err = strconv.Atoi(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("scrypt: invalid p: %w", err)
	}
	salt, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("scrypt: invalid salt encoding: %w", err)
	}
	key, err = base64.RawStdEncoding.DecodeString(parts[6])
	if err != nil {
		return 0, 0, 0, nil, nil, fmt.Errorf("scrypt: invalid key encoding: %w", err)
	}

	return n, r, p, salt, key, nil
}
