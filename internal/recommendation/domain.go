package recommendation

type Recommendation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Rationale   string `json:"rationale"`
	// MatchedOn lists the user inputs that caused this recommendation to be
	// selected, so the UI can explain why it fits the project.
	MatchedOn []MatchedSignal `json:"matchedOn,omitempty"`
}

// MatchedSignal is one user input that a matching rule required.
type MatchedSignal struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Result struct {
	ForStep         int              `json:"forStep"`
	Recommendations []Recommendation `json:"recommendations"`
}
