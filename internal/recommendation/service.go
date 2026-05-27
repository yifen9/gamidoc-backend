package recommendation

import (
	"errors"
	"strings"

	"github.com/gamidoc/backend/internal/wizard"
)

var ErrInvalidRecommendationStep = errors.New("invalid recommendation step")

// methodNameToID maps the display names stored by the frontend to the rule IDs
// used in recommendations.json. Any name not found here is normalised by
// lower-casing and replacing spaces with hyphens.
var methodNameToID = map[string]string{
	"Think-aloud testing":      "think-aloud",
	"Surveys & Questionnaires": "surveys",
	"Heuristic evaluation":     "heuristic-evaluation",
	"Expert review":            "expert-review",
	"Interview":                "interview",
	"Observation":              "observation",
	"Focus Group":              "focus-group",
	"Diary Study":              "diary-study",
	"Experience Report":        "experience-report",
}

var instrumentNameToID = map[string]string{
	"SUS":        "sus",
	"UEQ":        "ueq",
	"UMUX-Lite":  "umux-lite",
	"AttrakDiff": "attrakdiff",
	"meCUE":      "mecue",
	"SAM":        "sam",
	"PANAS":      "panas",
	"GEQ":        "geq",
	"NASA-TLX":   "nasa-tlx",
	"PENS":       "pens",
	"IMI":        "imi",
	"BNS":        "bns",
	"FSS-2":      "fss2",
	"EGameFlow":  "egameflow",
	"miniPXI":    "minipxi",
	"MEEGA+":     "meega",
}

func normalizeMethodName(name string) string {
	if id, ok := methodNameToID[name]; ok {
		return id
	}
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

func normalizeInstrumentName(name string) string {
	if id, ok := instrumentNameToID[name]; ok {
		return id
	}
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

type Service struct {
	engine *Engine
}

type Input struct {
	ForStep             int
	EvaluationGoals     []string
	ProjectType         string
	Participants        string
	DevelopmentStage    string
	SelectedMethods     []string
	SelectedInstruments []string
	Accessibility       string
	Time                string
	ExtraConstraints    []string
	ResearchEnabled     bool
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

	if step1, ok := wizard.DecodeStep1(status); ok {
		input.EvaluationGoals = step1.EvaluationGoals
		input.ProjectType = step1.ProjectType
		input.Participants = step1.Participants
		input.DevelopmentStage = step1.DevelopmentStage
		input.Accessibility = step1.Accessibility
		input.Time = step1.Time
		input.ExtraConstraints = step1.ExtraConstraints
		input.ResearchEnabled = step1.ResearchEnabled
	}

	if step2, ok := wizard.DecodeStep2(status); ok {
		normalized := make([]string, len(step2.SelectedMethods))
		for i, name := range step2.SelectedMethods {
			normalized[i] = normalizeMethodName(name)
		}
		input.SelectedMethods = normalized
	}

	if step3, ok := wizard.DecodeStep3(status); ok {
		normalized := make([]string, len(step3.SelectedInstruments))
		for i, name := range step3.SelectedInstruments {
			normalized[i] = normalizeInstrumentName(name)
		}
		input.SelectedInstruments = normalized
	}

	return Result{
		ForStep:         forStep,
		Recommendations: s.engine.Recommend(input),
	}, nil
}
