package posts

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, post *Model) error
	GetByID(ctx context.Context, id uint) (*Model, error)
	GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]Model, error)
	Update(ctx context.Context, post *Model) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]Model, error)
	Count(ctx context.Context) (int64, error)
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	SearchByTitle(ctx context.Context, title string, limit, offset int) ([]Model, error)
	GetByTags(ctx context.Context, tags []string, limit, offset int) ([]Model, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, post *Model) error {
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		return err
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Model, error) {
	var post Model
	if err := r.db.WithContext(ctx).First(&post, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *repository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]Model, error) {
	var posts []Model
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *repository) Update(ctx context.Context, post *Model) error {
	if err := r.db.WithContext(ctx).Save(post).Error; err != nil {
		return err
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&Model{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Model, error) {
	var posts []Model
	if err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *repository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Model{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&Model{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) SearchByTitle(ctx context.Context, title string, limit, offset int) ([]Model, error) {
	var posts []Model
	if err := r.db.WithContext(ctx).
		Where("LOWER(title) LIKE LOWER(?)", "%"+title+"%").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *repository) GetByTags(ctx context.Context, tags []string, limit, offset int) ([]Model, error) {
	var posts []Model
	query := r.db.WithContext(ctx)

	for _, tag := range tags {
		query = query.Or("? = ANY(tags)", tag)
	}

	if err := query.
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}
