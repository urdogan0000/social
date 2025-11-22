package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/urdogan0000/social/internal/events"
	"github.com/urdogan0000/social/users"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	users      map[uint]*users.Model
	createErr  error
	getByIDErr error
	updateErr  error
	deleteErr  error
}

func (m *mockRepository) Create(ctx context.Context, user *users.Model) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.users == nil {
		m.users = make(map[uint]*users.Model)
	}
	user.ID = uint(len(m.users) + 1)
	m.users[user.ID] = user
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uint) (*users.Model, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, users.ErrNotFound
}

func (m *mockRepository) GetByUsername(ctx context.Context, username string) (*users.Model, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, users.ErrNotFound
}

func (m *mockRepository) GetByEmail(ctx context.Context, email string) (*users.Model, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, users.ErrNotFound
}

func (m *mockRepository) Update(ctx context.Context, user *users.Model) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.users[user.ID]; !ok {
		return users.ErrNotFound
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.users[id]; !ok {
		return users.ErrNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]users.Model, error) {
	var result []users.Model
	for _, user := range m.users {
		result = append(result, *user)
	}
	return result, nil
}

func (m *mockRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(m.users)), nil
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name        string
		req         users.CreateRequest
		createErr   error
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "successful creation",
			req:     users.CreateRequest{Username: "testuser", Email: "test@example.com", Password: "password123"},
			wantErr: false,
		},
		{
			name:        "username already exists",
			req:         users.CreateRequest{Username: "existing", Email: "test@example.com", Password: "password123"},
			wantErr:     true,
			expectedErr: users.ErrAlreadyExists,
		},
		{
			name:        "email already exists",
			req:         users.CreateRequest{Username: "newuser", Email: "existing@example.com", Password: "password123"},
			wantErr:     true,
			expectedErr: users.ErrAlreadyExists,
		},
		{
			name:      "create error",
			req:       users.CreateRequest{Username: "testuser", Email: "test@example.com", Password: "password123"},
			createErr: errors.New("database error"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				users: map[uint]*users.Model{
					1: {ID: 1, Username: "existing", Email: "existing@example.com"},
				},
				createErr: tt.createErr,
			}
			eventBus := events.NewInMemoryEventBus()
			service := users.NewService(repo, eventBus, nil)

			ctx := context.Background()
			result, err := service.Create(ctx, tt.req)

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
				if result != nil && result.Username != tt.req.Username {
					t.Errorf("expected username %q, got %q", tt.req.Username, result.Username)
				}
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	repo := &mockRepository{
		users: map[uint]*users.Model{
			1: {ID: 1, Username: "testuser", Email: "test@example.com", Password: hashedPassword},
		},
	}
	eventBus := events.NewInMemoryEventBus()
	service := users.NewService(repo, eventBus, nil)

	ctx := context.Background()
	user, err := service.GetByID(ctx, 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user == nil {
		t.Errorf("expected user but got nil")
	}
	if user != nil && user.ID != 1 {
		t.Errorf("expected user ID 1, got %d", user.ID)
	}

	_, err = service.GetByID(ctx, 999)
	if err == nil {
		t.Errorf("expected error for non-existent user")
	}
	if !errors.Is(err, users.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_Update(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	repo := &mockRepository{
		users: map[uint]*users.Model{
			1: {ID: 1, Username: "testuser", Email: "test@example.com", Password: hashedPassword},
		},
	}
	eventBus := events.NewInMemoryEventBus()
	service := users.NewService(repo, eventBus, nil)

	ctx := context.Background()
	newUsername := "updateduser"
	req := users.UpdateRequest{Username: &newUsername}

	user, err := service.Update(ctx, 1, req)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user == nil {
		t.Errorf("expected user but got nil")
	}
	if user != nil && user.Username != "updateduser" {
		t.Errorf("expected username 'updateduser', got %q", user.Username)
	}
}

func TestService_Delete(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	repo := &mockRepository{
		users: map[uint]*users.Model{
			1: {ID: 1, Username: "testuser", Email: "test@example.com", Password: hashedPassword},
		},
	}
	eventBus := events.NewInMemoryEventBus()
	service := users.NewService(repo, eventBus, nil)

	ctx := context.Background()
	err := service.Delete(ctx, 1)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify user is deleted
	_, err = service.GetByID(ctx, 1)
	if err == nil {
		t.Errorf("expected error after deletion")
	}

	// Test delete non-existent user
	err = service.Delete(ctx, 999)
	if err == nil {
		t.Errorf("expected error for non-existent user")
	}
}

func TestService_List(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	repo := &mockRepository{
		users: map[uint]*users.Model{
			1: {ID: 1, Username: "user1", Email: "user1@example.com", Password: hashedPassword},
			2: {ID: 2, Username: "user2", Email: "user2@example.com", Password: hashedPassword},
			3: {ID: 3, Username: "user3", Email: "user3@example.com", Password: hashedPassword},
		},
	}
	eventBus := events.NewInMemoryEventBus()
	service := users.NewService(repo, eventBus, nil)

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
	if result != nil && len(result.Users) != 3 {
		t.Errorf("expected 3 users, got %d", len(result.Users))
	}
}

func TestService_GetByUsername(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	repo := &mockRepository{
		users: map[uint]*users.Model{
			1: {ID: 1, Username: "testuser", Email: "test@example.com", Password: hashedPassword},
		},
	}
	eventBus := events.NewInMemoryEventBus()
	service := users.NewService(repo, eventBus, nil)

	ctx := context.Background()
	user, err := service.GetByUsername(ctx, "testuser")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if user == nil {
		t.Errorf("expected user but got nil")
	}
	if user != nil && user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", user.Username)
	}

	_, err = service.GetByUsername(ctx, "nonexistent")
	if err == nil {
		t.Errorf("expected error for non-existent username")
	}
}

