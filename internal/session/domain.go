package session

import (
	"encoding/json"
	"time"
)

type WizardStatus struct {
	CurrentStep int                        `json:"currentStep"`
	IsComplete  bool                       `json:"isComplete"`
	Steps       map[string]json.RawMessage `json:"steps"`
}

type Session struct {
	ID        string       `json:"sessionId"`
	Wizard    WizardStatus `json:"wizardStatus"`
	CreatedAt time.Time    `json:"createdAt"`
	ExpiresAt time.Time    `json:"expiresAt"`
	PDFURL    *string      `json:"pdfUrl,omitempty"`
}

func NewInitialWizardStatus() WizardStatus {
	return WizardStatus{
		CurrentStep: 1,
		IsComplete:  false,
		Steps:       map[string]json.RawMessage{},
	}
}
