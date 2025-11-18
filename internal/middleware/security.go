package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

func SecurityHeaders(isDevelopment bool) func(http.Handler) http.Handler {
	secureMiddleware := secure.New(secure.Options{
		IsDevelopment: isDevelopment,

		// Strict Transport Security (HSTS)
		STSSeconds:            31536000, // 1 year
		STSIncludeSubdomains:  true,
		STSPreload:            true,

		// X-Frame-Options
		FrameDeny: true,

		// X-Content-Type-Options
		ContentTypeNosniff: true,

		// X-XSS-Protection
		BrowserXssFilter: true,

		// Referrer Policy
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// Permissions Policy (Feature Policy)
		PermissionsPolicy: "geolocation=(), microphone=(), camera=()",
	})

	return secureMiddleware.Handler
}

func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})

	return c.Handler
}

func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	return httprate.Limit(
		requestsPerMinute,
		time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	)
}

func RequestID() func(http.Handler) http.Handler {
	return middleware.RequestID
}

func RealIP() func(http.Handler) http.Handler {
	return middleware.RealIP
}

func Recoverer() func(http.Handler) http.Handler {
	return middleware.Recoverer
}

func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return middleware.Timeout(timeout)
}

