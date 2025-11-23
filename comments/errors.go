package comments

import (
	"errors"

	"github.com/urdogan0000/social/internal/domain"
	"gorm.io/gorm"
)

var (
	ErrNotFound  = errors.Join(domain.ErrNotFound, errors.New("comment"))
	ErrForbidden = errors.Join(domain.ErrForbidden, errors.New("you can only modify your own comments"))
)

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
