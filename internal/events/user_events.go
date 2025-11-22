package events

import "github.com/urdogan0000/social/internal/domain"

// UserCreated is fired when a user is created
type UserCreated struct {
	UserID   domain.UserID
	Username string
	Email    string
}

func (e UserCreated) Type() string {
	return "user.created"
}

// UserUpdated is fired when a user is updated
type UserUpdated struct {
	UserID   domain.UserID
	Username string
	Email    string
}

func (e UserUpdated) Type() string {
	return "user.updated"
}

// UserDeleted is fired when a user is deleted
type UserDeleted struct {
	UserID domain.UserID
}

func (e UserDeleted) Type() string {
	return "user.deleted"
}

