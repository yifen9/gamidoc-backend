package pdf

import (
	"encoding/json"
	"time"

	"github.com/yifen9/gamidoc-backend/internal/project"
	"github.com/yifen9/gamidoc-backend/internal/recommendation"
	"github.com/yifen9/gamidoc-backend/internal/session"
)

type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

type step1Data struct {
	EvaluationGoals []string `json:"evaluationGoals"`
}

type step2Data struct {
	SelectedMethods []string `json:"selectedMethods"`
}

func (b *Builder) BuildFromProject(item project.Project, recs []recommendation.Recommendation) (PlanData, error) {
	return b.build(item.Name, item.CreatedAt, item.Wizard.Steps, recs)
}

func (b *Builder) BuildFromSession(item session.Session, recs []recommendation.Recommendation) (PlanData, error) {
	return b.build("Anonymous Evaluation Plan", item.CreatedAt, item.Wizard.Steps, recs)
}

func (b *Builder) build(title string, createdAt time.Time, steps map[string]json.RawMessage, recs []recommendation.Recommendation) (PlanData, error) {
	var step1 step1Data
	var step2 step2Data

	if raw, ok := steps["1"]; ok {
		_ = json.Unmarshal(raw, &step1)
	}

	if raw, ok := steps["2"]; ok {
		_ = json.Unmarshal(raw, &step2)
	}

	var recommendedInstruments []string
	for _, rec := range recs {
		recommendedInstruments = append(recommendedInstruments, rec.Name)
	}

	nextSteps := []string{
		"Review the selected methods and instruments.",
		"Prepare the evaluation materials and participant setup.",
		"Run the evaluation and consolidate findings.",
	}

	return PlanData{
		Title:                  title,
		Date:                   createdAt,
		EvaluationGoals:        step1.EvaluationGoals,
		SelectedMethods:        step2.SelectedMethods,
		RecommendedInstruments: recommendedInstruments,
		NextSteps:              nextSteps,
	}, nil
}
