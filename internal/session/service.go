package session

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrSessionNotFound = errors.New("session not found")

type Service struct {
	sessions Repository
	ttl      time.Duration
}

func NewService(sessions Repository, ttl time.Duration) *Service {
	return &Service{
		sessions: sessions,
		ttl:      ttl,
	}
}

func (s *Service) Create(ctx context.Context) (Session, error) {
	now := time.Now().UTC()

	return s.sessions.Create(ctx, Session{
		ID:        uuid.NewString(),
		Wizard:    NewInitialWizardStatus(),
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	})
}

func (s *Service) Get(ctx context.Context, id string) (Session, error) {
	return s.sessions.FindByID(ctx, id)
}
