package recommendation

type Rule struct {
	ForStep                 int              `json:"forStep"`
	RequiredEvaluationGoals []string         `json:"requiredEvaluationGoals"`
	RequiredMethods         []string         `json:"requiredMethods"`
	Recommendations         []Recommendation `json:"recommendations"`
}
