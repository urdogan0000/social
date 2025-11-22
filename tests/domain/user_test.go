package domain_test

import (
	"testing"

	"github.com/urdogan0000/social/internal/domain"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *domain.User
		wantErr bool
	}{
		{
			name:    "valid user",
			user:    &domain.User{Username: "testuser", Email: "test@example.com"},
			wantErr: false,
		},
		{
			name:    "username too short",
			user:    &domain.User{Username: "ab", Email: "test@example.com"},
			wantErr: true,
		},
		{
			name:    "username too long",
			user:    &domain.User{Username: string(make([]byte, 101)), Email: "test@example.com"},
			wantErr: true,
		},
		{
			name:    "empty email",
			user:    &domain.User{Username: "testuser", Email: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	user := &domain.User{}
	
	// Test valid password
	err := user.SetPassword("password123")
	if err != nil {
		t.Errorf("SetPassword() error = %v, want nil", err)
	}
	if len(user.Password) == 0 {
		t.Errorf("SetPassword() password not set")
	}

	// Test short password
	err = user.SetPassword("short")
	if err == nil {
		t.Errorf("SetPassword() expected error for short password")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &domain.User{}
	password := "password123"
	
	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	// Test correct password
	if !user.CheckPassword(password) {
		t.Errorf("CheckPassword() should return true for correct password")
	}

	// Test incorrect password
	if user.CheckPassword("wrongpassword") {
		t.Errorf("CheckPassword() should return false for incorrect password")
	}
}

func TestUser_UpdateUsername(t *testing.T) {
	user := &domain.User{Username: "olduser"}

	// Test valid update
	err := user.UpdateUsername("newuser")
	if err != nil {
		t.Errorf("UpdateUsername() error = %v", err)
	}
	if user.Username != "newuser" {
		t.Errorf("UpdateUsername() username = %q, want 'newuser'", user.Username)
	}

	// Test invalid update
	err = user.UpdateUsername("ab")
	if err == nil {
		t.Errorf("UpdateUsername() expected error for short username")
	}
}

func TestUser_UpdateEmail(t *testing.T) {
	user := &domain.User{Email: "old@example.com"}

	// Test valid update
	err := user.UpdateEmail("new@example.com")
	if err != nil {
		t.Errorf("UpdateEmail() error = %v", err)
	}
	if user.Email != "new@example.com" {
		t.Errorf("UpdateEmail() email = %q, want 'new@example.com'", user.Email)
	}

	// Test invalid update
	err = user.UpdateEmail("")
	if err == nil {
		t.Errorf("UpdateEmail() expected error for empty email")
	}
}

