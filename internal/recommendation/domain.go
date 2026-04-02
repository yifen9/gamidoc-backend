package recommendation

type Recommendation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Rationale   string `json:"rationale"`
}

type Result struct {
	ForStep         int              `json:"forStep"`
	Recommendations []Recommendation `json:"recommendations"`
}
