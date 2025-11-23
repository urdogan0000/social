package comments

import (
	"context"
	"fmt"

	"github.com/urdogan0000/social/internal/db"
	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/internal/events"
)

type Service struct {
	repo           Repository
	userRepo       domain.UserRepository
	eventBus       events.EventBus
	transactionMgr db.TransactionManager
}

func NewService(repo Repository, userRepo domain.UserRepository, eventBus events.EventBus, transactionMgr db.TransactionManager) *Service {
	return &Service{
		repo:           repo,
		userRepo:       userRepo,
		eventBus:       eventBus,
		transactionMgr: transactionMgr,
	}
}

func (s *Service) Create(ctx context.Context, userID uint, req CreateRequest) (*Response, error) {
	comment := &Model{
		PostID:  req.PostID,
		Content: req.Content,
		UserID:  userID,
	}
	if err := s.repo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to create comment: %w", err)
	}
	response := s.toResponse(comment)
	return &response, nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Response, error) {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}
	response := s.toResponse(comment)
	return &response, nil
}

func (s *Service) GetByPostID(ctx context.Context, postID uint, limit, offset int) (*ListResponse, error) {
	comments, err := s.repo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments by post id: %w", err)
	}
	total, err := s.repo.CountByPostID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to count comments by post id: %w", err)
	}
	responses := make([]Response, len(comments))
	for i, comment := range comments {
		responses[i] = s.toResponse(&comment)
	}
	return &ListResponse{
		Comments: responses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *Service) Update(ctx context.Context, id uint, userID uint, req UpdateRequest) (*Response, error) {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment by id: %w", err)
	}

	// Check if user owns the comment
	if comment.UserID != userID {
		return nil, ErrForbidden
	}

	// Update content if provided
	if req.Content != nil {
		comment.Content = *req.Content
	}

	if err := s.repo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to update comment: %w", err)
	}

	response := s.toResponse(comment)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, id uint, userID uint) error {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get comment by id: %w", err)
	}

	// Check if user owns the comment
	if comment.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) (*ListResponse, error) {
	comments, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}

	responses := make([]Response, len(comments))
	for i, comment := range comments {
		responses[i] = s.toResponse(&comment)
	}

	return &ListResponse{
		Comments: responses,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

func (s *Service) toResponse(comment *Model) Response {
	return Response{
		ID:        comment.ID,
		PostID:    comment.PostID,
		Content:   comment.Content,
		UserID:    comment.UserID,
		CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: comment.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
