package errs

import (
	"fmt"
	"net/http"

	"github.com/moez-rd/memoria/pkg/util"
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

func NewResourceNotFoundError(resource string, id any) *AppError {
	return &AppError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s with ID %v not found", resource, id),
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

	// 404 Not Found
	ErrNotFound = New(http.StatusNotFound, "not found")

	// 409 Conflict
	ErrConflict              = New(http.StatusConflict, "conflict")
	ErrEmailAlreadyExists    = New(http.StatusConflict, "email already registered")
	ErrUsernameAlreadyExists = New(http.StatusConflict, "username already registered")
	ErrUserAlreadyExists     = New(http.StatusConflict, "user already exists")

	// 422 Unprocessable Entity
	ErrUnprocessableEntity = New(http.StatusUnprocessableEntity, "unprocessable entity")
)

type ValidationError struct {
	Errors []*util.ErrorResponse // assuming this is your util return type
}

func (e *ValidationError) Error() string {
	return "validation error"
}
