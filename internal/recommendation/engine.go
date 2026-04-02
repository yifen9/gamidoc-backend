package recommendation

type Engine struct {
	rules []Rule
}

func NewEngine(rules []Rule) *Engine {
	return &Engine{
		rules: rules,
	}
}

func (e *Engine) Recommend(input Input) []Recommendation {
	var result []Recommendation
	seen := map[string]bool{}

	for _, rule := range e.rules {
		if rule.ForStep != input.ForStep {
			continue
		}

		if !matchesAll(input.EvaluationGoals, rule.RequiredEvaluationGoals) {
			continue
		}

		if !matchesAll(input.SelectedMethods, rule.RequiredMethods) {
			continue
		}

		for _, rec := range rule.Recommendations {
			if seen[rec.ID] {
				continue
			}
			seen[rec.ID] = true
			result = append(result, rec)
		}
	}

	return result
}

func matchesAll(have []string, required []string) bool {
	if len(required) == 0 {
		return true
	}

	set := map[string]bool{}
	for _, item := range have {
		set[item] = true
	}

	for _, item := range required {
		if !set[item] {
			return false
		}
	}

	return true
}
