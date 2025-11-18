package domain

import (
	"context"
	"errors"
)

type UserID uint

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserExistsChecker interface {
	UserExists(ctx context.Context, userID UserID) (bool, error)
}
