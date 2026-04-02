package pdf

import "time"

type PlanData struct {
	Title                  string
	Date                   time.Time
	EvaluationGoals        []string
	SelectedMethods        []string
	RecommendedInstruments []string
	NextSteps              []string
}

type Generated struct {
	Key string
	URL string
}
