package session

import "context"

type Repository interface {
	Create(ctx context.Context, input Session) (Session, error)
	FindByID(ctx context.Context, id string) (Session, error)
}
