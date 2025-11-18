package posts

import "errors"

var (
	ErrNotFound  = errors.New("post not found")
	ErrForbidden = errors.New("forbidden: you can only modify your own posts")
)

