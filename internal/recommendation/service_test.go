package recommendation

import (
	"encoding/json"
	"testing"

	"github.com/yifen9/gamidoc-backend/internal/wizard"
)

func TestRecommendStep2(t *testing.T) {
	engine := NewEngine(LoadDefaultRules())
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
	engine := NewEngine(LoadDefaultRules())
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
