package project

import "context"

type Repository interface {
	Create(ctx context.Context, input Project) (Project, error)
	ListByUserID(ctx context.Context, userID string) ([]Project, error)
	FindByID(ctx context.Context, id string) (Project, error)
}
