package r2

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

type LocalStore struct {
	rootDir string
	baseURL string
}

func NewLocalStore(rootDir string, baseURL string) *LocalStore {
	return &LocalStore{
		rootDir: rootDir,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (s *LocalStore) Save(ctx context.Context, key string, data []byte) (string, error) {
	path := filepath.Join(s.rootDir, filepath.FromSlash(key))

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}

	return s.baseURL + "/" + strings.TrimLeft(key, "/"), nil
}

func (s *LocalStore) Read(ctx context.Context, key string) ([]byte, error) {
	path := filepath.Join(s.rootDir, filepath.FromSlash(key))
	return os.ReadFile(path)
}
