package users

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httputil "github.com/urdogan0000/social/internal/http"
	"github.com/urdogan0000/social/internal/logger"
	"github.com/urdogan0000/social/internal/validator"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with username, email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateRequest true "User creation request"
// @Success 201 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "validation_failed")
		return
	}

	user, err := h.service.Create(r.Context(), req)
	if err != nil {
		if err == ErrAlreadyExists {
			logger.Logger().Warn().
				Str("username", req.Username).
				Str("email", req.Email).
				Msg("User creation failed: already exists")
			httputil.RespondError(w, r, http.StatusConflict, "user_already_exists")
			return
		}
		logger.Logger().Error().
			Err(err).
			Str("username", req.Username).
			Str("email", req.Email).
			Msg("Failed to create user")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_create_user")
		return
	}

	logger.Logger().Info().
		Uint("user_id", user.ID).
		Str("username", user.Username).
		Str("email", user.Email).
		Msg("User created successfully")
	httputil.RespondJSON(w, http.StatusCreated, user)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_user_id")
		return
	}

	user, err := h.service.GetByID(r.Context(), uint(id))
	if err != nil {
		if err == ErrNotFound {
			logger.Logger().Debug().Uint("user_id", uint(id)).Msg("User not found")
			httputil.RespondError(w, r, http.StatusNotFound, "user_not_found")
			return
		}
		logger.Logger().Error().Err(err).Uint("user_id", uint(id)).Msg("Failed to get user")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_get_user")
		return
	}

	logger.Logger().Debug().Uint("user_id", user.ID).Str("username", user.Username).Msg("User retrieved")
	httputil.RespondJSON(w, http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body UpdateRequest true "User update request"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_user_id")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "validation_failed")
		return
	}

	user, err := h.service.Update(r.Context(), uint(id), req)
	if err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().Uint("user_id", uint(id)).Msg("User update failed: not found")
			httputil.RespondError(w, r, http.StatusNotFound, "user_not_found")
			return
		}
		if err == ErrAlreadyExists {
			logger.Logger().Warn().Uint("user_id", uint(id)).Msg("User update failed: already exists")
			httputil.RespondError(w, r, http.StatusConflict, "user_already_exists")
			return
		}
		logger.Logger().Error().Err(err).Uint("user_id", uint(id)).Msg("Failed to update user")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_update_user")
		return
	}

	logger.Logger().Info().
		Uint("user_id", user.ID).
		Str("username", user.Username).
		Msg("User updated successfully")
	httputil.RespondJSON(w, http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Soft delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_user_id")
		return
	}

	if err := h.service.Delete(r.Context(), uint(id)); err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().Uint("user_id", uint(id)).Msg("User delete failed: not found")
			httputil.RespondError(w, r, http.StatusNotFound, "user_not_found")
			return
		}
		logger.Logger().Error().Err(err).Uint("user_id", uint(id)).Msg("Failed to delete user")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_delete_user")
		return
	}

	logger.Logger().Info().Uint("user_id", uint(id)).Msg("User deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListResponse
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := httputil.GetPaginationParams(r)

	result, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_list_users")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, result)
}

