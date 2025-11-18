package posts

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post with title, content, user_id and optional tags
// @Tags posts
// @Accept json
// @Produce json
// @Param post body CreateRequest true "Post creation request"
// @Success 201 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts [post]
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

	if err := validator.Validate(&req); err != nil {
		httputil.RespondErrorWithMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	post, err := h.service.Create(r.Context(), userID, req)
	if err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().
				Uint("user_id", userID).
				Str("title", req.Title).
				Msg("Post creation failed: user not found")
			httputil.RespondError(w, r, http.StatusNotFound, "user_not_found")
			return
		}
		logger.Logger().Error().
			Err(err).
			Uint("user_id", userID).
			Str("title", req.Title).
			Msg("Failed to create post")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_create_post")
		return
	}

	logger.Logger().Info().
		Uint("post_id", post.ID).
		Uint("user_id", post.UserID).
		Str("title", post.Title).
		Msg("Post created successfully")
	httputil.RespondJSON(w, http.StatusCreated, post)
}

// GetPost godoc
// @Summary Get post by ID
// @Description Get a post by its ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id} [get]
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_post_id")
		return
	}

	post, err := h.service.GetByID(r.Context(), uint(id))
	if err != nil {
		if err == ErrNotFound {
			httputil.RespondError(w, r, http.StatusNotFound, "post_not_found")
			return
		}
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_get_post")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, post)
}

// UpdatePost godoc
// @Summary Update post
// @Description Update an existing post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param post body UpdateRequest true "Post update request"
// @Success 200 {object} Response
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_post_id")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		httputil.RespondErrorWithMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	post, err := h.service.Update(r.Context(), uint(id), userID, req)
	if err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().Uint("post_id", uint(id)).Msg("Post update failed: not found")
			httputil.RespondError(w, r, http.StatusNotFound, "post_not_found")
			return
		}
		if err == ErrForbidden {
			logger.Logger().Warn().
				Uint("post_id", uint(id)).
				Uint("user_id", userID).
				Msg("Post update failed: forbidden")
			httputil.RespondError(w, r, http.StatusForbidden, "forbidden")
			return
		}
		logger.Logger().Error().Err(err).Uint("post_id", uint(id)).Msg("Failed to update post")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_update_post")
		return
	}

	logger.Logger().Info().
		Uint("post_id", post.ID).
		Uint("user_id", post.UserID).
		Str("title", post.Title).
		Msg("Post updated successfully")
	httputil.RespondJSON(w, http.StatusOK, post)
}

// DeletePost godoc
// @Summary Delete post
// @Description Soft delete a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		httputil.RespondError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_post_id")
		return
	}

	if err := h.service.Delete(r.Context(), uint(id), userID); err != nil {
		if err == ErrNotFound {
			logger.Logger().Warn().Uint("post_id", uint(id)).Msg("Post delete failed: not found")
			httputil.RespondError(w, r, http.StatusNotFound, "post_not_found")
			return
		}
		if err == ErrForbidden {
			logger.Logger().Warn().
				Uint("post_id", uint(id)).
				Uint("user_id", userID).
				Msg("Post delete failed: forbidden")
			httputil.RespondError(w, r, http.StatusForbidden, "forbidden")
			return
		}
		logger.Logger().Error().Err(err).Uint("post_id", uint(id)).Msg("Failed to delete post")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_delete_post")
		return
	}

	logger.Logger().Info().Uint("post_id", uint(id)).Msg("Post deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

// ListPosts godoc
// @Summary List posts
// @Description Get a paginated list of posts
// @Tags posts
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListResponse
// @Failure 500 {object} map[string]string
// @Router /posts [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := httputil.GetPaginationParams(r)

	result, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_list_posts")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, result)
}

// GetPostsByUser godoc
// @Summary Get posts by user ID
// @Description Get all posts created by a specific user
// @Tags posts
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} ListResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{userID}/posts [get]
func (h *Handler) GetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(chi.URLParam(r, "userID"), 10, 32)
	if err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_user_id")
		return
	}

	limit, offset := httputil.GetPaginationParams(r)
	result, err := h.service.GetByUserID(r.Context(), uint(userID), limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_get_user_posts")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, result)
}

// SearchPosts godoc
// @Summary Search posts by title
// @Description Search posts by title (case-insensitive)
// @Tags posts
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/search [get]
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "search_query_required")
		return
	}

	limit, offset := httputil.GetPaginationParams(r)
	posts, err := h.service.SearchByTitle(r.Context(), query, limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_search_posts")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts,
		"query": query,
	})
}

// GetPostsByTags godoc
// @Summary Get posts by tags
// @Description Get posts that contain any of the specified tags
// @Tags posts
// @Accept json
// @Produce json
// @Param tags query string true "Comma-separated tags" example("golang,api,tutorial")
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/tags [get]
func (h *Handler) GetByTags(w http.ResponseWriter, r *http.Request) {
	tagsParam := r.URL.Query().Get("tags")
	if tagsParam == "" {
		httputil.RespondError(w, r, http.StatusBadRequest, "tags_parameter_required")
		return
	}

	tags := strings.Split(tagsParam, ",")
	for i, tag := range tags {
		tags[i] = strings.TrimSpace(tag)
	}

	limit, offset := httputil.GetPaginationParams(r)
	posts, err := h.service.GetByTags(r.Context(), tags, limit, offset)
	if err != nil {
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_get_posts_by_tags")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts,
		"tags":  tags,
	})
}
