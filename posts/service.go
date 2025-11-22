package posts

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
	// Check if user exists using domain repository
	_, err := s.userRepo.GetByID(ctx, domain.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// Create domain post
	post := &domain.Post{
		Title:   req.Title,
		Content: req.Content,
		UserID:  domain.UserID(userID),
		Tags:    req.Tags,
	}

	// Validate domain model
	if err := post.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert to model
	model := s.domainToModel(post)

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
		return nil, fmt.Errorf("failed to create post: %w", createErr)
	}

	// Update domain post with generated ID
	post.ID = domain.PostID(model.ID)

	// Publish event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.PostCreated{
			PostID: post.ID,
			UserID: post.UserID,
			Title:  post.Title,
		})
	}

	return s.toResponse(model), nil
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
		Posts:  responses,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

func (s *Service) Update(ctx context.Context, id uint, userID uint, req UpdateRequest) (*Response, error) {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by id %d: %w", id, err)
	}

	// Convert to domain model
	post := s.modelToDomain(model)

	// Check permission using domain method
	if !post.CanBeEditedBy(domain.UserID(userID)) {
		return nil, ErrForbidden
	}

	// Update fields using domain methods
	if req.Title != nil {
		if err := post.UpdateTitle(*req.Title); err != nil {
			return nil, fmt.Errorf("failed to update title: %w", err)
		}
	}

	if req.Content != nil {
		if err := post.UpdateContent(*req.Content); err != nil {
			return nil, fmt.Errorf("failed to update content: %w", err)
		}
	}

	if req.Tags != nil {
		post.UpdateTags(*req.Tags)
	}

	// Validate updated post
	if err := post.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert back to model
	updatedModel := s.domainToModel(post)
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
		return nil, fmt.Errorf("failed to update post %d: %w", id, updateErr)
	}

	// Publish event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.PostUpdated{
			PostID: post.ID,
			UserID: post.UserID,
			Title:  post.Title,
		})
	}

	return s.toResponse(updatedModel), nil
}

func (s *Service) Delete(ctx context.Context, id uint, userID uint) error {
	model, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get post by id %d: %w", id, err)
	}

	// Convert to domain model
	post := s.modelToDomain(model)

	// Check permission using domain method
	if !post.CanBeDeletedBy(domain.UserID(userID)) {
		return ErrForbidden
	}

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
		return fmt.Errorf("failed to delete post %d: %w", id, deleteErr)
	}

	// Publish event
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, events.PostDeleted{
			PostID: post.ID,
			UserID: post.UserID,
		})
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
		Posts:  responses,
		Total:  total,
		Limit:  limit,
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

// domainToModel converts domain Post to repository Model
func (s *Service) domainToModel(post *domain.Post) *Model {
	return &Model{
		Title:   post.Title,
		Content: post.Content,
		UserID:  uint(post.UserID),
		Tags:    StringArray(post.Tags),
	}
}

// modelToDomain converts repository Model to domain Post
func (s *Service) modelToDomain(model *Model) *domain.Post {
	return &domain.Post{
		ID:      domain.PostID(model.ID),
		Title:   model.Title,
		Content: model.Content,
		UserID:  domain.UserID(model.UserID),
		Tags:    []string(model.Tags),
	}
}
