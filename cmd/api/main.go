package main

// @title Social Network API
// @version 1.0
// @description Social Network API with User and Post management
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@example.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8081
// @BasePath /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token. You can enter just the token (e.g., "eyJhbGci...") or with "Bearer " prefix (e.g., "Bearer eyJhbGci...")
import (
	"context"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/urdogan0000/social/auth"
	"github.com/urdogan0000/social/comments"
	"github.com/urdogan0000/social/internal/api"
	"github.com/urdogan0000/social/internal/config"
	"github.com/urdogan0000/social/internal/di"
	"github.com/urdogan0000/social/internal/i18n"
	"github.com/urdogan0000/social/internal/logger"
	"github.com/urdogan0000/social/posts"
	"github.com/urdogan0000/social/users"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	logger.Init("info")
	i18n.Init()
	fx.New(
		di.Module,
		fx.Invoke(registerHooks),
		fx.Invoke(registerRoutes),
	).Run()
}

func registerHooks(
	lc fx.Lifecycle,
	db *gorm.DB,
	cfg *config.Config,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cfg.Server.IsDevelopment {
				if err := runMigrations(db); err != nil {
					return err
				}
				logger.Logger().Info().Msg("Database migrations completed (development mode)")
			} else {
				logger.Logger().Info().Msg("Skipping auto-migration in production. Use 'go run cmd/migrate/main.go' to run migrations manually.")
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			sqlDB, err := db.DB()
			if err != nil {
				return err
			}
			return sqlDB.Close()
		},
	})
}

func registerRoutes(
	lc fx.Lifecycle,
	userHandler *users.Handler,
	postHandler *posts.Handler,
	commentHandler *comments.Handler,
	authHandler *auth.Handler,
	authService *auth.Service,
	cfg *config.Config,
) {
	app := &api.Application{
		Config:         *cfg,
		UserHandler:    userHandler,
		PostHandler:    postHandler,
		CommentHandler: commentHandler,
		AuthHandler:    authHandler,
		AuthService:    authService,
	}

	var srv *http.Server

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mux := app.Mount()
			srv = &http.Server{
				Addr:    cfg.Server.Addr,
				Handler: mux,
			}

			go func() {
				logger.Logger().Info().Str("addr", cfg.Server.Addr).Msg("Server starting")
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Logger().Fatal().Err(err).Msg("Server failed")
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if srv != nil {
				return srv.Shutdown(ctx)
			}
			return nil
		},
	})
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&users.Model{},
		&posts.Model{},
		&comments.Model{},
	)
}
