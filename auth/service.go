package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/users"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	userRepo        users.Repository
	jwtSecret       string
	expirationHours int
}

func NewService(userRepo users.Repository, jwtSecret string, expirationHours int) *Service {
	return &Service{
		userRepo:        userRepo,
		jwtSecret:       jwtSecret,
		expirationHours: expirationHours,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrUserNotFound {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &users.Model{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: UserInfo{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (s *Service) generateToken(userID uint, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Duration(s.expirationHours) * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func (s *Service) ValidateToken(tokenString string) (uint, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return 0, "", ErrInvalidToken
	}

	if !token.Valid {
		return 0, "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, "", ErrInvalidToken
	}

	userID, err := extractUserID(claims)
	if err != nil {
		return 0, "", err
	}

	email, err := extractEmail(claims)
	if err != nil {
		return 0, "", err
	}

	return userID, email, nil
}

func extractUserID(claims jwt.MapClaims) (uint, error) {
	userIDVal, exists := claims["user_id"]
	if !exists {
		return 0, ErrInvalidToken
	}

	userIDFloat, ok := userIDVal.(float64)
	if !ok {
		return 0, ErrInvalidToken
	}

	return uint(userIDFloat), nil
}

func extractEmail(claims jwt.MapClaims) (string, error) {
	emailVal, exists := claims["email"]
	if !exists {
		return "", ErrInvalidToken
	}

	email, ok := emailVal.(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return email, nil
}
