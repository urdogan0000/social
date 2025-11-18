package auth

import (
	"encoding/json"
	"net/http"

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
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Register(r.Context(), req)
	if err != nil {
		if err.Error() == "username already exists" || err.Error() == "email already exists" {
			logger.Logger().Warn().
				Str("username", req.Username).
				Str("email", req.Email).
				Msg("User registration failed: already exists")
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		logger.Logger().Error().
			Err(err).
			Str("username", req.Username).
			Str("email", req.Email).
			Msg("Failed to register user")
		respondError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	logger.Logger().Info().
		Uint("user_id", response.User.ID).
		Str("username", response.User.Username).
		Str("email", response.User.Email).
		Msg("User registered successfully")
	respondJSON(w, http.StatusCreated, response)
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
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		if err.Error() == "invalid email or password" {
			logger.Logger().Warn().
				Str("email", req.Email).
				Msg("Login failed: invalid credentials")
			respondError(w, http.StatusUnauthorized, err.Error())
			return
		}
		logger.Logger().Error().
			Err(err).
			Str("email", req.Email).
			Msg("Failed to login")
		respondError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	logger.Logger().Info().
		Uint("user_id", response.User.ID).
		Str("username", response.User.Username).
		Str("email", response.User.Email).
		Msg("User logged in successfully")
	respondJSON(w, http.StatusOK, response)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}

