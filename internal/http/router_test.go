package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type fakePostgres struct {
	readyErr error
}

func (f *fakePostgres) Ready(ctx context.Context) error {
	return f.readyErr
}

type fakeRedis struct {
	readyErr error
}

func (f *fakeRedis) Ready(ctx context.Context) error {
	return f.readyErr
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func TestHealthRoute(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger: testLogger(),
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header to be set")
	}
}

func TestReadyRouteOK(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger:   testLogger(),
		Postgres: &fakePostgres{},
		Redis:    &fakeRedis{},
	})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestReadyRouteFail(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger:   testLogger(),
		Postgres: &fakePostgres{readyErr: errors.New("pg down")},
		Redis:    &fakeRedis{},
	})

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
	}
}

func TestAPIV1Ping(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger: testLogger(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestAPIV1Error(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger: testLogger(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/error", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger: testLogger(),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}
