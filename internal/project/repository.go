package project

import (
	"context"

	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

type Repository interface {
	Create(ctx context.Context, input Project) (Project, error)
	ListByUserID(ctx context.Context, userID string) ([]Project, error)
	FindByID(ctx context.Context, id string) (Project, error)
	UpdateWizard(ctx context.Context, projectID string, status wizard.Status) (Project, error)
}
