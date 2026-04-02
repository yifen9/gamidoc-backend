package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yifen9/gamidoc-backend/internal/auth"
	appmiddleware "github.com/yifen9/gamidoc-backend/internal/http/middleware"
	"github.com/yifen9/gamidoc-backend/internal/http/response"
)

type postgresReadyChecker interface {
	Ready(ctx context.Context) error
}

type redisReadyChecker interface {
	Ready(ctx context.Context) error
}

type Dependencies struct {
	Logger      *slog.Logger
	Postgres    postgresReadyChecker
	Redis       redisReadyChecker
	AuthHandler *auth.Handler
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
		writeJSON(w, http.StatusOK, healthResponse{
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
			writeJSON(w, http.StatusServiceUnavailable, resp)
			return
		}

		writeJSON(w, http.StatusOK, resp)
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, pingResponse{
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
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
