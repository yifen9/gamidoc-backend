package pdf

import (
	"bytes"
	"errors"
	"html/template"
	"strings"
	"time"
)

const (
	maxHTMLTemplateBytes = 128 * 1024
	maxCSSBytes          = 64 * 1024
)

var ErrInvalidPDFTemplate = errors.New("invalid pdf template")

type CustomTemplate struct {
	HTML string `json:"html"`
	CSS  string `json:"css,omitempty"`
}

type GenerateOptions struct {
	NotifyEmail string
	Template    *CustomTemplate
}

func RenderCustomHTML(data PlanData, custom CustomTemplate) (string, error) {
	if err := validateCustomTemplate(custom); err != nil {
		return "", err
	}

	tpl, err := template.New("pdf").Funcs(template.FuncMap{
		"join":       strings.Join,
		"formatDate": formatDate,
		"default":    defaultString,
	}).Parse(custom.HTML)
	if err != nil {
		return "", ErrInvalidPDFTemplate
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", ErrInvalidPDFTemplate
	}

	html := strings.TrimSpace(buf.String())
	return embedCustomCSS(html, custom.CSS), nil
}

func validateCustomTemplate(custom CustomTemplate) error {
	html := strings.TrimSpace(custom.HTML)
	if html == "" {
		return ErrInvalidPDFTemplate
	}
	if len([]byte(custom.HTML)) > maxHTMLTemplateBytes {
		return ErrInvalidPDFTemplate
	}
	if len([]byte(custom.CSS)) > maxCSSBytes {
		return ErrInvalidPDFTemplate
	}

	combined := strings.ToLower(custom.HTML + "\n" + custom.CSS)
	blocked := []string{
		"<script",
		"</script",
		"javascript:",
		" onabort=",
		" onblur=",
		" onchange=",
		" onclick=",
		" onerror=",
		" onfocus=",
		" oninput=",
		" onload=",
		" onmouseover=",
		" onsubmit=",
	}
	for _, item := range blocked {
		if strings.Contains(combined, item) {
			return ErrInvalidPDFTemplate
		}
	}

	return nil
}

func embedCustomCSS(html string, css string) string {
	css = strings.TrimSpace(css)
	if css == "" {
		return ensureHTMLDocument(html, "")
	}

	style := "<style>\n" + css + "\n</style>"
	lower := strings.ToLower(html)
	if idx := strings.Index(lower, "</head>"); idx >= 0 {
		return html[:idx] + style + "\n" + html[idx:]
	}

	return ensureHTMLDocument(html, style)
}

func ensureHTMLDocument(html string, headContent string) string {
	lower := strings.ToLower(strings.TrimSpace(html))
	if strings.HasPrefix(lower, "<!doctype") || strings.HasPrefix(lower, "<html") {
		if strings.TrimSpace(headContent) != "" {
			if idx := strings.Index(strings.ToLower(html), "<html"); idx >= 0 {
				afterOpen := strings.Index(html[idx:], ">")
				if afterOpen >= 0 {
					insertAt := idx + afterOpen + 1
					return html[:insertAt] + "<head><meta charset=\"utf-8\">\n" + headContent + "</head>" + html[insertAt:]
				}
			}
		}
		return html
	}

	var head strings.Builder
	head.WriteString(`<meta charset="utf-8">`)
	if strings.TrimSpace(headContent) != "" {
		head.WriteString("\n")
		head.WriteString(headContent)
	}

	return "<!doctype html><html><head>" + head.String() + "</head><body>" + html + "</body></html>"
}

func formatDate(value time.Time, layout string) string {
	if layout == "" {
		layout = "January 2, 2006"
	}
	return value.Format(layout)
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
