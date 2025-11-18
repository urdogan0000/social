package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUsernameExists     = errors.New("username already exists")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

