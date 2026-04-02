package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

var ErrSessionNotFound = errors.New("session not found")

type Service struct {
	sessions Repository
	ttl      time.Duration
	wizard   *wizard.Service
}

func NewService(sessions Repository, ttl time.Duration, wizardService *wizard.Service) *Service {
	return &Service{
		sessions: sessions,
		ttl:      ttl,
		wizard:   wizardService,
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

func (s *Service) SaveStep(ctx context.Context, sessionID string, stepNumber int, stepData json.RawMessage) (Session, error) {
	found, err := s.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return Session{}, err
	}

	updatedStatus, err := s.wizard.SaveStep(found.Wizard, stepNumber, stepData)
	if err != nil {
		return Session{}, err
	}

	return s.sessions.UpdateWizard(ctx, sessionID, updatedStatus)
}
