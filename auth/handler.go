package auth

import (
	"encoding/json"
	"net/http"

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

// Register godoc
// @Summary Register a new user
// @Description Register a new user and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "Registration request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		httputil.RespondValidationError(w, r, err)
		return
	}

	response, err := h.service.Register(r.Context(), req)
	if err != nil {
		if err == ErrUsernameExists || err == ErrEmailExists {
			logger.Logger().Warn().
				Str("username", req.Username).
				Str("email", req.Email).
				Msg("User registration failed: already exists")
			httputil.RespondError(w, r, http.StatusConflict, "user_already_exists")
			return
		}
		logger.Logger().Error().
			Err(err).
			Str("username", req.Username).
			Str("email", req.Email).
			Msg("Failed to register user")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_register_user")
		return
	}

	logger.Logger().Info().
		Uint("user_id", response.User.ID).
		Str("username", response.User.Username).
		Str("email", response.User.Email).
		Msg("User registered successfully")
	httputil.RespondJSON(w, http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Login with email and password, get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, r, http.StatusBadRequest, "invalid_request_body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		httputil.RespondValidationError(w, r, err)
		return
	}

	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		if err == ErrInvalidCredentials {
			logger.Logger().Warn().
				Str("email", req.Email).
				Msg("Login failed: invalid credentials")
			httputil.RespondError(w, r, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		logger.Logger().Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to login")
		httputil.RespondError(w, r, http.StatusInternalServerError, "failed_to_login")
		return
	}

	logger.Logger().Info().
		Uint("user_id", response.User.ID).
		Str("username", response.User.Username).
		Str("email", response.User.Email).
		Msg("User logged in successfully")
	httputil.RespondJSON(w, http.StatusOK, response)
}

