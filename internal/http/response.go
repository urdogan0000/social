package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/urdogan0000/social/internal/i18n"
	"github.com/urdogan0000/social/internal/logger"
)

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Logger().Error().
			Err(err).
			Int("status", status).
			Msg("Failed to encode JSON response")
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func RespondError(w http.ResponseWriter, r *http.Request, status int, messageID string) {
	message := i18n.T(r, messageID)
	RespondJSON(w, status, map[string]string{
		"error": message,
	})
}

func RespondErrorWithMessage(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{
		"error": message,
	})
}

func GetPaginationParams(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	return limit, offset
}
