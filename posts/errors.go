package posts

import "github.com/urdogan0000/social/internal/domain"

var (
	ErrNotFound  = domain.ErrPostNotFound
	ErrForbidden = domain.ErrPostForbidden
)

