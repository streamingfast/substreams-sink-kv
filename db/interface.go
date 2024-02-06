package db

import (
	"context"

	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
)

type Reader interface {
	Get(ctx context.Context, key string) (value []byte, err error)
	GetMany(ctx context.Context, keys []string) (values [][]byte, err error)
	GetByPrefix(ctx context.Context, prefix string, limit int) (values []*kvv1.KV, limitReached bool, err error)
	Scan(ctx context.Context, start string, exclusiveEnd string, limit int) (values []*kvv1.KV, limitReached bool, err error)
}
