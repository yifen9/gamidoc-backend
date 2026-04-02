package recommendation

type Rule struct {
	ForStep                 int
	RequiredEvaluationGoals []string
	RequiredMethods         []string
	Recommendations         []Recommendation
}
