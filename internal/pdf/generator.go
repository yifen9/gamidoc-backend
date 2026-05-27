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

const (
	lMargin    = 20.0
	rMargin    = 20.0
	pageW      = 210.0
	usableW    = pageW - lMargin - rMargin // 170mm
	cardIndent = 5.0
	cardW      = usableW - cardIndent // 165mm
)

// color helpers
func setFill(doc *fpdf.Fpdf, r, g, b int) { doc.SetFillColor(r, g, b) }
func setDraw(doc *fpdf.Fpdf, r, g, b int) { doc.SetDrawColor(r, g, b) }
func setText(doc *fpdf.Fpdf, r, g, b int) { doc.SetTextColor(r, g, b) }

func (g *FPDFGenerator) Generate(data PlanData) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetMargins(lMargin, 20, rMargin)
	doc.SetAutoPageBreak(true, 18)
	doc.SetTitle(data.Title, false)
	doc.SetAuthor("GamiDoc", false)
	doc.SetCreator("GamiDoc", false)

	doc.SetFooterFunc(func() {
		doc.SetY(-14)
		setDraw(doc, 180, 200, 220)
		doc.Line(lMargin, doc.GetY(), pageW-rMargin, doc.GetY())
		doc.Ln(2)
		doc.SetFont("Arial", "", 8)
		setText(doc, 120, 120, 120)
		doc.CellFormat(usableW/2, 5, "GamiDoc - Evaluation Plan", "", 0, "L", false, 0, "")
		doc.CellFormat(usableW/2, 5, fmt.Sprintf("Page %d", doc.PageNo()), "", 0, "R", false, 0, "")
	})

	doc.AddPage()

	writeTitleBlock(doc, data)
	writeContextSection(doc, data)
	writeConstraintsSection(doc, data)
	if data.ResearchEnabled {
		writeResearchSection(doc, data)
	}
	writeMethodsSection(doc, data.SelectedMethods)
	writeInstrumentsSection(doc, data.SelectedInstruments)
	writeSimpleSection(doc, "Next Steps", data.NextSteps)
	if strings.TrimSpace(data.Notes) != "" {
		writeNotesSection(doc, data.Notes)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeTitleBlock(doc *fpdf.Fpdf, data PlanData) {
	// Navy banner — title
	setFill(doc, 42, 87, 141)
	setText(doc, 255, 255, 255)
	doc.SetFont("Arial", "B", 20)
	doc.CellFormat(0, 14, data.Title, "", 1, "C", true, 0, "")

	// Slightly lighter subtitle bar
	setFill(doc, 60, 110, 170)
	doc.SetFont("Arial", "", 10)
	doc.CellFormat(0, 7, "Evaluation Plan - "+data.Date.Format("January 2, 2006"), "", 1, "C", true, 0, "")

	setText(doc, 30, 30, 30)
	doc.Ln(7)
}

// sectionHeader draws a light-blue filled bar with a navy left+bottom border.
func sectionHeader(doc *fpdf.Fpdf, title string) {
	setFill(doc, 235, 243, 252)
	setDraw(doc, 42, 87, 141)
	setText(doc, 42, 87, 141)
	doc.SetFont("Arial", "B", 12)
	doc.CellFormat(0, 9, "  "+title, "LB", 1, "L", true, 0, "")
	setText(doc, 30, 30, 30)
	setDraw(doc, 200, 214, 229)
	doc.Ln(3)
}

// writeKV renders a small uppercase label then the value in normal text below it.
func writeKV(doc *fpdf.Fpdf, key, value string) {
	doc.SetFont("Arial", "B", 8)
	setText(doc, 110, 110, 110)
	doc.CellFormat(0, 5, strings.ToUpper(key), "", 1, "L", false, 0, "")
	doc.SetFont("Arial", "", 11)
	setText(doc, 30, 30, 30)
	doc.MultiCell(0, 6, value, "", "L", false)
	doc.Ln(2)
}

func writeContextSection(doc *fpdf.Fpdf, data PlanData) {
	sectionHeader(doc, "Evaluation Context")
	writeKV(doc, "Evaluation Goals", joinOrNone(data.EvaluationGoals))
	writeKV(doc, "Project Type", valueOrNone(data.ProjectType))
	writeKV(doc, "Participants", valueOrNone(data.Participants))
	writeKV(doc, "Development Stage", valueOrNone(data.DevelopmentStage))
}

func writeConstraintsSection(doc *fpdf.Fpdf, data PlanData) {
	hasAccessibility := strings.TrimSpace(data.Accessibility) != ""
	hasTime := strings.TrimSpace(data.TimeConstraint) != ""
	hasExtra := len(data.ExtraConstraints) > 0
	if !hasAccessibility && !hasTime && !hasExtra {
		return
	}

	sectionHeader(doc, "Constraints")
	if hasAccessibility {
		writeKV(doc, "User Accessibility", data.Accessibility)
	}
	if hasTime {
		writeKV(doc, "Available Time", data.TimeConstraint)
	}
	if hasExtra {
		writeKV(doc, "Additional Constraints", strings.Join(data.ExtraConstraints, "; "))
	}
}

func writeResearchSection(doc *fpdf.Fpdf, data PlanData) {
	sectionHeader(doc, "Research Specification")

	if strings.TrimSpace(data.ResearchObjective) != "" {
		writeKV(doc, "Research Objective", data.ResearchObjective)
	}

	if hasContent(data.ResearchQuestions) {
		doc.SetFont("Arial", "B", 8)
		setText(doc, 110, 110, 110)
		doc.CellFormat(0, 5, "RESEARCH QUESTIONS", "", 1, "L", false, 0, "")
		doc.Ln(1)
		for i, rq := range data.ResearchQuestions {
			if strings.TrimSpace(rq) != "" {
				writeNumberedItem(doc, fmt.Sprintf("RQ%d", i+1), rq)
			}
		}
		doc.Ln(2)
	}

	if hasContent(data.Hypotheses) {
		doc.SetFont("Arial", "B", 8)
		setText(doc, 110, 110, 110)
		doc.CellFormat(0, 5, "HYPOTHESES", "", 1, "L", false, 0, "")
		doc.Ln(1)
		for i, h := range data.Hypotheses {
			if strings.TrimSpace(h) != "" {
				writeNumberedItem(doc, fmt.Sprintf("H%d", i+1), h)
			}
		}
		doc.Ln(2)
	}
}

// writeNumberedItem renders "RQ1" / "H1" inline with the text.
// SetLeftMargin ensures wrapped lines align with the text start, not the label.
func writeNumberedItem(doc *fpdf.Fpdf, label, text string) {
	const labelW = 13.0

	doc.SetFont("Arial", "B", 10)
	setText(doc, 42, 87, 141)
	doc.CellFormat(labelW, 6, label, "", 0, "L", false, 0, "")

	doc.SetLeftMargin(lMargin + labelW)
	doc.SetFont("Arial", "", 10)
	setText(doc, 30, 30, 30)
	doc.MultiCell(usableW-labelW, 6, strings.TrimSpace(text), "", "L", false)
	doc.SetLeftMargin(lMargin)
	doc.SetX(lMargin)
	doc.Ln(1)
}

func writeMethodsSection(doc *fpdf.Fpdf, items []MethodEntry) {
	sectionHeader(doc, "Selected Evaluation Methods")
	if len(items) == 0 {
		writeEmpty(doc, "No methods selected.")
		return
	}
	for _, item := range items {
		writeEntryCard(doc, item.Name, item.Description, item.Priority, item.Rationale)
	}
}

func writeInstrumentsSection(doc *fpdf.Fpdf, items []InstrumentEntry) {
	sectionHeader(doc, "Recommended Instruments")
	if len(items) == 0 {
		writeEmpty(doc, "No instruments selected.")
		return
	}
	for _, item := range items {
		writeEntryCard(doc, item.Name, item.Description, item.Priority, item.Rationale)
	}
}

// writeEntryCard renders a method or instrument with a navy left accent bar.
// SetLeftMargin(lMargin+cardIndent) ensures MultiCell wrapping stays indented.
// The accent bar is skipped when content crosses a page break to avoid drawing
// a line that spans from the wrong Y coordinate on the new page.
func writeEntryCard(doc *fpdf.Fpdf, name, description, priority, rationale string) {
	startPage := doc.PageNo()
	startY := doc.GetY()
	doc.SetLeftMargin(lMargin + cardIndent)

	// Name
	doc.SetFont("Arial", "B", 11)
	setText(doc, 42, 87, 141)
	doc.MultiCell(cardW, 7, strings.TrimSpace(name), "", "L", false)

	// Description
	if strings.TrimSpace(description) != "" {
		doc.SetFont("Arial", "", 10)
		setText(doc, 30, 30, 30)
		doc.MultiCell(cardW, 5, strings.TrimSpace(description), "", "L", false)
	}

	// Priority | Rationale (italic, gray)
	var meta []string
	if strings.TrimSpace(priority) != "" {
		meta = append(meta, "Priority: "+strings.TrimSpace(priority))
	}
	if strings.TrimSpace(rationale) != "" {
		meta = append(meta, "Rationale: "+strings.TrimSpace(rationale))
	}
	if len(meta) > 0 {
		doc.SetFont("Arial", "I", 9)
		setText(doc, 110, 110, 110)
		doc.MultiCell(cardW, 5, strings.Join(meta, "  |  "), "", "L", false)
	}

	endY := doc.GetY()
	endPage := doc.PageNo()
	doc.SetLeftMargin(lMargin)
	doc.SetX(lMargin)

	// Left accent bar: only draw when the card stays on a single page.
	if startPage == endPage {
		setDraw(doc, 42, 87, 141)
		doc.Line(lMargin, startY, lMargin, endY)
	}

	// Bottom separator (pale blue-gray) — always drawn on the current page.
	setDraw(doc, 200, 214, 229)
	doc.Line(lMargin+cardIndent, endY, lMargin+cardIndent+cardW, endY)

	setText(doc, 30, 30, 30)
	doc.Ln(5)
}

func writeSimpleSection(doc *fpdf.Fpdf, title string, items []string) {
	sectionHeader(doc, title)
	if len(items) == 0 {
		writeEmpty(doc, "None specified.")
		return
	}
	const bulletW = 6.0
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		doc.SetFont("Arial", "B", 11)
		setText(doc, 42, 87, 141)
		doc.CellFormat(bulletW, 6, "-", "", 0, "R", false, 0, "")
		doc.SetLeftMargin(lMargin + bulletW)
		doc.SetFont("Arial", "", 11)
		setText(doc, 30, 30, 30)
		doc.MultiCell(usableW-bulletW, 6, strings.TrimSpace(item), "", "L", false)
		doc.SetLeftMargin(lMargin)
		doc.SetX(lMargin)
	}
	doc.Ln(2)
}

func writeNotesSection(doc *fpdf.Fpdf, notes string) {
	sectionHeader(doc, "Notes")
	doc.SetFont("Arial", "", 11)
	setText(doc, 30, 30, 30)
	doc.MultiCell(0, 6, strings.TrimSpace(notes), "", "L", false)
	doc.Ln(2)
}

func writeEmpty(doc *fpdf.Fpdf, msg string) {
	doc.SetFont("Arial", "I", 10)
	setText(doc, 130, 130, 130)
	doc.CellFormat(0, 6, msg, "", 1, "L", false, 0, "")
	setText(doc, 30, 30, 30)
	doc.Ln(2)
}

func hasContent(items []string) bool {
	for _, s := range items {
		if strings.TrimSpace(s) != "" {
			return true
		}
	}
	return false
}

func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ", ")
}
