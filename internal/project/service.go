package project

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var ErrInvalidProjectName = errors.New("invalid project name")
var ErrProjectNotFound = errors.New("project not found")
var ErrForbiddenProject = errors.New("forbidden project")

type Service struct {
	projects Repository
}

type CreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewService(projects Repository) *Service {
	return &Service{
		projects: projects,
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
