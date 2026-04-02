package recommendation

import (
	"encoding/json"
	"errors"

	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

var ErrInvalidRecommendationStep = errors.New("invalid recommendation step")

type Service struct {
	engine *Engine
}

type Input struct {
	ForStep         int
	EvaluationGoals []string
	SelectedMethods []string
}

type step1Data struct {
	EvaluationGoals []string `json:"evaluationGoals"`
}

type step2Data struct {
	SelectedMethods []string `json:"selectedMethods"`
}

func NewService(engine *Engine) *Service {
	return &Service{
		engine: engine,
	}
}

func (s *Service) Recommend(status wizard.Status, forStep int) (Result, error) {
	if forStep < 2 || forStep > 4 {
		return Result{}, ErrInvalidRecommendationStep
	}

	input := Input{
		ForStep: forStep,
	}

	if raw, ok := status.Steps["1"]; ok {
		var step1 step1Data
		if err := json.Unmarshal(raw, &step1); err == nil {
			input.EvaluationGoals = step1.EvaluationGoals
		}
	}

	if raw, ok := status.Steps["2"]; ok {
		var step2 step2Data
		if err := json.Unmarshal(raw, &step2); err == nil {
			input.SelectedMethods = step2.SelectedMethods
		}
	}

	return Result{
		ForStep:         forStep,
		Recommendations: s.engine.Recommend(input),
	}, nil
}
