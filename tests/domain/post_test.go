package domain_test

import (
	"testing"

	"github.com/urdogan0000/social/internal/domain"
)

func TestPost_Validate(t *testing.T) {
	tests := []struct {
		name    string
		post    *domain.Post
		wantErr bool
	}{
		{
			name:    "valid post",
			post:    &domain.Post{Title: "Test Title", Content: "Test Content"},
			wantErr: false,
		},
		{
			name:    "empty title",
			post:    &domain.Post{Title: "", Content: "Test Content"},
			wantErr: true,
		},
		{
			name:    "title too long",
			post:    &domain.Post{Title: string(make([]byte, 256)), Content: "Test Content"},
			wantErr: true,
		},
		{
			name:    "empty content",
			post:    &domain.Post{Title: "Test Title", Content: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.post.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPost_CanBeEditedBy(t *testing.T) {
	post := &domain.Post{UserID: domain.UserID(1)}

	// Test owner can edit
	if !post.CanBeEditedBy(domain.UserID(1)) {
		t.Errorf("CanBeEditedBy() should return true for owner")
	}

	// Test other user cannot edit
	if post.CanBeEditedBy(domain.UserID(2)) {
		t.Errorf("CanBeEditedBy() should return false for non-owner")
	}
}

func TestPost_CanBeDeletedBy(t *testing.T) {
	post := &domain.Post{UserID: domain.UserID(1)}

	// Test owner can delete
	if !post.CanBeDeletedBy(domain.UserID(1)) {
		t.Errorf("CanBeDeletedBy() should return true for owner")
	}

	// Test other user cannot delete
	if post.CanBeDeletedBy(domain.UserID(2)) {
		t.Errorf("CanBeDeletedBy() should return false for non-owner")
	}
}

func TestPost_UpdateTitle(t *testing.T) {
	post := &domain.Post{Title: "Old Title"}

	// Test valid update
	err := post.UpdateTitle("New Title")
	if err != nil {
		t.Errorf("UpdateTitle() error = %v", err)
	}
	if post.Title != "New Title" {
		t.Errorf("UpdateTitle() title = %q, want 'New Title'", post.Title)
	}

	// Test invalid update
	err = post.UpdateTitle("")
	if err == nil {
		t.Errorf("UpdateTitle() expected error for empty title")
	}
}

func TestPost_UpdateContent(t *testing.T) {
	post := &domain.Post{Content: "Old Content"}

	// Test valid update
	err := post.UpdateContent("New Content")
	if err != nil {
		t.Errorf("UpdateContent() error = %v", err)
	}
	if post.Content != "New Content" {
		t.Errorf("UpdateContent() content = %q, want 'New Content'", post.Content)
	}

	// Test invalid update
	err = post.UpdateContent("")
	if err == nil {
		t.Errorf("UpdateContent() expected error for empty content")
	}
}

func TestPost_HasTag(t *testing.T) {
	post := &domain.Post{Tags: []string{"golang", "tutorial", "api"}}

	// Test existing tag
	if !post.HasTag("golang") {
		t.Errorf("HasTag() should return true for existing tag")
	}

	// Test non-existing tag
	if post.HasTag("python") {
		t.Errorf("HasTag() should return false for non-existing tag")
	}
}

func TestPost_UpdateTags(t *testing.T) {
	post := &domain.Post{Tags: []string{"old"}}
	newTags := []string{"golang", "tutorial"}

	post.UpdateTags(newTags)

	if len(post.Tags) != 2 {
		t.Errorf("UpdateTags() expected 2 tags, got %d", len(post.Tags))
	}
}

