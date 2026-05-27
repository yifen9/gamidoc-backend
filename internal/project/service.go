package project

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gamidoc/backend/internal/recommendation"
	"github.com/gamidoc/backend/internal/wizard"
	"github.com/google/uuid"
)

var ErrInvalidProjectName = errors.New("invalid project name")
var ErrProjectNotFound = errors.New("project not found")
var ErrForbiddenProject = errors.New("forbidden project")

type SessionSource struct {
	Wizard wizard.Status
	PDFURL *string
}

type SessionProjectReader interface {
	FindProjectSourceByID(ctx context.Context, sessionID string) (SessionSource, error)
	Delete(ctx context.Context, sessionID string) error
}

// PDFCleaner allows deleting a stored PDF file by URL.
type PDFCleaner interface {
	KeyFromURL(url string) (string, bool)
	Delete(ctx context.Context, key string) error
}

type Service struct {
	projects        Repository
	sessions        SessionProjectReader
	wizard          *wizard.Service
	recommendations *recommendation.Service
	store           PDFCleaner
}

type CreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ConvertInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SaveStepInput struct {
	StepData json.RawMessage `json:"stepData"`
}

func NewService(projects Repository, sessions SessionProjectReader, wizardService *wizard.Service, recommendationService *recommendation.Service) *Service {
	return &Service{
		projects:        projects,
		sessions:        sessions,
		wizard:          wizardService,
		recommendations: recommendationService,
	}
}

func (s *Service) WithPDFCleaner(store PDFCleaner) *Service {
	s.store = store
	return s
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Project, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if name == "" {
		return Project{}, ErrInvalidProjectName
	}

	return s.projects.Create(ctx, Project{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Wizard:      NewInitialWizardStatus(),
	})
}

func (s *Service) CreateFromSession(ctx context.Context, userID string, sessionID string, input ConvertInput) (Project, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)

	if name == "" {
		return Project{}, ErrInvalidProjectName
	}

	source, err := s.sessions.FindProjectSourceByID(ctx, sessionID)
	if err != nil {
		return Project{}, err
	}

	created, err := s.projects.Create(ctx, Project{
		ID:          uuid.NewString(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Wizard:      source.Wizard,
		PDFURL:      source.PDFURL,
	})
	if err != nil {
		return Project{}, err
	}

	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return Project{}, err
	}

	return created, nil
}

func (s *Service) List(ctx context.Context, userID string) ([]Project, error) {
	return s.projects.ListByUserID(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID string, projectID string) (Project, error) {
	found, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}

	if found.UserID != userID {
		return Project{}, ErrForbiddenProject
	}

	return found, nil
}

func (s *Service) SaveStep(ctx context.Context, userID string, projectID string, stepNumber int, stepData json.RawMessage) (Project, error) {
	found, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}

	if found.UserID != userID {
		return Project{}, ErrForbiddenProject
	}

	updatedStatus, err := s.wizard.SaveStep(found.Wizard, stepNumber, stepData)
	if err != nil {
		return Project{}, err
	}

	return s.projects.UpdateWizard(ctx, projectID, updatedStatus)
}

func (s *Service) Recommend(ctx context.Context, userID string, projectID string, forStep int) (recommendation.Result, error) {
	found, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return recommendation.Result{}, err
	}

	if found.UserID != userID {
		return recommendation.Result{}, ErrForbiddenProject
	}

	return s.recommendations.Recommend(found.Wizard, forStep)
}

func (s *Service) Update(ctx context.Context, userID string, projectID string, input UpdateInput) (Project, error) {
	found, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return Project{}, err
	}

	if found.UserID != userID {
		return Project{}, ErrForbiddenProject
	}

	name := found.Name
	if input.Name != nil {
		name = strings.TrimSpace(*input.Name)
		if name == "" {
			return Project{}, ErrInvalidProjectName
		}
	}

	description := found.Description
	if input.Description != nil {
		description = strings.TrimSpace(*input.Description)
	}

	return s.projects.UpdateInfo(ctx, projectID, name, description)
}

func (s *Service) Delete(ctx context.Context, userID string, projectID string) error {
	found, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return err
	}

	if found.UserID != userID {
		return ErrForbiddenProject
	}

	if s.store != nil && found.PDFURL != nil {
		if key, ok := s.store.KeyFromURL(*found.PDFURL); ok {
			_ = s.store.Delete(ctx, key)
		}
	}

	return s.projects.Delete(ctx, projectID)
}
