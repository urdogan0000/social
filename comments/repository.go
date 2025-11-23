package comments

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, comment *Model) error
	GetByID(ctx context.Context, id uint) (*Model, error)
	GetByPostID(ctx context.Context, postID uint, limit, offset int) ([]Model, error)
	Update(ctx context.Context, comment *Model) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]Model, error)
	Count(ctx context.Context) (int64, error)
	CountByPostID(ctx context.Context, postID uint) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, comment *Model) error {
	if err := r.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Model, error) {
	var comment Model
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&comment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}
	return &comment, nil
}

func (r *repository) GetByPostID(ctx context.Context, postID uint, limit, offset int) ([]Model, error) {
	var comments []Model
	query := r.db.WithContext(ctx).Where("post_id = ?", postID)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to get comments by post id: %w", err)
	}
	return comments, nil
}

func (r *repository) Update(ctx context.Context, comment *Model) error {
	if err := r.db.WithContext(ctx).Save(comment).Error; err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&Model{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Model, error) {
	var comments []Model
	query := r.db.WithContext(ctx)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Order("created_at DESC").Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}
	return comments, nil
}

func (r *repository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Model{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}
	return count, nil
}

func (r *repository) CountByPostID(ctx context.Context, postID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&Model{}).Where("post_id = ?", postID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count comments by post id: %w", err)
	}
	return count, nil
}
