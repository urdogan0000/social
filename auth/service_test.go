package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/users"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	users         map[uint]*users.Model
	createErr     error
	getByEmailErr error
}

func (m *mockUserRepository) Create(ctx context.Context, user *users.Model) error {
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

func (m *mockUserRepository) GetByID(ctx context.Context, id uint) (*users.Model, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*users.Model, error) {
	for _, user := range m.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*users.Model, error) {
	if m.getByEmailErr != nil {
		return nil, m.getByEmailErr
	}
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepository) Update(ctx context.Context, user *users.Model) error {
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]users.Model, error) {
	return nil, nil
}

func (m *mockUserRepository) Count(ctx context.Context) (int64, error) {
	return 0, nil
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name         string
		req          RegisterRequest
		existingUser *users.Model
		createErr    error
		wantErr      bool
		expectedErr  error
	}{
		{
			name:    "successful registration",
			req:     RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "password123"},
			wantErr: false,
		},
		{
			name:         "username already exists",
			req:          RegisterRequest{Username: "existing", Email: "test@example.com", Password: "password123"},
			existingUser: &users.Model{ID: 1, Username: "existing", Email: "existing@example.com"},
			wantErr:      true,
			expectedErr:  ErrUsernameExists,
		},
		{
			name:         "email already exists",
			req:          RegisterRequest{Username: "newuser", Email: "existing@example.com", Password: "password123"},
			existingUser: &users.Model{ID: 1, Username: "existing", Email: "existing@example.com"},
			wantErr:      true,
			expectedErr:  ErrEmailExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{
				users:     make(map[uint]*users.Model),
				createErr: tt.createErr,
			}
			if tt.existingUser != nil {
				repo.users[tt.existingUser.ID] = tt.existingUser
			}

			service := NewService(repo, "test-secret", 24)

			ctx := context.Background()
			result, err := service.Register(ctx, tt.req)

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
				if result != nil && result.Token == "" {
					t.Errorf("expected token but got empty string")
				}
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name         string
		req          LoginRequest
		existingUser *users.Model
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "successful login",
			req:          LoginRequest{Email: "test@example.com", Password: "password123"},
			existingUser: &users.Model{ID: 1, Email: "test@example.com", Password: hashedPassword},
			wantErr:      false,
		},
		{
			name:        "user not found",
			req:         LoginRequest{Email: "notfound@example.com", Password: "password123"},
			wantErr:     true,
			expectedErr: ErrInvalidCredentials,
		},
		{
			name:         "wrong password",
			req:          LoginRequest{Email: "test@example.com", Password: "wrongpassword"},
			existingUser: &users.Model{ID: 1, Email: "test@example.com", Password: hashedPassword},
			wantErr:      true,
			expectedErr:  ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepository{
				users: make(map[uint]*users.Model),
			}
			if tt.existingUser != nil {
				repo.users[tt.existingUser.ID] = tt.existingUser
			}

			service := NewService(repo, "test-secret", 24)

			ctx := context.Background()
			result, err := service.Login(ctx, tt.req)

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
				if result != nil && result.Token == "" {
					t.Errorf("expected token but got empty string")
				}
			}
		})
	}
}

func TestService_ValidateToken(t *testing.T) {
	service := NewService(&mockUserRepository{}, "test-secret", 24)

	// Generate a valid token
	userID := uint(1)
	email := "test@example.com"
	token, err := service.generateToken(userID, email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Test valid token
	validatedUserID, validatedEmail, err := service.ValidateToken(token)
	if err != nil {
		t.Errorf("unexpected error validating valid token: %v", err)
	}
	if validatedUserID != userID {
		t.Errorf("expected userID %d, got %d", userID, validatedUserID)
	}
	if validatedEmail != email {
		t.Errorf("expected email %q, got %q", email, validatedEmail)
	}

	// Test invalid token
	_, _, err = service.ValidateToken("invalid-token")
	if err == nil {
		t.Errorf("expected error for invalid token")
	}
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}

	// Test empty token
	_, _, err = service.ValidateToken("")
	if err == nil {
		t.Errorf("expected error for empty token")
	}
}
