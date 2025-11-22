package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/urdogan0000/social/internal/db"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, user *Model) error
	GetByID(ctx context.Context, id uint) (*Model, error)
	GetByUsername(ctx context.Context, username string) (*Model, error)
	GetByEmail(ctx context.Context, email string) (*Model, error)
	Update(ctx context.Context, user *Model) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]Model, error)
	Count(ctx context.Context) (int64, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// getDB retrieves the database connection from context or uses default
func (r *repository) getDB(ctx context.Context) *gorm.DB {
	return db.GetDBFromContext(ctx, r.db)
}

func (r *repository) Create(ctx context.Context, user *Model) error {
	if err := r.getDB(ctx).WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uint) (*Model, error) {
	var user Model
	if err := r.getDB(ctx).WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id %d: %w", id, err)
	}
	return &user, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (*Model, error) {
	var user Model
	if err := r.getDB(ctx).WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by username %q: %w", username, err)
	}
	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*Model, error) {
	var user Model
	if err := r.getDB(ctx).WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email %q: %w", email, err)
	}
	return &user, nil
}

func (r *repository) Update(ctx context.Context, user *Model) error {
	if err := r.getDB(ctx).WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user %d: %w", user.ID, err)
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.getDB(ctx).WithContext(ctx).Delete(&Model{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Model, error) {
	var users []Model
	if err := r.getDB(ctx).WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *repository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.getDB(ctx).WithContext(ctx).Model(&Model{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
