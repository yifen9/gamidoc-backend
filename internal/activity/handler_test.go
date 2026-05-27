package activity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gamidoc/backend/internal/token"
)

type fakeRecorder struct {
	events []Event
	err    error
}

func (r *fakeRecorder) Record(ctx context.Context, event Event) error {
	if r.err != nil {
		return r.err
	}
	r.events = append(r.events, event)
	return nil
}

func TestRecordFrontendEventAnonymous(t *testing.T) {
	recorder := &fakeRecorder{}
	handler := NewHandler(recorder, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{
		"type": "button_click",
		"sessionId": "session-1",
		"page": "/wizard/step/2",
		"metadata": {"target": "next"}
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.Routes().ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, resp.Code)
	}
	if len(recorder.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(recorder.events))
	}
	event := recorder.events[0]
	if event.Type != "frontend_button_click" {
		t.Fatalf("expected frontend_button_click, got %q", event.Type)
	}
	if event.SessionID != "session-1" || event.Path != "/wizard/step/2" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestRecordFrontendEventWithBearerToken(t *testing.T) {
	recorder := &fakeRecorder{}
	manager := token.NewManager("secret", time.Hour)
	tokenValue, err := manager.Generate("user-1")
	if err != nil {
		t.Fatal(err)
	}
	handler := NewHandler(recorder, manager, nil)

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"type":"page_view"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenValue)
	resp := httptest.NewRecorder()

	handler.Routes().ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, resp.Code)
	}
	if recorder.events[0].UserID != "user-1" {
		t.Fatalf("expected user-1, got %q", recorder.events[0].UserID)
	}
}

func TestRecordFrontendEventRejectsInvalidBearerToken(t *testing.T) {
	recorder := &fakeRecorder{}
	manager := token.NewManager("secret", time.Hour)
	handler := NewHandler(recorder, manager, nil)

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"type":"page_view"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer invalid")
	resp := httptest.NewRecorder()

	handler.Routes().ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, resp.Code)
	}
	if len(recorder.events) != 0 {
		t.Fatalf("expected no events, got %d", len(recorder.events))
	}
}

func TestRecordFrontendEventRejectsInvalidType(t *testing.T) {
	handler := NewHandler(&fakeRecorder{}, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(`{"type":"bad event!"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.Routes().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] == nil {
		t.Fatal("expected error envelope")
	}
}

func TestRecordFrontendEventRejectsOversizedMetadataAfterServerFields(t *testing.T) {
	payloadSize := maxMetadataBytes
	for payloadSize > 0 && !metadataSizeOK(map[string]any{"payload": strings.Repeat("x", payloadSize)}) {
		payloadSize--
	}

	body, err := json.Marshal(map[string]any{
		"type": "page_view",
		"metadata": map[string]any{
			"payload": strings.Repeat("x", payloadSize),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(&fakeRecorder{}, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	handler.Routes().ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}
