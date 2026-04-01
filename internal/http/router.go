package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type postgresReadyChecker interface {
	Ready(ctx context.Context) error
}

type redisReadyChecker interface {
	Ready(ctx context.Context) error
}

type Dependencies struct {
	Postgres postgresReadyChecker
	Redis    redisReadyChecker
}

type healthResponse struct {
	Status string `json:"status"`
}

type readyResponse struct {
	Status   string `json:"status"`
	Postgres string `json:"postgres"`
	Redis    string `json:"redis"`
}

func NewRouter(deps Dependencies) http.Handler {
	r := chi.NewRouter()

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

	return r
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
