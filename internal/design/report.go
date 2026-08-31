package design

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
)

type RenderSection struct {
	Number int
	Name   string
	Lines  []string
}

type ReportBuilder struct{}

func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{}
}

func (b *ReportBuilder) Build(title string, version string, spark string, sections []RenderSection) ([]byte, error) {
	doc := fpdf.New("P", "mm", "A4", "")
	doc.SetTitle(title, false)
	doc.AddPage()

	doc.SetFont("Arial", "B", 16)
	doc.MultiCell(0, 8, title, "", "L", false)
	doc.Ln(2)

	doc.SetFont("Arial", "", 10)
	doc.Cell(0, 6, "Version: "+version)
	doc.Ln(6)
	doc.Cell(0, 6, "Date: "+time.Now().UTC().Format("2006-01-02 15:04:05"))
	doc.Ln(10)

	if strings.TrimSpace(spark) != "" {
		doc.SetFont("Arial", "B", 13)
		doc.Cell(0, 8, "The Spark")
		doc.Ln(8)
		doc.SetFont("Arial", "", 11)
		doc.MultiCell(0, 6, spark, "", "L", false)
		doc.Ln(4)
	}

	for _, section := range sections {
		doc.SetFont("Arial", "B", 13)
		doc.Cell(0, 8, fmt.Sprintf("%d. %s", section.Number, section.Name))
		doc.Ln(8)
		doc.SetFont("Arial", "", 11)
		if len(section.Lines) == 0 {
			doc.MultiCell(0, 6, "Not provided.", "", "L", false)
		}
		for _, line := range section.Lines {
			doc.MultiCell(0, 6, line, "", "L", false)
		}
		doc.Ln(4)
	}

	var buffer bytes.Buffer
	if err := doc.Output(&buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func SectionLines(content json.RawMessage) []string {
	if len(content) == 0 {
		return nil
	}

	decoder := json.NewDecoder(bytes.NewReader(content))
	token, err := decoder.Token()
	if err != nil {
		return []string{string(content)}
	}

	if delim, ok := token.(json.Delim); ok && delim == '{' {
		var lines []string
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return []string{string(content)}
			}
			var value any
			if err := decoder.Decode(&value); err != nil {
				return []string{string(content)}
			}
			lines = append(lines, fmt.Sprintf("%v: %s", keyToken, flattenValue(value)))
		}
		return lines
	}

	var text string
	if err := json.Unmarshal(content, &text); err == nil {
		return []string{text}
	}

	return []string{string(content)}
}

func SectionPlainText(content json.RawMessage) string {
	return strings.Join(SectionLines(content), "\n")
}

func flattenValue(value any) string {
	switch typed := value.(type) {
	case []any:
		var parts []string
		for _, item := range typed {
			parts = append(parts, flattenValue(item))
		}
		return strings.Join(parts, ", ")
	case map[string]any:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	default:
		return fmt.Sprint(typed)
	}
}
