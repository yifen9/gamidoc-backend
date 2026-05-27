package recommendation

type Rule struct {
	ForStep                   int              `json:"forStep"`
	RequiredEvaluationGoals   []string         `json:"requiredEvaluationGoals"`
	RequiredProjectTypes      []string         `json:"requiredProjectTypes"`
	RequiredParticipants      []string         `json:"requiredParticipants"`
	RequiredDevelopmentStages []string         `json:"requiredDevelopmentStages"`
	RequiredMethods           []string         `json:"requiredMethods"`
	RequiredAccessibility     []string         `json:"requiredAccessibility"`
	RequiredTime              []string         `json:"requiredTime"`
	RequiredExtraConstraints  []string         `json:"requiredExtraConstraints"`
	RequiredResearchEnabled   *bool            `json:"requiredResearchEnabled"`
	Recommendations           []Recommendation `json:"recommendations"`
}
