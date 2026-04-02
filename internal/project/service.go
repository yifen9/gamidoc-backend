package project

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

var ErrInvalidProjectName = errors.New("invalid project name")
var ErrProjectNotFound = errors.New("project not found")
var ErrForbiddenProject = errors.New("forbidden project")

type Service struct {
	projects Repository
	wizard   *wizard.Service
}

type CreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SaveStepInput struct {
	StepData json.RawMessage `json:"stepData"`
}

func NewService(projects Repository, wizardService *wizard.Service) *Service {
	return &Service{
		projects: projects,
		wizard:   wizardService,
	}
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
