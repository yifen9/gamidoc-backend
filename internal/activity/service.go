package activity

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, event Event) error {
	if s == nil || s.repository == nil {
		return nil
	}

	event.Type = strings.TrimSpace(event.Type)
	if event.Type == "" {
		event.Type = EventAPIRequest
	}
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}

	return s.repository.Save(ctx, event)
}
