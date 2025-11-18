package users

import (
	"errors"

	"github.com/urdogan0000/social/internal/domain"
)

var (
	ErrNotFound      = domain.ErrUserNotFound
	ErrAlreadyExists = errors.New("user already exists")
)

