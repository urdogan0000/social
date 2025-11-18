package users

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil && err != ErrNotFound {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrAlreadyExists
	}

	existingUser, err = s.repo.GetByEmail(ctx, req.Email)
	if err != nil && err != ErrNotFound {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &Model{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toResponse(user), nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Response, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %d: %w", id, err)
	}
	return s.toResponse(user), nil
}

func (s *Service) GetByUsername(ctx context.Context, username string) (*Response, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username %q: %w", username, err)
	}
	return s.toResponse(user), nil
}

func (s *Service) Update(ctx context.Context, id uint, req UpdateRequest) (*Response, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %d: %w", id, err)
	}

	if req.Username != nil {
		existingUser, err := s.repo.GetByUsername(ctx, *req.Username)
		if err != nil && err != ErrNotFound {
			return nil, fmt.Errorf("failed to check username existence: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrAlreadyExists
		}
		user.Username = *req.Username
	}

	if req.Email != nil {
		existingUser, err := s.repo.GetByEmail(ctx, *req.Email)
		if err != nil && err != ErrNotFound {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrAlreadyExists
		}
		user.Email = *req.Email
	}

	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user %d: %w", id, err)
	}

	return s.toResponse(user), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user %d: %w", id, err)
	}
	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) (*ListResponse, error) {
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	responses := make([]Response, len(users))
	for i, user := range users {
		responses[i] = *s.toResponse(&user)
	}

	return &ListResponse{
		Users: responses,
		Total: total,
		Limit: limit,
		Offset: offset,
	}, nil
}

func (s *Service) toResponse(user *Model) *Response {
	return &Response{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

