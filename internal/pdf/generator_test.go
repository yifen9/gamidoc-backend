package pdf

import "testing"

func TestFPDFGeneratorGenerate(t *testing.T) {
	generator := NewFPDFGenerator()

	data, err := generator.Generate(PlanData{
		Title:                  "Test Plan",
		EvaluationGoals:        []string{"Usability & Playability"},
		SelectedMethods:        []string{"surveys"},
		RecommendedInstruments: []string{"USEQ-Like", "SUS"},
		NextSteps:              []string{"Prepare materials"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(data) == 0 {
		t.Fatal("expected non-empty pdf bytes")
	}
}
