package db

import (
	"context"

	sink "github.com/streamingfast/substreams-sink"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"
)

type Loader interface {
	AddOperations(ops *pbkv.KVOperations)
	AddOperation(op *pbkv.KVOperation)
	Flush(ctx context.Context, cursor *sink.Cursor) (count int, err error)
	GetCursor(ctx context.Context) (*sink.Cursor, error)
	WriteCursor(ctx context.Context, c *sink.Cursor) error
}

type Reader interface {
	Get(ctx context.Context, key string) (value []byte, err error)
	GetMany(ctx context.Context, keys []string) (values [][]byte, err error)
	GetByPrefix(ctx context.Context, prefix string, limit int) (values []*pbkv.KV, limitReached bool, err error)
	Scan(ctx context.Context, start string, exclusiveEnd string, limit int) (values []*pbkv.KV, limitReached bool, err error)
}
