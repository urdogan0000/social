package comments

import (
	"time"

	"gorm.io/gorm"
)

type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PostID    uint           `gorm:"not null;index" json:"post_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Model) TableName() string {
	return "comments"
}
