package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/urdogan0000/social/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func AuthMiddleware(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "authorization header required")
				return
			}

			var token string
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				token = parts[1]
			} else if len(parts) == 1 {
				token = parts[0]
			} else {
				respondError(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			if token == "" {
				respondError(w, http.StatusUnauthorized, "token is required")
				return
			}

			userID, _, err := authService.ValidateToken(token)
			if err != nil {
				respondError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDKey).(uint)
	return userID, ok
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	}); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
