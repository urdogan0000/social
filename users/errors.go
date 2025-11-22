package users

import "github.com/urdogan0000/social/internal/domain"

var (
	ErrNotFound      = domain.ErrUserNotFound
	ErrAlreadyExists = domain.ErrUserAlreadyExists
)

