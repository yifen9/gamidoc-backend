package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yifen9/gamidoc-backend/internal/auth"
	appmiddleware "github.com/yifen9/gamidoc-backend/internal/http/middleware"
	"github.com/yifen9/gamidoc-backend/internal/project"
	"github.com/yifen9/gamidoc-backend/internal/recommendation"
	"github.com/yifen9/gamidoc-backend/internal/session"
	"github.com/yifen9/gamidoc-backend/internal/token"
	"github.com/yifen9/gamidoc-backend/internal/user"
	"github.com/yifen9/gamidoc-backend/internal/wizard"
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

type fakeUserRepository struct {
	usersByEmail map[string]user.User
	usersByID    map[string]user.User
}

func (r *fakeUserRepository) Create(ctx context.Context, input user.User) (user.User, error) {
	input.CreatedAt = time.Now()
	if r.usersByEmail == nil {
		r.usersByEmail = map[string]user.User{}
	}
	if r.usersByID == nil {
		r.usersByID = map[string]user.User{}
	}
	r.usersByEmail[input.Email] = input
	r.usersByID[input.ID] = input
	return input, nil
}

func (r *fakeUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	u, ok := r.usersByEmail[email]
	if !ok {
		return user.User{}, errors.New("not found")
	}
	return u, nil
}

func (r *fakeUserRepository) FindByID(ctx context.Context, id string) (user.User, error) {
	u, ok := r.usersByID[id]
	if !ok {
		return user.User{}, errors.New("not found")
	}
	return u, nil
}

type fakeProjectRepository struct {
	items []project.Project
	byID  map[string]project.Project
}

func (r *fakeProjectRepository) Create(ctx context.Context, input project.Project) (project.Project, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = input.CreatedAt
	if r.byID == nil {
		r.byID = map[string]project.Project{}
	}
	r.items = append(r.items, input)
	r.byID[input.ID] = input
	return input, nil
}

func (r *fakeProjectRepository) ListByUserID(ctx context.Context, userID string) ([]project.Project, error) {
	var result []project.Project
	for _, item := range r.items {
		if item.UserID == userID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeProjectRepository) FindByID(ctx context.Context, id string) (project.Project, error) {
	item, ok := r.byID[id]
	if !ok {
		return project.Project{}, project.ErrProjectNotFound
	}
	return item, nil
}

func (r *fakeProjectRepository) UpdateWizard(ctx context.Context, projectID string, status wizard.Status) (project.Project, error) {
	item, ok := r.byID[projectID]
	if !ok {
		return project.Project{}, project.ErrProjectNotFound
	}
	item.Wizard = status
	item.UpdatedAt = time.Now()
	r.byID[projectID] = item
	for i := range r.items {
		if r.items[i].ID == projectID {
			r.items[i] = item
		}
	}
	return item, nil
}

type fakeSessionRepository struct {
	byID map[string]session.Session
}

func (r *fakeSessionRepository) Create(ctx context.Context, input session.Session) (session.Session, error) {
	if r.byID == nil {
		r.byID = map[string]session.Session{}
	}
	r.byID[input.ID] = input
	return input, nil
}

func (r *fakeSessionRepository) FindByID(ctx context.Context, id string) (session.Session, error) {
	item, ok := r.byID[id]
	if !ok {
		return session.Session{}, session.ErrSessionNotFound
	}
	return item, nil
}

func (r *fakeSessionRepository) UpdateWizard(ctx context.Context, id string, status wizard.Status) (session.Session, error) {
	item, ok := r.byID[id]
	if !ok {
		return session.Session{}, session.ErrSessionNotFound
	}
	item.Wizard = status
	r.byID[id] = item
	return item, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func testRecommendationService() *recommendation.Service {
	engine := recommendation.NewEngine(recommendation.LoadDefaultRules())
	return recommendation.NewService(engine)
}

func testAuthHandler() *auth.Handler {
	repo := &fakeUserRepository{
		usersByEmail: map[string]user.User{},
		usersByID:    map[string]user.User{},
	}
	manager := token.NewManager("secret", time.Hour)
	appmiddleware.SetTokenManager(manager)
	service := auth.NewService(repo, manager)
	return auth.NewHandler(service)
}

func testProjectHandler() *project.Handler {
	repo := &fakeProjectRepository{
		items: []project.Project{},
		byID:  map[string]project.Project{},
	}
	service := project.NewService(repo, wizard.NewService(), testRecommendationService())
	return project.NewHandler(service)
}

func testSessionHandler() *session.Handler {
	repo := &fakeSessionRepository{
		byID: map[string]session.Session{},
	}
	service := session.NewService(repo, 48*time.Hour, wizard.NewService(), testRecommendationService())
	return session.NewHandler(service)
}

func authToken() string {
	manager := token.NewManager("secret", time.Hour)
	value, _ := manager.Generate("user-1", "test@example.com")
	appmiddleware.SetTokenManager(manager)
	return value
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

func TestRegisterRoute(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger:      testLogger(),
		AuthHandler: testAuthHandler(),
	})

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestCreateSessionRoute(t *testing.T) {
	router := NewRouter(Dependencies{
		Logger:         testLogger(),
		SessionHandler: testSessionHandler(),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/create", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestCreateProjectRoute(t *testing.T) {
	tokenValue := authToken()

	router := NewRouter(Dependencies{
		Logger:         testLogger(),
		ProjectHandler: testProjectHandler(),
	})

	body := `{"name":"My Project","description":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenValue)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestSaveSessionStepRoute(t *testing.T) {
	handler := testSessionHandler()
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/create", nil)
	createRec := httptest.NewRecorder()

	router := NewRouter(Dependencies{
		Logger:         testLogger(),
		SessionHandler: handler,
	})

	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var created session.Session
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	body := `{"stepData":{"evaluationGoals":["Usability"]}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/sessions/"+created.ID+"/wizard/step/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestSaveProjectStepRoute(t *testing.T) {
	tokenValue := authToken()
	handler := testProjectHandler()

	router := NewRouter(Dependencies{
		Logger:         testLogger(),
		ProjectHandler: handler,
	})

	createBody := `{"name":"My Project","description":"Test"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+tokenValue)
	createRec := httptest.NewRecorder()

	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var created project.Project
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	body := `{"stepData":{"evaluationGoals":["Usability"]}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+created.ID+"/wizard/step/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenValue)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRecommendSessionRoute(t *testing.T) {
	handler := testSessionHandler()

	router := NewRouter(Dependencies{
		Logger:         testLogger(),
		SessionHandler: handler,
	})

	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/create", nil)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	var created session.Session
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	saveBody := `{"stepData":{"evaluationGoals":["Usability & Playability"]}}`
	saveReq := httptest.NewRequest(http.MethodPut, "/api/v1/sessions/"+created.ID+"/wizard/step/1", strings.NewReader(saveBody))
	saveReq.Header.Set("Content-Type", "application/json")
	saveRec := httptest.NewRecorder()
	router.ServeHTTP(saveRec, saveReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/"+created.ID+"/wizard/recommendations", strings.NewReader(`{"forStep":2}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRecommendProjectRoute(t *testing.T) {
	tokenValue := authToken()
	handler := testProjectHandler()

	router := NewRouter(Dependencies{
		Logger:         testLogger(),
		ProjectHandler: handler,
	})

	createBody := `{"name":"My Project","description":"Test"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/v1/projects", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+tokenValue)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	var created project.Project
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}

	saveBody := `{"stepData":{"evaluationGoals":["Usability & Playability"]}}`
	saveReq := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+created.ID+"/wizard/step/1", strings.NewReader(saveBody))
	saveReq.Header.Set("Content-Type", "application/json")
	saveReq.Header.Set("Authorization", "Bearer "+tokenValue)
	saveRec := httptest.NewRecorder()
	router.ServeHTTP(saveRec, saveReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+created.ID+"/wizard/recommendations", strings.NewReader(`{"forStep":2}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tokenValue)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
