package db

import (
	"context"

	"github.com/streamingfast/kvdb/store"
	"github.com/streamingfast/logging"
	sink "github.com/streamingfast/substreams-sink"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DBLoader interface {
	AddOperations(ops *kvv1.KVOperations)
	AddOperation(op *kvv1.KVOperation)
	Flush(ctx context.Context, cursor *sink.Cursor) (count int, err error)
	GetCursor(ctx context.Context) (*sink.Cursor, error)
	WriteCursor(ctx context.Context, c *sink.Cursor) error
}

type DBReader interface {
	Get(ctx context.Context, key string) (value []byte, err error)
	GetMany(ctx context.Context, keys []string) (values [][]byte, err error)
	GetByPrefix(ctx context.Context, prefix string, limit int) (values []*kvv1.KV, limitReached bool, err error)
	// Scan(ctx context.Context, start, exclusiveEnd []byte, limit int, options ...ReadOption) *Iterator
}

type CursorError struct {
	error
}

type DB struct {
	store store.KVStore

	QueryRowsLimit    int
	pendingOperations []*kvv1.KVOperation
	logger            *zap.Logger
	tracer            logging.Tracer
}

func New(dsn string, logger *zap.Logger, tracer logging.Tracer) (*DB, error) {
	s, err := store.New(dsn)
	if err != nil {
		return nil, err
	}
	return &DB{
		QueryRowsLimit: 1000,
		store:          s,
		logger:         logger,
		tracer:         tracer,
	}, nil
}

func (l *DB) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	//TODO implement me
	//encoder.AddUint64("entries_count", l.EntriesCount)
	return nil
}
