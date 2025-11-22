package users

import (
	"context"
	"fmt"

	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/internal/events"
	"github.com/urdogan0000/social/internal/db"
)

type Service struct {
	repo            Repository
	eventBus        events.EventBus
	transactionMgr  db.TransactionManager
}

func NewService(repo Repository, eventBus events.EventBus, transactionMgr db.TransactionManager) *Service {
	return &Service{
		repo:           repo,
		eventBus:       eventBus,
		transactionMgr: transactionMgr,
	}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	// Create domain user
	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
	}

	// Validate domain model
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set password using domain method
	if err := user.SetPassword(req.Password); err != nil {
		return nil, fmt.Errorf("failed to set password: %w", err)
	}

	// Check if username exists
	existingUser, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil && err != ErrNotFound {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrAlreadyExists
	}

	// Check if email exists
	existingUser, err = s.repo.GetByEmail(ctx, req.Email)
	if err != nil && err != ErrNotFound {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrAlreadyExists
	}

	// Convert domain user to model
	model := s.domainToModel(user)

	// Use transaction if available
	var createErr error
	if s.transactionMgr != nil {
		createErr = s.transactionMgr.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.repo.Create(txCtx, model)
		})
	} else {
		createErr = s.repo.Create(ctx, model)
	}

	if createErr != nil {
		return nil, fmt.Errorf("failed to create user: %w", createErr)
	}

	// Update domain user with generated ID
	user.ID = domain.UserID(model.ID)

	// Publish event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.UserCreated{
			UserID:   user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}

	return s.toResponse(model), nil
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
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %d: %w", id, err)
	}

	// Convert to domain model
	user := s.modelToDomain(model)

	// Update fields using domain methods
	if req.Username != nil {
		existingUser, err := s.repo.GetByUsername(ctx, *req.Username)
		if err != nil && err != ErrNotFound {
			return nil, fmt.Errorf("failed to check username existence: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrAlreadyExists
		}
		if err := user.UpdateUsername(*req.Username); err != nil {
			return nil, fmt.Errorf("failed to update username: %w", err)
		}
	}

	if req.Email != nil {
		existingUser, err := s.repo.GetByEmail(ctx, *req.Email)
		if err != nil && err != ErrNotFound {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrAlreadyExists
		}
		if err := user.UpdateEmail(*req.Email); err != nil {
			return nil, fmt.Errorf("failed to update email: %w", err)
		}
	}

	if req.Password != nil {
		if err := user.SetPassword(*req.Password); err != nil {
			return nil, fmt.Errorf("failed to set password: %w", err)
		}
	}

	// Validate updated user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert back to model
	updatedModel := s.domainToModel(user)
	updatedModel.ID = model.ID
	updatedModel.CreatedAt = model.CreatedAt

	// Use transaction if available
	var updateErr error
	if s.transactionMgr != nil {
		updateErr = s.transactionMgr.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.repo.Update(txCtx, updatedModel)
		})
	} else {
		updateErr = s.repo.Update(ctx, updatedModel)
	}

	if updateErr != nil {
		return nil, fmt.Errorf("failed to update user %d: %w", id, updateErr)
	}

	// Publish event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.UserUpdated{
			UserID:   user.ID,
			Username: user.Username,
			Email:    user.Email,
		})
	}

	return s.toResponse(updatedModel), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	// Check if user exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user by id %d: %w", id, err)
	}

	userID := domain.UserID(id)

	// Use transaction if available
	var deleteErr error
	if s.transactionMgr != nil {
		deleteErr = s.transactionMgr.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.repo.Delete(txCtx, id)
		})
	} else {
		deleteErr = s.repo.Delete(ctx, id)
	}

	if deleteErr != nil {
		return fmt.Errorf("failed to delete user %d: %w", id, deleteErr)
	}

	// Publish event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.UserDeleted{
			UserID: userID,
		})
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

// domainToModel converts domain User to repository Model
func (s *Service) domainToModel(user *domain.User) *Model {
	return &Model{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}
}

// modelToDomain converts repository Model to domain User
func (s *Service) modelToDomain(model *Model) *domain.User {
	return &domain.User{
		ID:       domain.UserID(model.ID),
		Username: model.Username,
		Email:    model.Email,
		Password: model.Password,
	}
}

