package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/streamingfast/kvdb/store"
	sink "github.com/streamingfast/substreams-sink"
	"go.uber.org/zap"
)

var ErrCursorNotFound = errors.New("cursor not found")
var cursorKey = []byte("xc")

func (l *DB) GetCursor(ctx context.Context) (*sink.Cursor, error) {
	val, err := l.store.Get(ctx, cursorKey)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrCursorNotFound
		}
		return nil, err
	}

	l.logger.Debug("cursor found", zap.String("cursor", string(val)))

	return cursorFromBytes(val)
}

func (l *DB) WriteCursor(ctx context.Context, c *sink.Cursor) error {
	val := cursorToBytes(c)
	if err := l.store.Put(ctx, cursorKey, val); err != nil {
		return err
	}
	return l.store.FlushPuts(ctx)
}

func cursorToBytes(c *sink.Cursor) []byte {
	out := fmt.Sprintf("%s:%s:%d", c.String(), c.Block().ID(), c.Block().Num())
	return []byte(out)
}

func cursorFromBytes(in []byte) (*sink.Cursor, error) {
	parts := strings.Split(string(in), ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid cursor")
	}

	// We validate the parts but don't read them, all information is taken from the cursor itself
	_, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor")
	}

	return sink.NewCursor(parts[0])
}
