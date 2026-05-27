package pdf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGotenbergHTMLRendererPostsMultipartHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := r.ParseMultipartForm(1024 * 1024); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		files := r.MultipartForm.File["files"]
		if len(files) != 1 {
			t.Fatalf("expected one html file, got %d", len(files))
		}
		if files[0].Filename != "index.html" {
			t.Fatalf("expected index.html, got %q", files[0].Filename)
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF rendered"))
	}))
	defer server.Close()

	renderer := NewGotenbergHTMLRenderer(server.URL, time.Second)
	data, err := renderer.Render(context.Background(), HTMLDocument{HTML: "<h1>Hello</h1>"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(data) != "%PDF rendered" {
		t.Fatalf("unexpected pdf bytes: %q", string(data))
	}
}

func TestGotenbergHTMLRendererReportsHTTPFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "render failed", http.StatusBadGateway)
	}))
	defer server.Close()

	renderer := NewGotenbergHTMLRenderer(server.URL, time.Second)
	_, err := renderer.Render(context.Background(), HTMLDocument{HTML: "<h1>Hello</h1>"})
	if err == nil || !strings.Contains(err.Error(), "502") {
		t.Fatalf("expected status error, got %v", err)
	}
}
