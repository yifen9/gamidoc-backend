package activity

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gamidoc/backend/internal/http/response"
	"github.com/gamidoc/backend/internal/token"
	"github.com/go-chi/chi/v5"
)

const maxMetadataBytes = 16 * 1024

type Handler struct {
	recorder  Recorder
	tokens    *token.Manager
	blacklist *token.Blacklist
}

type RecordFrontendEventInput struct {
	Type      string         `json:"type"`
	SessionID string         `json:"sessionId,omitempty"`
	ProjectID string         `json:"projectId,omitempty"`
	Page      string         `json:"page,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func NewHandler(recorder Recorder, tokens *token.Manager, blacklist *token.Blacklist) *Handler {
	return &Handler{
		recorder:  recorder,
		tokens:    tokens,
		blacklist: blacklist,
	}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/events", h.record)
	return r
}

func (h *Handler) record(w http.ResponseWriter, r *http.Request) {
	if h.recorder == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "ACTIVITY_RECORDER_UNAVAILABLE", "Activity recorder is not configured", nil)
		return
	}

	var input RecordFrontendEventInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	eventType, err := normalizeFrontendEventType(input.Type)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_ACTIVITY_EVENT", "Invalid activity event type", map[string]any{"field": "type"})
		return
	}

	if len(input.SessionID) > 128 || len(input.ProjectID) > 128 || len(input.Page) > 512 {
		response.WriteError(w, http.StatusBadRequest, "INVALID_ACTIVITY_EVENT", "Invalid activity event fields", nil)
		return
	}

	metadata := input.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["user_agent"] = r.UserAgent()
	metadata["remote_addr"] = r.RemoteAddr
	if !metadataSizeOK(metadata) {
		response.WriteError(w, http.StatusBadRequest, "INVALID_ACTIVITY_EVENT", "Activity metadata is too large", map[string]any{"field": "metadata"})
		return
	}

	userID, ok := h.optionalUserID(r)
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid bearer token", nil)
		return
	}

	page := strings.TrimSpace(input.Page)
	if page == "" {
		page = r.URL.Path
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()
	if err := h.recorder.Record(ctx, Event{
		Type:       eventType,
		UserID:     userID,
		SessionID:  strings.TrimSpace(input.SessionID),
		ProjectID:  strings.TrimSpace(input.ProjectID),
		Method:     "FRONTEND",
		Path:       page,
		StatusCode: http.StatusAccepted,
		Metadata:   metadata,
	}); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "ACTIVITY_RECORD_FAILED", "Activity event could not be recorded", nil)
		return
	}

	response.WriteJSON(w, http.StatusAccepted, map[string]any{"status": "accepted"})
}

func (h *Handler) optionalUserID(r *http.Request) (string, bool) {
	header := strings.TrimSpace(r.Header.Get("Authorization"))
	if header == "" {
		return "", true
	}
	if h.tokens == nil {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}

	claims, err := h.tokens.Parse(parts[1])
	if err != nil {
		return "", false
	}
	if h.blacklist != nil && claims.ID != "" {
		revoked, err := h.blacklist.IsBlacklisted(r.Context(), claims.ID)
		if err == nil && revoked {
			return "", false
		}
	}

	return claims.UserID, true
}

func normalizeFrontendEventType(value string) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, EventFrontendPrefix)
	if value == "" || len(value) > 64 {
		return "", errInvalidFrontendEventType{}
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' {
			continue
		}
		return "", errInvalidFrontendEventType{}
	}
	return EventFrontendPrefix + value, nil
}

func metadataSizeOK(metadata map[string]any) bool {
	data, err := json.Marshal(metadata)
	return err == nil && len(data) <= maxMetadataBytes
}

type errInvalidFrontendEventType struct{}

func (errInvalidFrontendEventType) Error() string {
	return "invalid frontend event type"
}
