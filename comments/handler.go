package comments

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	httputil "github.com/urdogan0000/social/internal/http"
	"github.com/urdogan0000/social/internal/logger"
	"github.com/urdogan0000/social/internal/middleware"
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

// CreateComment godoc
// @Summary Create a new comment
// @Description Create a new comment for a post
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param postID path int true "Post ID"
// @Param comment body CreateRequest true "Comment creation request"
// @Success 201 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{postID}/comments [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	// Get postID from URL
	postID, err := strconv.ParseUint(chi.URLParam(r, "postID"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_post_id")
		return
	}
	req.PostID = uint(postID)

	if err := validator.Validate(&req); err != nil {
		httputil.RespondValidationError(w, r, err)
		return
	}

	comment, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		logger.Logger().Error().
			Err(err).
			Uint("user_id", userID).
			Uint("post_id", req.PostID).
			Msg("Failed to create comment")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_create_comment")
		return
	}

	logger.Logger().Info().
		Uint("comment_id", comment.ID).
		Uint("user_id", userID).
		Uint("post_id", req.PostID).
		Msg("Comment created successfully")
	httputil.RespondJSON(w, http.StatusCreated, comment)
}

// GetCommentsByPostID godoc
// @Summary Get comments by post ID
// @Description Get all comments for a specific post with pagination
// @Tags comments
// @Accept json
// @Produce json
// @Param postID path int true "Post ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{postID}/comments [get]
func (h *Handler) GetByPostID(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.ParseUint(chi.URLParam(r, "postID"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_post_id")
		return
	}

	limit, offset := httputil.GetPaginationParams(r)
	result, err := h.service.GetByPostID(r.Context(), uint(postID), limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_get_comments")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, result)
}

// GetComment godoc
// @Summary Get comment by ID
// @Description Get a comment by its ID
// @Tags comments
// @Accept json
// @Produce json
// @Param id path int true "Comment ID"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_comment_id")
		return
	}

	comment, err := h.service.GetByID(r.Context(), uint(id))
	if err != nil {
		if err == ErrNotFound {
			httputil.RespondError(w, r, http.StatusNotFound, "comment_not_found")
			return
		}
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_get_comment")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, comment)
}

// UpdateComment godoc
// @Summary Update comment
// @Description Update an existing comment
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Param comment body UpdateRequest true "Comment update request"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_comment_id")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		httputil.RespondValidationError(w, r, err)
		return
	}

	comment, err := h.service.Update(r.Context(), uint(id), userID, req)
	if err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().Uint("comment_id", uint(id)).Msg("Comment update failed: not found")
			httputil.RespondError(w, r, http.StatusNotFound, "comment_not_found")
			return
		}
		if err == ErrForbidden {
			logger.Logger().Warn().
				Uint("comment_id", uint(id)).
				Uint("user_id", userID).
				Msg("Comment update failed: forbidden")
			httputil.RespondError(w, r, http.StatusForbidden, "forbidden")
			return
		}
		logger.Logger().Error().Err(err).Uint("comment_id", uint(id)).Msg("Failed to update comment")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_update_comment")
		return
	}

	logger.Logger().Info().
		Uint("comment_id", comment.ID).
		Uint("user_id", userID).
		Msg("Comment updated successfully")
	httputil.RespondJSON(w, http.StatusOK, comment)
}

// DeleteComment godoc
// @Summary Delete comment
// @Description Soft delete a comment by ID
// @Tags comments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Comment ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /comments/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_comment_id")
		return
	}

	if err := h.service.Delete(r.Context(), uint(id), userID); err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().Uint("comment_id", uint(id)).Msg("Comment delete failed: not found")
			httputil.RespondError(w, r, http.StatusNotFound, "comment_not_found")
			return
		}
		if err == ErrForbidden {
			logger.Logger().Warn().
				Uint("comment_id", uint(id)).
				Uint("user_id", userID).
				Msg("Comment delete failed: forbidden")
			httputil.RespondError(w, r, http.StatusForbidden, "forbidden")
			return
		}
		logger.Logger().Error().Err(err).Uint("comment_id", uint(id)).Msg("Failed to delete comment")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_delete_comment")
		return
	}

	logger.Logger().Info().Uint("comment_id", uint(id)).Msg("Comment deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

// ListComments godoc
// @Summary List comments
// @Description Get a paginated list of all comments
// @Tags comments
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListResponse
// @Failure 500 {object} map[string]string
// @Router /comments [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := httputil.GetPaginationParams(r)

	result, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_list_comments")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, result)
}
