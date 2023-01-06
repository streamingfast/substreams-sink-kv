package db

import (
	"github.com/streamingfast/kvdb/store"
	"github.com/streamingfast/logging"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CursorError struct {
	error
}

type Loader struct {
	store store.KVStore

	pendingOperations []*pbkv.KVOperation
	logger            *zap.Logger
	tracer            logging.Tracer
}

func NewLoader(dsn string, logger *zap.Logger, tracer logging.Tracer) (*Loader, error) {
	s, err := store.New(dsn)
	if err != nil {
		return nil, err
	}
	return &Loader{
		store:  s,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (l *Loader) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	//TODO implement me
	//encoder.AddUint64("entries_count", l.EntriesCount)
	return nil
}
