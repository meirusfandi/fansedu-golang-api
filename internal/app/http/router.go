package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/handlers"
	"github.com/meirusfandi/fansedu-golang-api/internal/app/http/middleware"
)

type Deps struct {
	DB        *pgxpool.Pool
	JWTSecret []byte
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID())
	r.Use(middleware.Recover())
	r.Use(chimw.RealIP)
	r.Use(middleware.Logger())

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", handlers.Health())

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handlers.AuthRegister(deps.DB, deps.JWTSecret))
			r.Post("/login", handlers.AuthLogin(deps.DB, deps.JWTSecret))
		})
	})

	return r
}

