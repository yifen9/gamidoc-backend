package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yifen9/gamidoc-backend/internal/auth"
	appmiddleware "github.com/yifen9/gamidoc-backend/internal/http/middleware"
	"github.com/yifen9/gamidoc-backend/internal/http/response"
	"github.com/yifen9/gamidoc-backend/internal/pdf"
	"github.com/yifen9/gamidoc-backend/internal/project"
	"github.com/yifen9/gamidoc-backend/internal/session"
)

type postgresReadyChecker interface {
	Ready(ctx context.Context) error
}

type redisReadyChecker interface {
	Ready(ctx context.Context) error
}

type Dependencies struct {
	Logger         *slog.Logger
	Postgres       postgresReadyChecker
	Redis          redisReadyChecker
	AuthHandler    *auth.Handler
	ProjectHandler *project.Handler
	SessionHandler *session.Handler
	PDFHandler     *pdf.Handler
}

type healthResponse struct {
	Status string `json:"status"`
}

type readyResponse struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

type pingResponse struct {
	Message string `json:"message"`
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Recovery(deps.Logger))
	r.Use(appmiddleware.Logging(deps.Logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJSON(w, http.StatusOK, healthResponse{
			Status: "ok",
		})
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		resp := readyResponse{
			Status:   "ok",
			Postgres: "ok",
			Redis:    "ok",
		}

		if deps.Postgres == nil || deps.Postgres.Ready(ctx) != nil {
			resp.Status = "error"
			resp.Postgres = "error"
		}

		if deps.Redis == nil || deps.Redis.Ready(ctx) != nil {
			resp.Status = "error"
			resp.Redis = "error"
		}

		if resp.Status != "ok" {
			response.WriteJSON(w, http.StatusServiceUnavailable, resp)
			return
		}

		response.WriteJSON(w, http.StatusOK, resp)
	})

	if deps.PDFHandler != nil {
		r.Get("/files/pdfs/*", deps.PDFHandler.Download)
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			response.WriteJSON(w, http.StatusOK, pingResponse{
				Message: "pong",
			})
		})

		r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
			panic("panic route triggered")
		})

		r.Get("/error", func(w http.ResponseWriter, r *http.Request) {
			response.WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "Bad request", map[string]any{
				"path": r.URL.Path,
			})
		})

		if deps.AuthHandler != nil {
			r.Mount("/auth", deps.AuthHandler.Routes())
		}

		if deps.ProjectHandler != nil {
			r.With(appmiddleware.RequireAuth).Mount("/projects", deps.ProjectHandler.Routes())
		}

		if deps.SessionHandler != nil {
			r.Mount("/sessions", deps.SessionHandler.Routes())
		}

		if deps.PDFHandler != nil {
			r.With(appmiddleware.RequireAuth).Post("/projects/{projectId}/generate-pdf", deps.PDFHandler.ProjectGenerate)
			r.Post("/sessions/{sessionId}/generate-pdf", deps.PDFHandler.SessionGenerate)
		}
	})

	return r
}
