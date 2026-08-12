package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ngodingvareng/memoria/internal/errs"
)

// Claims are intentionally minimal — just the subject. Anything that can change
// (name, email, role) is intentionally not embedded in the token so it doesn't become
// stale during the token's lifetime; if needed, it is fetched fresh when required.
type accessTokenClaims struct {
	jwt.RegisteredClaims
}

// JWTAccessTokenIssuer implements usecase.AccessTokenIssuer using
// HS256. usecase never imports golang-jwt directly — just
// like every other infra detail that is hidden behind
// the interface.
type JWTAccessTokenIssuer struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

func NewJWTAccessTokenIssuer(secret []byte, ttl time.Duration, issuer string) *JWTAccessTokenIssuer {
	return &JWTAccessTokenIssuer{secret: secret, ttl: ttl, issuer: issuer}
}

func (j *JWTAccessTokenIssuer) Generate(userID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(j.ttl)

	claims := accessTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing access token: %w", err)
	}
	return signed, expiresAt, nil
}

func (j *JWTAccessTokenIssuer) Parse(tokenString string) (uuid.UUID, error) {
	claims := &accessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return uuid.Nil, errs.ErrTokenExpired
		}
		return uuid.Nil, errs.ErrUnauthorized
	}
	if !token.Valid {
		return uuid.Nil, errs.ErrUnauthorized
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("parsing subject claim: %w", err)
	}
	return userID, nil
}
