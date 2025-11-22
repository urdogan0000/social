package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	appi18n "github.com/urdogan0000/social/internal/i18n"
	"github.com/urdogan0000/social/internal/logger"
	"github.com/urdogan0000/social/internal/validator"
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
	message := appi18n.T(r, messageID)
	RespondJSON(w, status, map[string]string{
		"error": message,
	})
}

func RespondErrorWithMessage(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{
		"error": message,
	})
}

// RespondValidationError responds with i18n localized validation errors
func RespondValidationError(w http.ResponseWriter, r *http.Request, validationErr error) {
	// Check if it's our structured validation error
	if ve, ok := validationErr.(validator.ValidationErrors); ok {
		var localizedErrors []string
		localizer := appi18n.GetLocalizer(r)

		for _, err := range ve {
			var message string
			switch err.Tag {
			case "required":
				msg, _ := localizer.Localize(&i18n.LocalizeConfig{
					MessageID: "field_required",
				})
				message = fmt.Sprintf("%s: %s", err.Field, msg)
			case "email":
				msg, _ := localizer.Localize(&i18n.LocalizeConfig{
					MessageID: "field_email",
				})
				message = fmt.Sprintf("%s: %s", err.Field, msg)
			case "min":
				msg, _ := localizer.Localize(&i18n.LocalizeConfig{
					MessageID:    "field_min",
					TemplateData: map[string]string{"Min": err.Param},
				})
				message = fmt.Sprintf("%s: %s", err.Field, msg)
			case "max":
				msg, _ := localizer.Localize(&i18n.LocalizeConfig{
					MessageID:    "field_max",
					TemplateData: map[string]string{"Max": err.Param},
				})
				message = fmt.Sprintf("%s: %s", err.Field, msg)
			default:
				// Fallback to original message
				message = fmt.Sprintf("%s: %s", err.Field, err.Message)
			}
			localizedErrors = append(localizedErrors, message)
		}

		errorsStr := strings.Join(localizedErrors, ", ")
		msg, _ := localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    "validation_failed",
			TemplateData: map[string]string{"Errors": errorsStr},
		})

		RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": msg,
		})
		return
	}

	// Fallback for non-structured errors (parse the error message)
	errorStr := validationErr.Error()
	if strings.Contains(errorStr, "validation failed:") {
		// Try to extract and localize
		msg, _ := appi18n.GetLocalizer(r).Localize(&i18n.LocalizeConfig{
			MessageID:    "validation_failed",
			TemplateData: map[string]string{"Errors": errorStr},
		})
		RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": msg,
		})
		return
	}

	// Last resort: return as-is
	RespondErrorWithMessage(w, http.StatusBadRequest, errorStr)
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
