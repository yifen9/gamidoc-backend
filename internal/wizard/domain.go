package wizard

import "encoding/json"

type Status struct {
	CurrentStep int                        `json:"currentStep"`
	IsComplete  bool                       `json:"isComplete"`
	Steps       map[string]json.RawMessage `json:"steps"`
}

func NewInitialStatus() Status {
	return Status{
		CurrentStep: 1,
		IsComplete:  false,
		Steps:       map[string]json.RawMessage{},
	}
}
