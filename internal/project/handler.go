package project

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	appmiddleware "github.com/yifen9/gamidoc-backend/internal/http/middleware"
	"github.com/yifen9/gamidoc-backend/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.create)
	r.Get("/", h.list)
	r.Get("/{projectId}", h.get)

	return r
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID := appmiddleware.GetAuthUserID(r.Context())
	if userID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", nil)
		return
	}

	var input CreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	created, err := h.service.Create(r.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidProjectName):
			response.WriteError(w, http.StatusBadRequest, "INVALID_PROJECT_NAME", "Project name is required", map[string]any{"field": "name"})
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	response.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID := appmiddleware.GetAuthUserID(r.Context())
	if userID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", nil)
		return
	}

	projects, err := h.service.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"projects": projects,
		"total":    len(projects),
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID := appmiddleware.GetAuthUserID(r.Context())
	if userID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized", nil)
		return
	}

	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		response.WriteError(w, http.StatusBadRequest, "INVALID_PROJECT_ID", "Invalid project id", nil)
		return
	}

	found, err := h.service.Get(r.Context(), userID, projectID)
	if err != nil {
		switch {
		case errors.Is(err, ErrProjectNotFound):
			response.WriteError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "Project not found", nil)
		case errors.Is(err, ErrForbiddenProject):
			response.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Project does not belong to user", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, found)
}
