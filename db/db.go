package db

import (
	"github.com/streamingfast/kvdb/store"
	"github.com/streamingfast/logging"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CursorError struct {
	error
}

var _ Loader = (*DB)(nil)
var _ Reader = (*DB)(nil)

type DB struct {
	store store.KVStore

	QueryRowsLimit    int
	pendingOperations []*pbkv.KVOperation
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
