package db

import (
	"context"

	sink "github.com/streamingfast/substreams-sink"
)

func (l *Loader) Flush(ctx context.Context, moduleHash string, cursor *sink.Cursor) (err error) {
	// not implemented
	return nil
}

func (l *Loader) reset() {}
