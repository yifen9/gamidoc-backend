package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestHealthRoute(t *testing.T) {
	router := NewRouter(Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestReadyRouteOK(t *testing.T) {
	router := NewRouter(Dependencies{
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
