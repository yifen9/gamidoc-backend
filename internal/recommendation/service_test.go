package recommendation

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

func TestLoadRulesFromFile(t *testing.T) {
	path := filepath.Join("..", "..", "rule", "recommendations.json")

	rules, err := LoadRulesFromFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(rules) == 0 {
		t.Fatal("expected at least one rule")
	}
}

func TestRecommendStep2(t *testing.T) {
	engine := NewEngine(LoadDefaultRulesForTest())
	service := NewService(engine)

	step1, _ := json.Marshal(map[string]any{
		"evaluationGoals": []string{"Usability & Playability"},
	})

	status := wizard.Status{
		CurrentStep: 2,
		IsComplete:  false,
		Steps: map[string]json.RawMessage{
			"1": step1,
		},
	}

	result, err := service.Recommend(status, 2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Recommendations) == 0 {
		t.Fatal("expected at least one recommendation")
	}
}

func TestRecommendStep3(t *testing.T) {
	engine := NewEngine(LoadDefaultRulesForTest())
	service := NewService(engine)

	step2, _ := json.Marshal(map[string]any{
		"selectedMethods": []string{"surveys"},
	})

	status := wizard.Status{
		CurrentStep: 3,
		IsComplete:  false,
		Steps: map[string]json.RawMessage{
			"2": step2,
		},
	}

	result, err := service.Recommend(status, 3)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Recommendations) == 0 {
		t.Fatal("expected at least one recommendation")
	}
}

func LoadDefaultRulesForTest() []Rule {
	return []Rule{
		{
			ForStep:                 2,
			RequiredEvaluationGoals: []string{"Usability & Playability"},
			Recommendations: []Recommendation{
				{
					ID: "think-aloud",
				},
			},
		},
		{
			ForStep:         3,
			RequiredMethods: []string{"surveys"},
			Recommendations: []Recommendation{
				{
					ID: "sus",
				},
			},
		},
	}
}
