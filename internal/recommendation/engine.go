package recommendation

import "sort"

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
	indexByID := map[string]int{}

	for _, rule := range e.rules {
		if rule.ForStep != input.ForStep {
			continue
		}

		if !matchesAll(input.EvaluationGoals, rule.RequiredEvaluationGoals) {
			continue
		}

		if !matchesAny(input.ProjectType, rule.RequiredProjectTypes) {
			continue
		}

		if !matchesAny(input.Participants, rule.RequiredParticipants) {
			continue
		}

		if !matchesAny(input.DevelopmentStage, rule.RequiredDevelopmentStages) {
			continue
		}

		if !matchesAll(input.SelectedMethods, rule.RequiredMethods) {
			continue
		}

		if !matchesAll(input.SelectedInstruments, rule.RequiredInstruments) {
			continue
		}

		if !matchesAny(input.Accessibility, rule.RequiredAccessibility) {
			continue
		}

		if !matchesAny(input.Time, rule.RequiredTime) {
			continue
		}

		if !matchesAnyInSlice(input.ExtraConstraints, rule.RequiredExtraConstraints) {
			continue
		}

		if rule.RequiredResearchEnabled != nil && *rule.RequiredResearchEnabled != input.ResearchEnabled {
			continue
		}

		signals := matchedSignals(input, rule)

		for _, rec := range rule.Recommendations {
			if i, ok := indexByID[rec.ID]; ok {
				// The same recommendation can be produced by several rules;
				// collect every signal that justifies it.
				result[i].MatchedOn = appendUniqueSignals(result[i].MatchedOn, signals)
				continue
			}
			rec.MatchedOn = appendUniqueSignals(nil, signals)
			indexByID[rec.ID] = len(result)
			result = append(result, rec)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		pi := priorityRank(result[i].Priority)
		pj := priorityRank(result[j].Priority)
		if pi != pj {
			return pi < pj
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// matchedSignals reports which of the user's inputs a matching rule required.
// A rule field left empty imposes no constraint and therefore yields no signal.
func matchedSignals(input Input, rule Rule) []MatchedSignal {
	var out []MatchedSignal

	for _, goal := range rule.RequiredEvaluationGoals {
		out = append(out, MatchedSignal{Kind: "goal", Value: goal})
	}
	for _, method := range rule.RequiredMethods {
		out = append(out, MatchedSignal{Kind: "method", Value: method})
	}
	if len(rule.RequiredProjectTypes) > 0 && input.ProjectType != "" {
		out = append(out, MatchedSignal{Kind: "projectType", Value: input.ProjectType})
	}
	if len(rule.RequiredParticipants) > 0 && input.Participants != "" {
		out = append(out, MatchedSignal{Kind: "participants", Value: input.Participants})
	}
	if len(rule.RequiredDevelopmentStages) > 0 && input.DevelopmentStage != "" {
		out = append(out, MatchedSignal{Kind: "developmentStage", Value: input.DevelopmentStage})
	}
	if len(rule.RequiredAccessibility) > 0 && input.Accessibility != "" {
		out = append(out, MatchedSignal{Kind: "accessibility", Value: input.Accessibility})
	}
	if len(rule.RequiredTime) > 0 && input.Time != "" {
		out = append(out, MatchedSignal{Kind: "time", Value: input.Time})
	}
	if len(rule.RequiredExtraConstraints) > 0 {
		have := map[string]bool{}
		for _, c := range input.ExtraConstraints {
			have[c] = true
		}
		for _, c := range rule.RequiredExtraConstraints {
			if have[c] {
				out = append(out, MatchedSignal{Kind: "constraint", Value: c})
			}
		}
	}
	if rule.RequiredResearchEnabled != nil && *rule.RequiredResearchEnabled {
		out = append(out, MatchedSignal{Kind: "research", Value: "Research specification"})
	}

	return out
}

func appendUniqueSignals(dst []MatchedSignal, src []MatchedSignal) []MatchedSignal {
	seen := map[MatchedSignal]bool{}
	for _, s := range dst {
		seen[s] = true
	}
	for _, s := range src {
		if seen[s] {
			continue
		}
		seen[s] = true
		dst = append(dst, s)
	}
	return dst
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

func matchesAny(have string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	for _, item := range required {
		if have == item {
			return true
		}
	}
	return false
}

// matchesAnyInSlice returns true if at least one element of required
// is present in have, or if required is empty (no constraint).
func matchesAnyInSlice(have []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := map[string]bool{}
	for _, item := range have {
		set[item] = true
	}
	for _, item := range required {
		if set[item] {
			return true
		}
	}
	return false
}

func priorityRank(priority string) int {
	switch priority {
	case "Recommended":
		return 0
	case "Engagement":
		return 1
	default:
		return 2
	}
}
