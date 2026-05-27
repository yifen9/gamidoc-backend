package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gamidoc/backend/internal/activity"
)

func TestFrontendActivityRoute(t *testing.T) {
	handler := activity.NewHandler(activity.NoopRecorder{}, testTokenManager(), nil)
	router := NewRouter(Dependencies{
		Logger:          testLogger(),
		TokenManager:    testTokenManager(),
		ActivityHandler: handler.Routes(),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/activity/events", strings.NewReader(`{"type":"page_view","page":"/wizard"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}
