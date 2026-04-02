package project

import (
	"encoding/json"
	"time"
)

type WizardStatus struct {
	CurrentStep int                        `json:"currentStep"`
	IsComplete  bool                       `json:"isComplete"`
	Steps       map[string]json.RawMessage `json:"steps"`
}

type Project struct {
	ID          string       `json:"projectId"`
	UserID      string       `json:"-"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Wizard      WizardStatus `json:"wizardStatus"`
	PDFURL      *string      `json:"pdfUrl"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

func NewInitialWizardStatus() WizardStatus {
	return WizardStatus{
		CurrentStep: 1,
		IsComplete:  false,
		Steps:       map[string]json.RawMessage{},
	}
}
