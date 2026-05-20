package pdf

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

type Generator interface {
	Generate(data PlanData) ([]byte, error)
}

type FPDFGenerator struct{}

func NewFPDFGenerator() *FPDFGenerator {
	return &FPDFGenerator{}
}

func (g *FPDFGenerator) Generate(data PlanData) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetTitle(data.Title, false)
	doc.AddPage()
	doc.SetFont("Arial", "B", 16)
	doc.Cell(0, 10, data.Title)
	doc.Ln(12)

	doc.SetFont("Arial", "", 11)
	doc.Cell(0, 8, "Date: "+data.Date.Format("2006-01-02 15:04:05"))
	doc.Ln(10)

	writeContextSection(doc, data)
	writeConstraintsSection(doc, data)
	if data.ResearchEnabled {
		writeResearchSection(doc, data)
	}
	writeMethodsSection(doc, data.SelectedMethods)
	writeInstrumentsSection(doc, data.SelectedInstruments)
	writeSimpleSection(doc, "Next Steps", data.NextSteps)

	if strings.TrimSpace(data.Notes) != "" {
		doc.SetFont("Arial", "B", 13)
		doc.Cell(0, 8, "Notes")
		doc.Ln(9)
		doc.SetFont("Arial", "", 11)
		doc.MultiCell(0, 6, data.Notes, "", "L", false)
		doc.Ln(2)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeContextSection(doc *fpdf.Fpdf, data PlanData) {
	doc.SetFont("Arial", "B", 13)
	doc.Cell(0, 8, "Evaluation Context")
	doc.Ln(9)

	doc.SetFont("Arial", "", 11)
	doc.MultiCell(0, 6, "Evaluation Goals: "+joinOrNone(data.EvaluationGoals), "", "L", false)
	doc.MultiCell(0, 6, "Project Type: "+valueOrNone(data.ProjectType), "", "L", false)
	doc.MultiCell(0, 6, "Participants: "+valueOrNone(data.Participants), "", "L", false)
	doc.MultiCell(0, 6, "Development Stage: "+valueOrNone(data.DevelopmentStage), "", "L", false)
	doc.Ln(2)
}

func writeConstraintsSection(doc *fpdf.Fpdf, data PlanData) {
	hasAccessibility := strings.TrimSpace(data.Accessibility) != ""
	hasTime := strings.TrimSpace(data.TimeConstraint) != ""
	hasExtra := len(data.ExtraConstraints) > 0
	if !hasAccessibility && !hasTime && !hasExtra {
		return
	}

	doc.SetFont("Arial", "B", 13)
	doc.Cell(0, 8, "Constraints")
	doc.Ln(9)
	doc.SetFont("Arial", "", 11)
	if hasAccessibility {
		doc.MultiCell(0, 6, "User Accessibility: "+data.Accessibility, "", "L", false)
	}
	if hasTime {
		doc.MultiCell(0, 6, "Available Time: "+data.TimeConstraint, "", "L", false)
	}
	if hasExtra {
		doc.MultiCell(0, 6, "Additional Constraints: "+strings.Join(data.ExtraConstraints, ", "), "", "L", false)
	}
	doc.Ln(2)
}

func writeResearchSection(doc *fpdf.Fpdf, data PlanData) {
	doc.SetFont("Arial", "B", 13)
	doc.Cell(0, 8, "Research Specification")
	doc.Ln(9)
	doc.SetFont("Arial", "", 11)

	if strings.TrimSpace(data.ResearchObjective) != "" {
		doc.MultiCell(0, 6, "Objective: "+data.ResearchObjective, "", "L", false)
	}
	if len(data.ResearchQuestions) > 0 {
		doc.MultiCell(0, 6, "Research Questions:", "", "L", false)
		for i, rq := range data.ResearchQuestions {
			if strings.TrimSpace(rq) != "" {
				doc.MultiCell(0, 6, fmt.Sprintf("  RQ%d: %s", i+1, rq), "", "L", false)
			}
		}
	}
	if len(data.Hypotheses) > 0 {
		doc.MultiCell(0, 6, "Hypotheses:", "", "L", false)
		for i, h := range data.Hypotheses {
			if strings.TrimSpace(h) != "" {
				doc.MultiCell(0, 6, fmt.Sprintf("  H%d: %s", i+1, h), "", "L", false)
			}
		}
	}
	doc.Ln(2)
}

func writeMethodsSection(doc *fpdf.Fpdf, items []MethodEntry) {
	doc.SetFont("Arial", "B", 13)
	doc.Cell(0, 8, "Selected Methods")
	doc.Ln(9)

	doc.SetFont("Arial", "", 11)
	if len(items) == 0 {
		doc.MultiCell(0, 6, "- None", "", "L", false)
		doc.Ln(2)
		return
	}

	for _, item := range items {
		lines := []string{fmt.Sprintf("- %s", strings.TrimSpace(item.Name))}
		if strings.TrimSpace(item.Description) != "" {
			lines = append(lines, "  Description: "+item.Description)
		}
		if strings.TrimSpace(item.Priority) != "" {
			lines = append(lines, "  Priority: "+item.Priority)
		}
		if strings.TrimSpace(item.Rationale) != "" {
			lines = append(lines, "  Rationale: "+item.Rationale)
		}
		doc.MultiCell(0, 6, strings.Join(lines, "\n"), "", "L", false)
	}
	doc.Ln(2)
}

func writeInstrumentsSection(doc *fpdf.Fpdf, items []InstrumentEntry) {
	doc.SetFont("Arial", "B", 13)
	doc.Cell(0, 8, "Recommended Instruments")
	doc.Ln(9)

	doc.SetFont("Arial", "", 11)
	if len(items) == 0 {
		doc.MultiCell(0, 6, "- None", "", "L", false)
		doc.Ln(2)
		return
	}

	for _, item := range items {
		lines := []string{fmt.Sprintf("- %s", strings.TrimSpace(item.Name))}
		if strings.TrimSpace(item.Description) != "" {
			lines = append(lines, "  Description: "+item.Description)
		}
		if strings.TrimSpace(item.Priority) != "" {
			lines = append(lines, "  Priority: "+item.Priority)
		}
		if strings.TrimSpace(item.Rationale) != "" {
			lines = append(lines, "  Rationale: "+item.Rationale)
		}
		doc.MultiCell(0, 6, strings.Join(lines, "\n"), "", "L", false)
	}
	doc.Ln(2)
}

func writeSimpleSection(doc *fpdf.Fpdf, title string, items []string) {
	doc.SetFont("Arial", "B", 13)
	doc.Cell(0, 8, title)
	doc.Ln(9)

	doc.SetFont("Arial", "", 11)
	if len(items) == 0 {
		doc.MultiCell(0, 6, "- None", "", "L", false)
		doc.Ln(2)
		return
	}

	for _, item := range items {
		doc.MultiCell(0, 6, "- "+strings.TrimSpace(item), "", "L", false)
	}
	doc.Ln(2)
}

func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "None"
	}
	return value
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "None"
	}
	return strings.Join(items, ", ")
}
