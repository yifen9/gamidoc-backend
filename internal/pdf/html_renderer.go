package pdf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

var ErrPDFRendererUnavailable = errors.New("pdf html renderer unavailable")

type HTMLDocument struct {
	HTML string
}

type HTMLRenderer interface {
	Render(ctx context.Context, document HTMLDocument) ([]byte, error)
}

type GotenbergHTMLRenderer struct {
	endpoint string
	client   *http.Client
}

func NewGotenbergHTMLRenderer(endpoint string, timeout time.Duration) *GotenbergHTMLRenderer {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &GotenbergHTMLRenderer{
		endpoint: strings.TrimSpace(endpoint),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (r *GotenbergHTMLRenderer) Render(ctx context.Context, document HTMLDocument) ([]byte, error) {
	if r == nil || strings.TrimSpace(r.endpoint) == "" {
		return nil, ErrPDFRendererUnavailable
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write([]byte(document.HTML)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Accept", "application/pdf")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("html renderer returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if len(data) == 0 {
		return nil, errors.New("html renderer returned empty pdf")
	}

	return data, nil
}
