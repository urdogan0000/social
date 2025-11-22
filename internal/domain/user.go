package domain

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserID uint

type User struct {
	ID       UserID
	Username string
	Email    string
	Password []byte
}

// Validate validates user data
func (u *User) Validate() error {
	if len(u.Username) < 3 {
		return ErrInvalidUsername
	}
	if len(u.Username) > 100 {
		return ErrInvalidUsername
	}
	if len(u.Email) == 0 {
		return ErrInvalidEmail
	}
	// Email format validation should be done at DTO level with validator
	return nil
}

// SetPassword hashes and sets the password
func (u *User) SetPassword(plainPassword string) error {
	if len(plainPassword) < 6 {
		return ErrInvalidPassword
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.Join(ErrInternal, err)
	}
	u.Password = hashed
	return nil
}

// CheckPassword verifies if the provided password matches
func (u *User) CheckPassword(plainPassword string) bool {
	if len(u.Password) == 0 {
		return false
	}
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(plainPassword))
	return err == nil
}

// UpdateUsername updates username if valid
func (u *User) UpdateUsername(newUsername string) error {
	if len(newUsername) < 3 || len(newUsername) > 100 {
		return ErrInvalidUsername
	}
	u.Username = newUsername
	return nil
}

// UpdateEmail updates email if valid
func (u *User) UpdateEmail(newEmail string) error {
	if len(newEmail) == 0 {
		return ErrInvalidEmail
	}
	u.Email = newEmail
	return nil
}

type UserExistsChecker interface {
	UserExists(ctx context.Context, userID UserID) (bool, error)
}
