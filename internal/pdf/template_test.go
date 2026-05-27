package pdf

import (
	"strings"
	"testing"
	"time"
)

func TestRenderCustomHTMLInjectsCSSAndData(t *testing.T) {
	html, err := RenderCustomHTML(PlanData{
		Title:           "Custom Plan",
		Date:            time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC),
		EvaluationGoals: []string{"Usability"},
	}, CustomTemplate{
		HTML: `<main><h1>{{ .Title }}</h1><p>{{ join .EvaluationGoals ", " }}</p><time>{{ formatDate .Date "2006-01-02" }}</time></main>`,
		CSS:  `h1 { color: #234567; }`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, want := range []string{
		"<!doctype html>",
		"<h1>Custom Plan</h1>",
		"Usability",
		"2026-05-27",
		"h1 { color: #234567; }",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered html to contain %q, got %s", want, html)
		}
	}
}

func TestRenderCustomHTMLRejectsScript(t *testing.T) {
	_, err := RenderCustomHTML(PlanData{}, CustomTemplate{
		HTML: `<script>alert("x")</script>`,
	})
	if err != ErrInvalidPDFTemplate {
		t.Fatalf("expected ErrInvalidPDFTemplate, got %v", err)
	}
}

func TestRenderCustomHTMLInjectsCSSIntoFullDocumentWithoutHead(t *testing.T) {
	html, err := RenderCustomHTML(PlanData{Title: "Full"}, CustomTemplate{
		HTML: `<html><body><h1>{{ .Title }}</h1></body></html>`,
		CSS:  `body { margin: 0; }`,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(html, `<head><meta charset="utf-8">`) {
		t.Fatalf("expected injected head, got %s", html)
	}
	if !strings.Contains(html, `body { margin: 0; }`) {
		t.Fatalf("expected injected css, got %s", html)
	}
}

func TestRenderCustomHTMLRejectsEmptyTemplate(t *testing.T) {
	_, err := RenderCustomHTML(PlanData{}, CustomTemplate{})
	if err != ErrInvalidPDFTemplate {
		t.Fatalf("expected ErrInvalidPDFTemplate, got %v", err)
	}
}
