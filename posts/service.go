package posts

import (
	"context"
	"fmt"

	"github.com/urdogan0000/social/internal/domain"
)

type Service struct {
	repo        Repository
	userChecker domain.UserExistsChecker
}

func NewService(repo Repository, userChecker domain.UserExistsChecker) *Service {
	return &Service{
		repo:        repo,
		userChecker: userChecker,
	}
}

func (s *Service) Create(ctx context.Context, userID uint, req CreateRequest) (*Response, error) {
	exists, err := s.userChecker.UserExists(ctx, domain.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	post := &Model{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
		Tags:    StringArray(req.Tags),
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return s.toResponse(post), nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Response, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id %d: %w", id, err)
	}
	return s.toResponse(post), nil
}

func (s *Service) GetByUserID(ctx context.Context, userID uint, limit, offset int) (*ListResponse, error) {
	posts, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by user id %d: %w", userID, err)
	}

	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts by user id %d: %w", userID, err)
	}

	responses := make([]Response, len(posts))
	for i, post := range posts {
		responses[i] = *s.toResponse(&post)
	}

	return &ListResponse{
		Posts: responses,
		Total: total,
		Limit: limit,
		Offset: offset,
	}, nil
}

func (s *Service) Update(ctx context.Context, id uint, userID uint, req UpdateRequest) (*Response, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id %d: %w", id, err)
	}

	if post.UserID != userID {
		return nil, ErrForbidden
	}

	if req.Title != nil {
		post.Title = *req.Title
	}

	if req.Content != nil {
		post.Content = *req.Content
	}

	if req.Tags != nil {
		post.Tags = StringArray(*req.Tags)
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to update post %d: %w", id, err)
	}

	return s.toResponse(post), nil
}

func (s *Service) Delete(ctx context.Context, id uint, userID uint) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get post by id %d: %w", id, err)
	}

	if post.UserID != userID {
		return ErrForbidden
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete post %d: %w", id, err)
	}

	return nil
}

func (s *Service) List(ctx context.Context, limit, offset int) (*ListResponse, error) {
	posts, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	responses := make([]Response, len(posts))
	for i, post := range posts {
		responses[i] = *s.toResponse(&post)
	}

	return &ListResponse{
		Posts: responses,
		Total: total,
		Limit: limit,
		Offset: offset,
	}, nil
}

func (s *Service) SearchByTitle(ctx context.Context, title string, limit, offset int) ([]Response, error) {
	posts, err := s.repo.SearchByTitle(ctx, title, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts by title %q: %w", title, err)
	}

	responses := make([]Response, len(posts))
	for i, post := range posts {
		responses[i] = *s.toResponse(&post)
	}

	return responses, nil
}

func (s *Service) GetByTags(ctx context.Context, tags []string, limit, offset int) ([]Response, error) {
	posts, err := s.repo.GetByTags(ctx, tags, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts by tags: %w", err)
	}

	responses := make([]Response, len(posts))
	for i, post := range posts {
		responses[i] = *s.toResponse(&post)
	}

	return responses, nil
}

func (s *Service) toResponse(post *Model) *Response {
	return &Response{
		ID:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		UserID:    post.UserID,
		Tags:      []string(post.Tags),
		CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

