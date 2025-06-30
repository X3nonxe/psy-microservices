package entities

import "errors"

var (
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInternal           = errors.New("internal server error")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenRevoked       = errors.New("token has been revoked")
)
