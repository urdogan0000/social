package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urdogan0000/social/auth"
	"github.com/urdogan0000/social/internal/middleware"
	"github.com/urdogan0000/social/users"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepoForAuth struct {
	users map[uint]*users.Model
}

func (m *mockUserRepoForAuth) Create(ctx context.Context, user *users.Model) error { return nil }
func (m *mockUserRepoForAuth) GetByID(ctx context.Context, id uint) (*users.Model, error) {
	if user, ok := m.users[id]; ok {
		return user, nil
	}
	return nil, users.ErrNotFound
}
func (m *mockUserRepoForAuth) GetByUsername(ctx context.Context, username string) (*users.Model, error) { return nil, nil }
func (m *mockUserRepoForAuth) GetByEmail(ctx context.Context, email string) (*users.Model, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, users.ErrNotFound
}
func (m *mockUserRepoForAuth) Update(ctx context.Context, user *users.Model) error { return nil }
func (m *mockUserRepoForAuth) Delete(ctx context.Context, id uint) error { return nil }
func (m *mockUserRepoForAuth) List(ctx context.Context, limit, offset int) ([]users.Model, error) { return nil, nil }
func (m *mockUserRepoForAuth) Count(ctx context.Context) (int64, error) { return 0, nil }

func TestAuthMiddleware(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	repo := &mockUserRepoForAuth{
		users: map[uint]*users.Model{
			1: {ID: 1, Email: "test@example.com", Password: hashedPassword},
		},
	}
	authService := auth.NewService(repo, "test-secret", 24)

	// Generate a valid token
	ctx := context.Background()
	loginReq := auth.LoginRequest{Email: "test@example.com", Password: "password123"}
	loginResult, err := authService.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	validToken := loginResult.Token

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantUserID uint
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
			wantUserID: 1,
		},
		{
			name:       "valid token without bearer",
			authHeader: validToken,
			wantStatus: http.StatusOK,
			wantUserID: 1,
		},
		{
			name:       "missing authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.AuthMiddleware(authService)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantStatus == http.StatusOK {
					userID, ok := middleware.GetUserID(r.Context())
					if !ok {
						t.Errorf("GetUserID() should return true")
					}
					if userID != tt.wantUserID {
						t.Errorf("GetUserID() = %d, want %d", userID, tt.wantUserID)
					}
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, uint(123))
	
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		t.Errorf("GetUserID() should return true")
	}
	if userID != 123 {
		t.Errorf("GetUserID() = %d, want 123", userID)
	}

	// Test without user ID
	ctx2 := context.Background()
	_, ok = middleware.GetUserID(ctx2)
	if ok {
		t.Errorf("GetUserID() should return false when user ID not in context")
	}
}

