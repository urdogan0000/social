package posts

import (
	"context"
	"errors"
	"testing"

	"github.com/urdogan0000/social/internal/domain"
)

type mockRepository struct {
	posts      map[uint]*Model
	createErr  error
	getByIDErr error
	updateErr  error
	deleteErr  error
}

func (m *mockRepository) Create(ctx context.Context, post *Model) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.posts == nil {
		m.posts = make(map[uint]*Model)
	}
	post.ID = uint(len(m.posts) + 1)
	m.posts[post.ID] = post
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uint) (*Model, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if post, ok := m.posts[id]; ok {
		return post, nil
	}
	return nil, ErrNotFound
}

func (m *mockRepository) GetByUserID(ctx context.Context, userID uint, limit, offset int) ([]Model, error) {
	var result []Model
	for _, post := range m.posts {
		if post.UserID == userID {
			result = append(result, *post)
		}
	}
	return result, nil
}

func (m *mockRepository) Update(ctx context.Context, post *Model) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.posts[post.ID]; !ok {
		return ErrNotFound
	}
	m.posts[post.ID] = post
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.posts[id]; !ok {
		return ErrNotFound
	}
	delete(m.posts, id)
	return nil
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]Model, error) {
	var result []Model
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

func (m *mockRepository) SearchByTitle(ctx context.Context, title string, limit, offset int) ([]Model, error) {
	var result []Model
	for _, post := range m.posts {
		if post.Title == title {
			result = append(result, *post)
		}
	}
	return result, nil
}

func (m *mockRepository) GetByTags(ctx context.Context, tags []string, limit, offset int) ([]Model, error) {
	var result []Model
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

type mockUserChecker struct {
	exists bool
	err    error
}

func (m *mockUserChecker) UserExists(ctx context.Context, userID domain.UserID) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.exists, nil
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		req         CreateRequest
		userExists  bool
		userErr     error
		createErr   error
		wantErr     bool
		expectedErr error
	}{
		{
			name:       "successful creation",
			userID:     1,
			req:        CreateRequest{Title: "Test", Content: "Content"},
			userExists: true,
			wantErr:    false,
		},
		{
			name:        "user not found",
			userID:      1,
			req:         CreateRequest{Title: "Test", Content: "Content"},
			userExists:  false,
			wantErr:     true,
			expectedErr: ErrNotFound,
		},
		{
			name:       "user check error",
			userID:     1,
			req:        CreateRequest{Title: "Test", Content: "Content"},
			userErr:    errors.New("database error"),
			wantErr:    true,
		},
		{
			name:       "create error",
			userID:     1,
			req:        CreateRequest{Title: "Test", Content: "Content"},
			userExists: true,
			createErr:  errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{createErr: tt.createErr}
			checker := &mockUserChecker{exists: tt.userExists, err: tt.userErr}
			service := NewService(repo, checker)

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
		posts: map[uint]*Model{
			1: {ID: 1, Title: "Test", Content: "Content", UserID: 1},
		},
	}
	checker := &mockUserChecker{}
	service := NewService(repo, checker)

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
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

