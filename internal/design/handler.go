package design

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gamidoc/backend/internal/ai"
	appmiddleware "github.com/gamidoc/backend/internal/http/middleware"
	"github.com/gamidoc/backend/internal/http/response"
	"github.com/gamidoc/backend/internal/project"
	"github.com/gamidoc/backend/internal/session"
	"github.com/gamidoc/backend/internal/storage/objectstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SessionGuard interface {
	FindByID(ctx context.Context, id string) (session.Session, error)
}

type ProjectGuard interface {
	FindByID(ctx context.Context, id string) (project.Project, error)
}

type Handler struct {
	service       *Service
	assistant     ai.Assistant
	builder       *ReportBuilder
	store         objectstore.ObjectStore
	sessions      SessionGuard
	sessionStates StateStore
	projects      ProjectGuard
	projectStates StateStore
	reports       ReportRepository
}

func NewHandler(
	service *Service,
	assistant ai.Assistant,
	builder *ReportBuilder,
	store objectstore.ObjectStore,
	sessions SessionGuard,
	sessionStates StateStore,
	projects ProjectGuard,
	projectStates StateStore,
	reports ReportRepository,
) *Handler {
	return &Handler{
		service:       service,
		assistant:     assistant,
		builder:       builder,
		store:         store,
		sessions:      sessions,
		sessionStates: sessionStates,
		projects:      projects,
		projectStates: projectStates,
		reports:       reports,
	}
}

type owner struct {
	kind   string
	id     string
	states StateStore
}

func (h *Handler) SessionRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.forSession(h.getState))
	r.Put("/spark", h.forSession(h.saveSpark))
	r.Get("/branch", h.forSession(h.branch))
	r.Post("/path", h.forSession(h.choosePath))
	r.Put("/section/{sectionNumber}", h.forSession(h.saveSection))
	r.Get("/dashboard", h.forSession(h.dashboard))
	r.Post("/generate-pdf", h.forSession(h.generatePDF))

	return r
}

func (h *Handler) ProjectRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.forProject(h.getState))
	r.Put("/spark", h.forProject(h.saveSpark))
	r.Get("/branch", h.forProject(h.branch))
	r.Post("/path", h.forProject(h.choosePath))
	r.Put("/section/{sectionNumber}", h.forProject(h.saveSection))
	r.Get("/dashboard", h.forProject(h.dashboard))
	r.Post("/generate-pdf", h.forProject(h.generatePDF))
	r.Get("/reports", h.forProject(h.listReports))
	r.Post("/import-session", h.forProject(h.importSession))

	return r
}

func (h *Handler) AIRoutes() chi.Router {
	r := chi.NewRouter()

	r.Post("/rewrite", h.rewrite)
	r.Post("/chat", h.chat)

	return r
}

func (h *Handler) forSession(next func(http.ResponseWriter, *http.Request, owner)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "sessionId")
		if sessionID == "" {
			response.WriteError(w, http.StatusBadRequest, "INVALID_SESSION_ID", "Invalid session id", nil)
			return
		}

		if _, err := h.sessions.FindByID(r.Context(), sessionID); err != nil {
			if errors.Is(err, session.ErrSessionNotFound) {
				response.WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", nil)
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
			return
		}

		next(w, r, owner{kind: "sessions", id: sessionID, states: h.sessionStates})
	}
}

func (h *Handler) forProject(next func(http.ResponseWriter, *http.Request, owner)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		found, err := h.projects.FindByID(r.Context(), projectID)
		if err != nil {
			if errors.Is(err, project.ErrProjectNotFound) {
				response.WriteError(w, http.StatusNotFound, "PROJECT_NOT_FOUND", "Project not found", nil)
				return
			}
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
			return
		}

		if found.UserID != userID {
			response.WriteError(w, http.StatusForbidden, "FORBIDDEN", "Project does not belong to user", nil)
			return
		}

		next(w, r, owner{kind: "projects", id: projectID, states: h.projectStates})
	}
}

func (h *Handler) getState(w http.ResponseWriter, r *http.Request, o owner) {
	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) saveSpark(w http.ResponseWriter, r *http.Request, o owner) {
	var input struct {
		Spark string `json:"spark"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	status = h.service.SaveSpark(status, input.Spark)

	if status.Spark != "" {
		prefill, err := h.assistant.Prefill(r.Context(), status.Spark, sectionsMeta())
		if err == nil {
			converted := make(map[string]json.RawMessage, len(prefill))
			for number, content := range prefill {
				converted[SectionKey(number)] = content
			}
			status = h.service.ApplyPrefill(status, converted)
		}
	}

	if err := o.states.Save(r.Context(), o.id, status); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) branch(w http.ResponseWriter, r *http.Request, o owner) {
	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	if status.Spark == "" {
		response.WriteJSON(w, http.StatusOK, map[string]any{
			"recommendedPath": PathExperienceFirst,
			"basis":           "default",
		})
		return
	}

	recommended, err := h.assistant.RecommendBranch(r.Context(), status.Spark)
	if err != nil {
		response.WriteError(w, http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider error", nil)
		return
	}
	if recommended != PathMechanicsFirst {
		recommended = PathExperienceFirst
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"recommendedPath": recommended,
		"basis":           "spark",
	})
}

func (h *Handler) choosePath(w http.ResponseWriter, r *http.Request, o owner) {
	var input struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	updated, err := h.service.ChoosePath(status, input.Path)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidPath):
			response.WriteError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid path", nil)
		case errors.Is(err, ErrPathAlreadyChosen):
			response.WriteError(w, http.StatusConflict, "PATH_ALREADY_CHOSEN", "Path already chosen", nil)
		case errors.Is(err, ErrSectionLocked):
			response.WriteError(w, http.StatusBadRequest, "SECTION_LOCKED", "Complete the Context section first", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	if err := o.states.Save(r.Context(), o.id, updated); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handler) saveSection(w http.ResponseWriter, r *http.Request, o owner) {
	sectionValue := chi.URLParam(r, "sectionNumber")
	sectionNumber, err := strconv.Atoi(sectionValue)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_SECTION_NUMBER", "Invalid section number", nil)
		return
	}

	var input struct {
		Content  json.RawMessage `json:"content"`
		Complete bool            `json:"complete"`
		Skip     bool            `json:"skip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	updated, err := h.service.SaveSection(status, sectionNumber, input.Content, input.Complete, input.Skip)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidSectionNumber):
			response.WriteError(w, http.StatusBadRequest, "INVALID_SECTION_NUMBER", "Invalid section number", nil)
		case errors.Is(err, ErrInvalidSectionData):
			response.WriteError(w, http.StatusBadRequest, "INVALID_SECTION_DATA", "Invalid section data", nil)
		case errors.Is(err, ErrSectionLocked):
			response.WriteError(w, http.StatusBadRequest, "SECTION_LOCKED", "Sections must be traversed in order on the first pass", nil)
		case errors.Is(err, ErrPathNotChosen):
			response.WriteError(w, http.StatusBadRequest, "PATH_NOT_CHOSEN", "Choose a path after the Context section", nil)
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		}
		return
	}

	if err := o.states.Save(r.Context(), o.id, updated); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"sectionNumber": sectionNumber,
		"section":       updated.Section(sectionNumber),
		"designStatus":  updated,
	})
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request, o owner) {
	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	dashboard, err := h.service.Dashboard(status)
	if err != nil {
		if errors.Is(err, ErrDashboardLocked) {
			response.WriteError(w, http.StatusForbidden, "DASHBOARD_LOCKED", "Complete a full first pass with at least one filled section first", nil)
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, dashboard)
}

func (h *Handler) generatePDF(w http.ResponseWriter, r *http.Request, o owner) {
	var input struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}
	if input.Version == "" {
		input.Version = ReportVersionStandard
	}
	if input.Version != ReportVersionStandard && input.Version != ReportVersionEnhanced {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REPORT_VERSION", "Invalid report version", nil)
		return
	}

	status, err := o.states.Get(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	if _, err := h.service.Dashboard(status); err != nil {
		response.WriteError(w, http.StatusForbidden, "DASHBOARD_LOCKED", "Complete a full first pass with at least one filled section first", nil)
		return
	}

	sections, err := h.renderSections(r.Context(), status, input.Version)
	if err != nil {
		response.WriteError(w, http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider error", nil)
		return
	}

	data, err := h.builder.Build("Gamification Design Report", input.Version, status.Spark, sections)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	reportID := uuid.NewString()
	url, err := h.store.Save(r.Context(), "design/"+o.kind+"/"+o.id+"/"+reportID+".pdf", data)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	if o.kind == "projects" {
		created, err := h.reports.Create(r.Context(), Report{
			ID:        reportID,
			ProjectID: o.id,
			Version:   input.Version,
			URL:       url,
		})
		if err != nil {
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
			return
		}
		response.WriteJSON(w, http.StatusOK, created)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"reportId": reportID,
		"version":  input.Version,
		"url":      url,
	})
}

func (h *Handler) renderSections(ctx context.Context, status Status, version string) ([]RenderSection, error) {
	if version == ReportVersionStandard {
		var sections []RenderSection
		for number := 1; number <= SectionCount; number++ {
			sections = append(sections, RenderSection{
				Number: number,
				Name:   SectionName(number),
				Lines:  SectionLines(status.Section(number).Content),
			})
		}
		return sections, nil
	}

	var inputs []ai.SectionText
	for number := 1; number <= SectionCount; number++ {
		text := SectionPlainText(status.Section(number).Content)
		if text == "" {
			continue
		}
		inputs = append(inputs, ai.SectionText{
			Number: number,
			Name:   SectionName(number),
			Text:   text,
		})
	}

	enhanced, err := h.assistant.Enhance(ctx, inputs)
	if err != nil {
		return nil, err
	}

	prose := make(map[int]string, len(enhanced))
	for _, section := range enhanced {
		prose[section.Number] = section.Text
	}

	var sections []RenderSection
	for number := 1; number <= SectionCount; number++ {
		var lines []string
		if text, ok := prose[number]; ok && text != "" {
			lines = []string{text}
		}
		sections = append(sections, RenderSection{
			Number: number,
			Name:   SectionName(number),
			Lines:  lines,
		})
	}
	return sections, nil
}

func (h *Handler) listReports(w http.ResponseWriter, r *http.Request, o owner) {
	reports, err := h.reports.ListByProjectID(r.Context(), o.id)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"reports": reports,
		"total":   len(reports),
	})
}

func (h *Handler) importSession(w http.ResponseWriter, r *http.Request, o owner) {
	var input struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.SessionID == "" {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	if _, err := h.sessions.FindByID(r.Context(), input.SessionID); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			response.WriteError(w, http.StatusNotFound, "SESSION_NOT_FOUND", "Session not found", nil)
			return
		}
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	status, err := h.sessionStates.Get(r.Context(), input.SessionID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	if err := o.states.Save(r.Context(), o.id, status); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Internal server error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, status)
}

func (h *Handler) rewrite(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}

	rewritten, err := h.assistant.Rewrite(r.Context(), input.Text)
	if err != nil {
		if errors.Is(err, ai.ErrEmptyText) {
			response.WriteError(w, http.StatusBadRequest, "EMPTY_TEXT", "Text is required", nil)
			return
		}
		response.WriteError(w, http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"text":         rewritten,
		"previousText": input.Text,
	})
}

func (h *Handler) chat(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SectionNumber int    `json:"sectionNumber"`
		Message       string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_INPUT", "Invalid request body", nil)
		return
	}
	if input.SectionNumber < 1 || input.SectionNumber > SectionCount {
		response.WriteError(w, http.StatusBadRequest, "INVALID_SECTION_NUMBER", "Invalid section number", nil)
		return
	}

	reply, err := h.assistant.Chat(r.Context(), ai.Section{
		Number: input.SectionNumber,
		Name:   SectionName(input.SectionNumber),
	}, input.Message)
	if err != nil {
		response.WriteError(w, http.StatusBadGateway, "AI_PROVIDER_ERROR", "AI provider error", nil)
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]any{
		"sectionNumber": input.SectionNumber,
		"reply":         reply,
	})
}

func sectionsMeta() []ai.Section {
	var sections []ai.Section
	for number := 1; number <= SectionCount; number++ {
		sections = append(sections, ai.Section{
			Number: number,
			Name:   SectionName(number),
		})
	}
	return sections
}
