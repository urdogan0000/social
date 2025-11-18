package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/urdogan0000/social/docs/swagger"
	"github.com/urdogan0000/social/auth"
	"github.com/urdogan0000/social/internal/config"
	"github.com/urdogan0000/social/internal/middleware"
	"github.com/urdogan0000/social/posts"
	"github.com/urdogan0000/social/users"
)

type Application struct {
	Config      config.Config
	UserHandler *users.Handler
	PostHandler *posts.Handler
	AuthHandler *auth.Handler
}

func (app *Application) Mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.SecurityHeaders(app.Config.Server.IsDevelopment))
	if app.Config.Server.EnableCORS {
		r.Use(middleware.CORS(app.Config.Server.AllowedOrigins))
	}

	r.Use(middleware.RequestID())
	r.Use(middleware.RealIP())
	r.Use(middleware.Recoverer())
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RateLimit(app.Config.Server.RateLimitRPM))

	r.Route("/v1", func(r chi.Router) {
		swaggerURL := "http://localhost" + app.Config.Server.Addr + "/v1/swagger/doc.json"
		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL(swaggerURL),
			httpSwagger.DeepLinking(true),
		))
		r.Get("/health", app.healthCheckHandler)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", app.AuthHandler.Register)
			r.Post("/login", app.AuthHandler.Login)
		})

		r.Route("/users", func(r chi.Router) {
			r.Post("/", app.UserHandler.Create)
			r.Get("/", app.UserHandler.List)
			r.Get("/{id}", app.UserHandler.Get)
			r.Put("/{id}", app.UserHandler.Update)
			r.Delete("/{id}", app.UserHandler.Delete)
			r.Get("/{userID}/posts", app.PostHandler.GetByUser)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.PostHandler.Create)
			r.Get("/", app.PostHandler.List)
			r.Get("/search", app.PostHandler.Search)
			r.Get("/tags", app.PostHandler.GetByTags)
			r.Get("/{id}", app.PostHandler.Get)
			r.Put("/{id}", app.PostHandler.Update)
			r.Delete("/{id}", app.PostHandler.Delete)
		})
	})

	return r
}

func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
