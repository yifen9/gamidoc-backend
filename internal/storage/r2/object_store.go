package r2

import "context"

type ObjectStore interface {
	Save(ctx context.Context, key string, data []byte) (string, error)
	Read(ctx context.Context, key string) ([]byte, error)
}
