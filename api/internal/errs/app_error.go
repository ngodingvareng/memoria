package errs

import (
	"net/http"

	"github.com/ngodingvareng/memoria/internal/validate"
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

var (
	// 400 Bad Request
	ErrBadRequest       = New(http.StatusBadRequest, "bad request")
	ErrInvalidInput     = New(http.StatusBadRequest, "invalid input")
	ErrValidationFailed = New(http.StatusBadRequest, "validation failed")

	// 401 Unauthorized
	ErrUnauthorized       = New(http.StatusUnauthorized, "unauthorized")
	ErrInvalidToken       = New(http.StatusUnauthorized, "invalid token")
	ErrTokenExpired       = New(http.StatusUnauthorized, "token expired")
	ErrInvalidCredentials = New(http.StatusUnauthorized, "invalid credentials")

	// 403 Forbidden
	ErrForbidden              = New(http.StatusForbidden, "forbidden")
	ErrAccessDenied           = New(http.StatusForbidden, "access denied")
	ErrInsufficientPermission = New(http.StatusForbidden, "insufficient permission")
	// ErrOnboardingIncomplete guards every activity endpoint behind
	// middleware.RequireOnboarded — an account that hasn't finished
	// onboarding yet (currently: claiming a username, the mandatory
	// post-register/post-Google-login step) can only call the endpoints
	// needed to finish it, nothing else.
	ErrOnboardingIncomplete = New(http.StatusForbidden, "onboarding incomplete")

	// 404 Not Found
	ErrNotFound = New(http.StatusNotFound, "not found")

	// 409 Conflict
	ErrConflict              = New(http.StatusConflict, "conflict")
	ErrEmailAlreadyExists    = New(http.StatusConflict, "email already registered")
	ErrUsernameAlreadyExists = New(http.StatusConflict, "username already registered")
	ErrUserAlreadyExists     = New(http.StatusConflict, "user already exists")
	ErrLastCircleAdmin       = New(http.StatusConflict, "circle must keep at least one admin; promote another member first")

	// 422 Unprocessable Entity
	ErrUnprocessableEntity = New(http.StatusUnprocessableEntity, "unprocessable entity")

	// 423 Locked
	ErrAccountLocked = New(http.StatusLocked, "account temporarily locked due to too many failed login attempts")

	// 429 Too Many Requests
	ErrTooManyRequests = New(http.StatusTooManyRequests, "too many requests")
)

type ValidationError struct {
	Errors []*validate.ErrorResponse
}

func (e *ValidationError) Error() string {
	return "validation error"
}
