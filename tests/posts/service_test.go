package posts_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/internal/events"
	"github.com/urdogan0000/social/posts"
)

type mockRepository struct {
	posts      map[uint]*posts.Model
	createErr  error
	getByIDErr error
	updateErr  error
	deleteErr  error
}

func (m *mockRepository) Create(ctx context.Context, post *posts.Model) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.posts == nil {
		m.posts = make(map[uint]*posts.Model)
	}
	post.ID = uint(len(m.posts) + 1)
	m.posts[post.ID] = post
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uint) (*posts.Model, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if post, ok := m.posts[id]; ok {
		return post, nil
	}
	return nil, posts.ErrNotFound
}

func (m *mockRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]posts.Model, error) {
	var result []posts.Model
	for _, post := range m.posts {
		if post.UserID == userID {
			result = append(result, *post)
		}
	}
	return result, nil
}

func (m *mockRepository) Update(ctx context.Context, post *posts.Model) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.posts[post.ID]; !ok {
		return posts.ErrNotFound
	}
	m.posts[post.ID] = post
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.posts[id]; !ok {
		return posts.ErrNotFound
	}
	delete(m.posts, id)
	return nil
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]posts.Model, error) {
	var result []posts.Model
	for _, post := range m.posts {
		result = append(result, *post)
	}
	return result, nil
}

func (m *mockRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.posts)), nil
}

func (m *mockRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	count := int64(0)
	for _, post := range m.posts {
		if post.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockRepository) SearchByTitle(ctx context.Context, title string, limit, offset int) ([]posts.Model, error) {
	var result []posts.Model
	titleLower := strings.ToLower(title)
	for _, post := range m.posts {
		if strings.Contains(strings.ToLower(post.Title), titleLower) {
			result = append(result, *post)
		}
	}
	return result, nil
}

func (m *mockRepository) GetByTags(ctx context.Context, tags []string, limit, offset int) ([]posts.Model, error) {
	var result []posts.Model
	for _, post := range m.posts {
		for _, tag := range tags {
			for _, postTag := range post.Tags {
				if tag == string(postTag) {
					result = append(result, *post)
					break
				}
			}
		}
	}
	return result, nil
}

type mockUserRepository struct {
	users map[domain.UserID]*domain.User
}

func (m *mockUserRepository) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Exists(ctx context.Context, id domain.UserID) (bool, error) {
	_, ok := m.users[id]
	return ok, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		req         posts.CreateRequest
		userExists  bool
		userErr     error
		createErr   error
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful creation",
			userID:     1,
			req:        posts.CreateRequest{Title: "Test", Content: "Content"},
			userExists: true,
			wantErr:    false,
		},
		{
			name:        "user not found",
			userID:      1,
			req:         posts.CreateRequest{Title: "Test", Content: "Content"},
			userExists:  false,
			wantErr:     true,
			expectedErr: nil, // Error will be wrapped, so we just check for error existence
		},
		{
			name:       "user check error",
			userID:     1,
			req:        posts.CreateRequest{Title: "Test", Content: "Content"},
			userErr:    errors.New("database error"),
			wantErr:    true,
		},
		{
			name:       "create error",
			userID:     1,
			req:        posts.CreateRequest{Title: "Test", Content: "Content"},
			userExists: true,
			createErr:  errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{createErr: tt.createErr}
			userRepo := &mockUserRepository{
				users: make(map[domain.UserID]*domain.User),
			}
			if tt.userExists {
				userRepo.users[domain.UserID(tt.userID)] = &domain.User{
					ID: domain.UserID(tt.userID),
				}
			}
			eventBus := events.NewInMemoryEventBus()
			service := posts.NewService(repo, userRepo, eventBus, nil)

			ctx := context.Background()
			result, err := service.Create(ctx, tt.userID, tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Errorf("expected result but got nil")
				}
				if result != nil && result.UserID != tt.userID {
					t.Errorf("expected userID %d, got %d", tt.userID, result.UserID)
				}
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Test", Content: "Content", UserID: 1},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()
	post, err := service.GetByID(ctx, 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if post == nil {
		t.Errorf("expected post but got nil")
	}
	if post != nil && post.ID != 1 {
		t.Errorf("expected post ID 1, got %d", post.ID)
	}

	_, err = service.GetByID(ctx, 999)
	if err == nil {
		t.Errorf("expected error for non-existent post")
	}
	if !errors.Is(err, posts.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Update(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Original", Content: "Content", UserID: 1},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()
	newTitle := "Updated Title"
	req := posts.UpdateRequest{Title: &newTitle}

	// Test successful update by owner
	post, err := service.Update(ctx, 1, 1, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if post == nil {
		t.Errorf("expected post but got nil")
	}
	if post != nil && post.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %q", post.Title)
	}

	// Test forbidden - different user
	_, err = service.Update(ctx, 1, 2, req)
	if err == nil {
		t.Errorf("expected error for forbidden update")
	}
	if !errors.Is(err, posts.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_Delete(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Test", Content: "Content", UserID: 1},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()

	// Test successful delete by owner
	err := service.Delete(ctx, 1, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Test forbidden - different user
	repo.posts[1] = &posts.Model{ID: 1, Title: "Test", Content: "Content", UserID: 1}
	err = service.Delete(ctx, 1, 2)
	if err == nil {
		t.Errorf("expected error for forbidden delete")
	}
	if !errors.Is(err, posts.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestService_List(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Post 1", Content: "Content 1", UserID: 1},
			2: {ID: 2, Title: "Post 2", Content: "Content 2", UserID: 1},
			3: {ID: 3, Title: "Post 3", Content: "Content 3", UserID: 2},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()
	result, err := service.List(ctx, 10, 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Errorf("expected result but got nil")
	}
	if result != nil && result.Total != 3 {
		t.Errorf("expected total 3, got %d", result.Total)
	}
	if result != nil && len(result.Posts) != 3 {
		t.Errorf("expected 3 posts, got %d", len(result.Posts))
	}
}

func TestService_GetByUserID(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Post 1", Content: "Content 1", UserID: 1},
			2: {ID: 2, Title: "Post 2", Content: "Content 2", UserID: 1},
			3: {ID: 3, Title: "Post 3", Content: "Content 3", UserID: 2},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()
	result, err := service.GetByUserID(ctx, 1, 10, 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Errorf("expected result but got nil")
	}
	if result != nil && result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result != nil && len(result.Posts) != 2 {
		t.Errorf("expected 2 posts, got %d", len(result.Posts))
	}
}

func TestService_SearchByTitle(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Golang Tutorial", Content: "Content", UserID: 1},
			2: {ID: 2, Title: "Python Guide", Content: "Content", UserID: 1},
			3: {ID: 3, Title: "Golang Best Practices", Content: "Content", UserID: 2},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()
	results, err := service.SearchByTitle(ctx, "Golang", 10, 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestService_GetByTags(t *testing.T) {
	repo := &mockRepository{
		posts: map[uint]*posts.Model{
			1: {ID: 1, Title: "Post 1", Content: "Content", UserID: 1, Tags: posts.StringArray{"golang", "tutorial"}},
			2: {ID: 2, Title: "Post 2", Content: "Content", UserID: 1, Tags: posts.StringArray{"python", "guide"}},
			3: {ID: 3, Title: "Post 3", Content: "Content", UserID: 2, Tags: posts.StringArray{"golang", "best-practices"}},
		},
	}
	userRepo := &mockUserRepository{}
	eventBus := events.NewInMemoryEventBus()
	service := posts.NewService(repo, userRepo, eventBus, nil)

	ctx := context.Background()
	results, err := service.GetByTags(ctx, []string{"golang"}, 10, 0)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

