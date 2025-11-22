package domain

import "context"

// UserRepository defines the interface for user repository operations
// This allows other modules to depend on the interface rather than concrete implementation
type UserRepository interface {
	GetByID(ctx context.Context, id UserID) (*User, error)
	Exists(ctx context.Context, id UserID) (bool, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// PostRepository defines the interface for post repository operations
type PostRepository interface {
	GetByID(ctx context.Context, id PostID) (*Post, error)
	GetByUserID(ctx context.Context, userID UserID) ([]*Post, error)
	Exists(ctx context.Context, id PostID) (bool, error)
}

