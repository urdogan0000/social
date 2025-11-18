package posts

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, nil
	}
	return pq.Array(a).Value()
}

func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		var arr pq.StringArray
		if err := arr.Scan(v); err != nil {
			return err
		}
		*a = StringArray(arr)
		return nil
	case string:
		var arr pq.StringArray
		if err := arr.Scan(v); err != nil {
			return err
		}
		*a = StringArray(arr)
		return nil
	case []string:
		*a = StringArray(v)
		return nil
	default:
		return errors.New("cannot scan into StringArray")
	}
}

type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"not null;size:255" json:"title"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	Tags      StringArray    `gorm:"type:text[]" json:"tags"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Model) TableName() string {
	return "posts"
}

