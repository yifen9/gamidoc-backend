package wizard

import (
	"encoding/json"
	"errors"
	"strconv"
)

var ErrInvalidStepNumber = errors.New("invalid step number")
var ErrInvalidStepData = errors.New("invalid step data")

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) SaveStep(current Status, stepNumber int, stepData json.RawMessage) (Status, error) {
	if stepNumber < 1 || stepNumber > 4 {
		return Status{}, ErrInvalidStepNumber
	}

	if len(stepData) == 0 || !json.Valid(stepData) {
		return Status{}, ErrInvalidStepData
	}

	if current.Steps == nil {
		current.Steps = map[string]json.RawMessage{}
	}

	current.Steps[strconv.Itoa(stepNumber)] = stepData
	current.CurrentStep = s.computeCurrentStep(current)
	current.IsComplete = s.computeIsComplete(current)

	return current, nil
}

func (s *Service) computeCurrentStep(status Status) int {
	for step := 1; step <= 4; step++ {
		if _, ok := status.Steps[strconv.Itoa(step)]; !ok {
			return step
		}
	}
	return 4
}

func (s *Service) computeIsComplete(status Status) bool {
	for step := 1; step <= 4; step++ {
		if _, ok := status.Steps[strconv.Itoa(step)]; !ok {
			return false
		}
	}
	return true
}
