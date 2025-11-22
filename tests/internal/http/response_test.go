package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	httputil "github.com/urdogan0000/social/internal/http"
	"github.com/urdogan0000/social/internal/i18n"
)

var initOnce sync.Once

func initI18N(t *testing.T) {
	t.Helper()
	initOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get working directory: %v", err)
		}
		root := filepath.Join(wd, "..", "..", "..")
		if err := os.Chdir(root); err != nil {
			t.Fatalf("failed to change directory: %v", err)
		}
		t.Cleanup(func() {
			if err := os.Chdir(wd); err != nil {
				t.Fatalf("failed to restore working directory: %v", err)
			}
		})
		i18n.Init()
	})
}

func TestRespondJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := map[string]string{"status": "ok"}

	httputil.RespondJSON(rr, http.StatusCreated, payload)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", got)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected body status=ok, got %v", body["status"])
	}
}

func TestRespondErrorWithMessage(t *testing.T) {
	rr := httptest.NewRecorder()
	message := "custom error"

	httputil.RespondErrorWithMessage(rr, http.StatusBadRequest, message)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["error"] != message {
		t.Fatalf("expected error message %q, got %q", message, body["error"])
	}
}

func TestRespondError_Localized(t *testing.T) {
	initI18N(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/?lang=tr", nil)

	httputil.RespondError(rr, req, http.StatusNotFound, "user_not_found")

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	expected := "Kullanıcı bulunamadı"
	if body["error"] != expected {
		t.Fatalf("expected localized message %q, got %q", expected, body["error"])
	}
}

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantLimit  int
		wantOffset int
	}{
		{"defaults", "/posts", 20, 0},
		{"valid params", "/posts?limit=10&offset=5", 10, 5},
		{"limit capped", "/posts?limit=500", 100, 0},
		{"invalid numbers", "/posts?limit=-1&offset=-2", 20, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			limit, offset := httputil.GetPaginationParams(req)
			if limit != tt.wantLimit || offset != tt.wantOffset {
				t.Fatalf("expected (limit=%d, offset=%d), got (%d, %d)", tt.wantLimit, tt.wantOffset, limit, offset)
			}
		})
	}
}

