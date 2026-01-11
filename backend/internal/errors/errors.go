package errors

import (
	"errors"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInvalidInput  = errors.New("invalid input")
	ErrAlreadyExists = errors.New("resource already exists")
)
