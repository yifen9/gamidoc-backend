package recommendation

func LoadDefaultRules() []Rule {
	return []Rule{
		{
			ForStep:                 2,
			RequiredEvaluationGoals: []string{"Usability & Playability"},
			Recommendations: []Recommendation{
				{
					ID:          "think-aloud",
					Name:        "Think-aloud testing",
					Description: "Users verbalize thoughts while interacting with the system",
					Priority:    "Recommended",
					Rationale:   "Useful for early usability evaluation with limited participants",
				},
				{
					ID:          "surveys",
					Name:        "Surveys & Questionnaires",
					Description: "Collect structured user feedback through questionnaires",
					Priority:    "Recommended",
					Rationale:   "Helpful for measuring perceived usability and satisfaction",
				},
			},
		},
		{
			ForStep:                 2,
			RequiredEvaluationGoals: []string{"Guidance & Feedback"},
			Recommendations: []Recommendation{
				{
					ID:          "heuristic-evaluation",
					Name:        "Heuristic evaluation",
					Description: "Experts inspect the interface against established usability principles",
					Priority:    "Recommended",
					Rationale:   "Useful for identifying guidance and feedback issues early",
				},
				{
					ID:          "surveys",
					Name:        "Surveys & Questionnaires",
					Description: "Collect structured user feedback through questionnaires",
					Priority:    "Engagement",
					Rationale:   "Can complement expert-based findings with user perceptions",
				},
			},
		},
		{
			ForStep:         3,
			RequiredMethods: []string{"surveys"},
			Recommendations: []Recommendation{
				{
					ID:          "useq-like",
					Name:        "USEQ-Like",
					Description: "A short questionnaire targeting usability experience",
					Priority:    "Recommended",
					Rationale:   "Suitable when using questionnaires for usability evaluation",
				},
				{
					ID:          "sus",
					Name:        "SUS",
					Description: "System Usability Scale",
					Priority:    "Engagement",
					Rationale:   "Widely used instrument for perceived usability",
				},
			},
		},
		{
			ForStep:         3,
			RequiredMethods: []string{"think-aloud"},
			Recommendations: []Recommendation{
				{
					ID:          "observation-grid",
					Name:        "Observation Grid",
					Description: "Structured note-taking sheet for observed usability issues",
					Priority:    "Recommended",
					Rationale:   "Useful for capturing qualitative insights during think-aloud sessions",
				},
			},
		},
	}
}
