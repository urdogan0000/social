package domain

import "errors"

// Base domain errors - tüm modüller bu base error'ları kullanabilir
var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation failed")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal error")
)

// User specific errors
var (
	ErrUserNotFound      = errors.Join(ErrNotFound, errors.New("user"))
	ErrUserAlreadyExists = errors.Join(ErrConflict, errors.New("user already exists"))
	ErrInvalidUsername   = errors.Join(ErrValidation, errors.New("invalid username"))
	ErrInvalidEmail       = errors.Join(ErrValidation, errors.New("invalid email"))
	ErrInvalidPassword    = errors.Join(ErrValidation, errors.New("invalid password"))
)

// Post specific errors
var (
	ErrPostNotFound  = errors.Join(ErrNotFound, errors.New("post"))
	ErrPostForbidden = errors.Join(ErrForbidden, errors.New("you can only modify your own posts"))
	ErrInvalidTitle  = errors.Join(ErrValidation, errors.New("invalid title"))
	ErrInvalidContent = errors.Join(ErrValidation, errors.New("invalid content"))
)

// Auth specific errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.Join(ErrUnauthorized, errors.New("invalid or expired token"))
)
