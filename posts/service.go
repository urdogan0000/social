package posts

import (
	"context"

	"github.com/urdogan0000/social/users"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uint) (*users.Model, error)
}

type Service struct {
	repo     Repository
	userRepo UserRepository
}

func NewService(repo Repository, userRepo UserRepository) *Service {
	return &Service{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID uint, req CreateRequest) (*Response, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	post := &Model{
		Title:   req.Title,
		Content: req.Content,
		UserID:  userID,
		Tags:    req.Tags,
	}

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, err
	}

	return s.toResponse(post), nil
}

func (s *Service) GetByID(ctx context.Context, id uint) (*Response, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResponse(post), nil
}

func (s *Service) GetByUserID(ctx context.Context, userID uint, limit, offset int) (*ListResponse, error) {
	posts, err := s.repo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
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
		return nil, err
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
		post.Tags = *req.Tags
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, err
	}

	return s.toResponse(post), nil
}

func (s *Service) Delete(ctx context.Context, id uint, userID uint) error {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return ErrForbidden
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int) (*ListResponse, error) {
	posts, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
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
		Tags:      post.Tags,
		CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: post.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

