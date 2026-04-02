package pdf

import (
	"bytes"
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

	writeSection(doc, "Evaluation Goals", data.EvaluationGoals)
	writeSection(doc, "Selected Methods", data.SelectedMethods)
	writeSection(doc, "Recommended Instruments", data.RecommendedInstruments)
	writeSection(doc, "Next Steps", data.NextSteps)

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func writeSection(doc *fpdf.Fpdf, title string, items []string) {
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
