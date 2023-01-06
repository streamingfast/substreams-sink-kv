package db

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/kvdb/store"
	sink "github.com/streamingfast/substreams-sink"
)

var ErrCursorNotFound = errors.New("cursor not found")
var cursorPrefix = "xc"

func CursorKey(outputModuleHash string) []byte {
	return []byte(cursorPrefix + outputModuleHash)
}

func (l *Loader) GetCursor(ctx context.Context, moduleHash string) (*sink.Cursor, error) {
	val, err := l.store.Get(ctx, CursorKey(moduleHash))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, ErrCursorNotFound
		} else {
			fmt.Println("wtf is this", err)
		}
		return nil, err
	}
	return cursorFromBytes(val)
}

func (l *Loader) WriteCursor(ctx context.Context, moduleHash string, c *sink.Cursor) error {
	val := cursorToBytes(c)
	if err := l.store.Put(ctx, CursorKey(moduleHash), val); err != nil {
		return err
	}
	return l.store.FlushPuts(ctx)
}

func cursorToBytes(c *sink.Cursor) []byte {
	out := fmt.Sprintf("%s:%s:%d", c.Cursor, c.Block.ID(), c.Block.Num())
	return []byte(out)
}

func cursorFromBytes(in []byte) (*sink.Cursor, error) {
	parts := strings.Split(string(in), ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid cursor")
	}
	blockNum, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor")
	}

	return &sink.Cursor{
		Cursor: parts[0],
		Block:  bstream.NewBlockRef(parts[1], blockNum),
	}, nil
}
