package pdf

import (
	"net/http"
	"strings"

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

func (h *Handler) ProjectGenerate(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.service.GenerateProjectPDF(r.Context(), userID, projectID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "PDF_GENERATION_FAILED", "PDF generation failed", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"pdfUrl": result.URL,
	})
}

func (h *Handler) SessionGenerate(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "sessionId")
	if sessionID == "" {
		response.WriteError(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Invalid session id", nil)
		return
	}

	result, err := h.service.GenerateSessionPDF(r.Context(), sessionID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "PDF_GENERATION_FAILED", "PDF generation failed", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"pdfUrl": result.URL,
	})
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "*")
	key = strings.TrimLeft(key, "/")

	data, err := h.service.Download(r.Context(), key)
	if err != nil {
		response.WriteError(w, http.StatusNotFound, "PDF_NOT_FOUND", "PDF not found", nil)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="evaluation-plan.pdf"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
