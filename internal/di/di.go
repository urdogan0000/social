package di

import (
	"context"
	"errors"
	"time"

	"github.com/urdogan0000/social/auth"
	"github.com/urdogan0000/social/internal/config"
	"github.com/urdogan0000/social/internal/db"
	"github.com/urdogan0000/social/internal/domain"
	"github.com/urdogan0000/social/internal/events"
	"github.com/urdogan0000/social/posts"
	"github.com/urdogan0000/social/users"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(config.Load),
	fx.Provide(provideDatabase),
	fx.Provide(provideTransactionManager),
	fx.Provide(provideEventBus),
	fx.Provide(provideUserRepository),
	fx.Provide(providePostRepository),
	fx.Provide(provideDomainUserRepository),
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

func provideTransactionManager(gormDB *gorm.DB) db.TransactionManager {
	return db.NewTransactionManager(gormDB)
}

func provideEventBus(cfg *config.Config) (events.EventBus, error) {
	switch cfg.EventBus.Type {
	case "kafka":
		// TODO: Kafka implementasyonu eklenecek
		// return events.NewKafkaEventBus(events.KafkaConfig{
		//     Brokers: cfg.EventBus.Kafka.Brokers,
		//     TopicPrefix: cfg.EventBus.Kafka.TopicPrefix,
		// })
		return events.NewInMemoryEventBus(), nil
	case "nats":
		// TODO: NATS implementasyonu eklenecek
		// return events.NewNATSEventBus(cfg.EventBus.NATS.URL)
		return events.NewInMemoryEventBus(), nil
	default:
		return events.NewInMemoryEventBus(), nil
	}
}

func provideUserRepository(db *gorm.DB) users.Repository {
	return users.NewRepository(db)
}

func providePostRepository(db *gorm.DB) posts.Repository {
	return posts.NewRepository(db)
}

// provideDomainUserRepository provides domain.UserRepository interface
// This allows other modules to depend on domain interface instead of concrete implementation
func provideDomainUserRepository(userRepo users.Repository) domain.UserRepository {
	return &domainUserRepositoryAdapter{repo: userRepo}
}

func provideUserService(
	userRepo users.Repository,
	eventBus events.EventBus,
	transactionMgr db.TransactionManager,
) *users.Service {
	return users.NewService(userRepo, eventBus, transactionMgr)
}

func providePostService(
	postRepo posts.Repository,
	userRepo domain.UserRepository,
	eventBus events.EventBus,
	transactionMgr db.TransactionManager,
) *posts.Service {
	return posts.NewService(postRepo, userRepo, eventBus, transactionMgr)
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

// domainUserRepositoryAdapter adapts users.Repository to domain.UserRepository
type domainUserRepositoryAdapter struct {
	repo users.Repository
}

func (a *domainUserRepositoryAdapter) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	model, err := a.repo.GetByID(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:       domain.UserID(model.ID),
		Username: model.Username,
		Email:    model.Email,
		Password: model.Password,
	}, nil
}

func (a *domainUserRepositoryAdapter) Exists(ctx context.Context, id domain.UserID) (bool, error) {
	_, err := a.repo.GetByID(ctx, uint(id))
	if errors.Is(err, users.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *domainUserRepositoryAdapter) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	model, err := a.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:       domain.UserID(model.ID),
		Username: model.Username,
		Email:    model.Email,
		Password: model.Password,
	}, nil
}

func (a *domainUserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	model, err := a.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:       domain.UserID(model.ID),
		Username: model.Username,
		Email:    model.Email,
		Password: model.Password,
	}, nil
}
