package events

import "github.com/urdogan0000/social/internal/domain"

// PostCreated is fired when a post is created
type PostCreated struct {
	PostID domain.PostID
	UserID domain.UserID
	Title  string
}

func (e PostCreated) Type() string {
	return "post.created"
}

// PostUpdated is fired when a post is updated
type PostUpdated struct {
	PostID domain.PostID
	UserID domain.UserID
	Title  string
}

func (e PostUpdated) Type() string {
	return "post.updated"
}

// PostDeleted is fired when a post is deleted
type PostDeleted struct {
	PostID domain.PostID
	UserID domain.UserID
}

func (e PostDeleted) Type() string {
	return "post.deleted"
}

