package security

import (
	"context"
	"fmt"

	"github.com/ngodingvareng/memoria/internal/errs"
	"github.com/ngodingvareng/memoria/internal/usecase"
	"google.golang.org/api/idtoken"
)

// GoogleIDTokenVerifier implements usecase.GoogleIDTokenVerifier using
// google.golang.org/api/idtoken. usecase never imports that package
// directly — same reasoning as everywhere else infrastructure details
// are hidden behind an interface.
type GoogleIDTokenVerifier struct {
	clientID string
}

func NewGoogleIDTokenVerifier(clientID string) *GoogleIDTokenVerifier {
	return &GoogleIDTokenVerifier{clientID: clientID}
}

// Verify validates the token's signature, expiry, issuer, and audience
// (must match clientID) against Google's published keys, then extracts
// the identity claims a login flow needs. Any failure collapses to
// errs.ErrInvalidToken — the caller shouldn't distinguish "expired" from
// "bad signature" from "wrong audience" any more than ResetPassword does
// for its own token.
func (v *GoogleIDTokenVerifier) Verify(ctx context.Context, idToken string) (*usecase.GoogleIdentity, error) {
	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return nil, errs.ErrInvalidToken
	}

	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	name, _ := payload.Claims["name"].(string)

	if payload.Subject == "" || email == "" {
		return nil, fmt.Errorf("%w: missing sub or email claim", errs.ErrInvalidToken)
	}

	return &usecase.GoogleIdentity{
		Subject:       payload.Subject,
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
	}, nil
}
