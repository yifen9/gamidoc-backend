package session

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/yifen9/gamidoc-backend/internal/http/response"
	"github.com/yifen9/gamidoc-backend/internal/recommendation"
	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/create", h.create)
	r.Get("/{sessionId}", h.get)
	r.Put("/{sessionId}/wizard/step/{stepNumber}", h.saveStep)
	r.Post("/{sessionId}/wizard/recommendations", h.recommend)

	return r
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	created, err := h.service.Create(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		response.WriteError(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Invalid session id", nil)
		return
	}

	found, err := h.service.Get(r.Context(), sessionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			response.WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found or expired", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, found)
}

func (h *Handler) saveStep(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		response.WriteError(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Invalid session id", nil)
		return
	}

	stepValue := chi.URLParam(r, "stepNumber")
	stepNumber, err := strconv.Atoi(stepValue)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_STEP_NUMBER", "Invalid step number", nil)
		return
	}

	var input struct {
		StepData json.RawMessage `json:"stepData"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	updated, err := h.service.SaveStep(r.Context(), sessionID, stepNumber, input.StepData)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			response.WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found or expired", nil)
		case errors.Is(err, wizard.ErrInvalidStepNumber):
			response.WriteError(w, http.StatusBadRequest, "INVALID_STEP_NUMBER", "Invalid step number", nil)
		case errors.Is(err, wizard.ErrInvalidStepData):
			response.WriteError(w, http.StatusBadRequest, "INVALID_STEP_DATA", "Invalid step data", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"sessionId":    updated.ID,
		"stepNumber":   stepNumber,
		"stepData":     updated.Wizard.Steps[strconv.Itoa(stepNumber)],
		"wizardStatus": updated.Wizard,
		"createdAt":    updated.CreatedAt,
		"expiresAt":    updated.ExpiresAt,
	})
}

func (h *Handler) recommend(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		response.WriteError(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Invalid session id", nil)
		return
	}

	var input struct {
		ForStep int `json:"forStep"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	result, err := h.service.Recommend(r.Context(), sessionID, input.ForStep)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			response.WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found or expired", nil)
		case errors.Is(err, recommendation.ErrInvalidRecommendationStep):
			response.WriteError(w, http.StatusBadRequest, "INVALID_RECOMMENDATION_STEP", "Invalid recommendation step", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, result)
}
