package di

import (
	"context"
	"time"

	"github.com/urdogan0000/social/auth"
	"github.com/urdogan0000/social/internal/config"
	"github.com/urdogan0000/social/internal/db"
	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/posts"
	"github.com/urdogan0000/social/users"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(config.Load),
	fx.Provide(provideDatabase),
	fx.Provide(provideUserRepository),
	fx.Provide(providePostRepository),
	fx.Provide(provideUserService),
	fx.Provide(providePostService),
	fx.Provide(provideUserHandler),
	fx.Provide(providePostHandler),
	fx.Provide(provideAuthService),
	fx.Provide(provideAuthHandler),
)

func provideDatabase(cfg *config.Config) (*gorm.DB, error) {
	gormDB, err := db.NewGORM(cfg.DB.Addr)
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	duration, err := time.ParseDuration(cfg.DB.MaxIdleTime)
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(duration)

	return gormDB, nil
}

func provideUserRepository(db *gorm.DB) users.Repository {
	return users.NewRepository(db)
}

func providePostRepository(db *gorm.DB) posts.Repository {
	return posts.NewRepository(db)
}

func provideUserService(userRepo users.Repository) *users.Service {
	return users.NewService(userRepo)
}

func providePostService(postRepo posts.Repository, userRepo users.Repository) *posts.Service {
	checker := userExistsChecker{repo: userRepo}
	return posts.NewService(postRepo, checker)
}

func provideUserHandler(userService *users.Service) *users.Handler {
	return users.NewHandler(userService)
}

func providePostHandler(postService *posts.Service) *posts.Handler {
	return posts.NewHandler(postService)
}

func provideAuthService(userRepo users.Repository, cfg *config.Config) *auth.Service {
	return auth.NewService(userRepo, cfg.JWT.SecretKey, cfg.JWT.ExpirationHours)
}

func provideAuthHandler(authService *auth.Service) *auth.Handler {
	return auth.NewHandler(authService)
}

type userExistsChecker struct {
	repo users.Repository
}

func (c userExistsChecker) UserExists(ctx context.Context, userID domain.UserID) (bool, error) {
	_, err := c.repo.GetByID(ctx, uint(userID))
	if err == domain.ErrUserNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
