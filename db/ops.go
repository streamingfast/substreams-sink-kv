package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/streamingfast/kvdb/store"
	sink "github.com/streamingfast/substreams-sink"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"
	"go.uber.org/zap"
)

var ErrInvalidArguments = errors.New("invalid arguments")
var ErrNotFound = errors.New("not found")

// FIXME: open-ended scans need to be implemented in kvdb
var InfiniteEndBytes = []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

func (l *DB) AddOperations(ops *kvv1.KVOperations) {
	for _, op := range ops.Operations {
		l.AddOperation(op)
	}
}

func (l *DB) AddOperation(op *kvv1.KVOperation) {
	l.pendingOperations = append(l.pendingOperations, op)
}

func (l *DB) Flush(ctx context.Context, cursor *sink.Cursor) (count int, err error) {
	puts, deletes := lastOperationPerKey(l.pendingOperations)
	for _, put := range puts {
		if err := l.store.Put(ctx, userKey(put.Key), put.Value); err != nil {
			return 0, err
		}
	}

	if err := l.store.BatchDelete(ctx, deletes); err != nil {
		return 0, err
	}

	if err := l.WriteCursor(ctx, cursor); err != nil {
		return 0, err
	}
	l.reset()
	return len(puts) + len(deletes), nil
}

func lastOperationPerKey(ops []*kvv1.KVOperation) (puts []*kvv1.KVOperation, deletes [][]byte) {
	opsPerKey := make(map[string][]*kvv1.KVOperation)

	for _, op := range ops {
		opsPerKey[op.Key] = append(opsPerKey[op.Key], op)
	}

	for _, ops := range opsPerKey {
		//		sortByOrdinal(ops)
		lastOp := ops[len(ops)-1]
		switch lastOp.Type {
		case kvv1.KVOperation_SET:
			puts = append(puts, lastOp)
		case kvv1.KVOperation_DELETE:
			deletes = append(deletes, userKey(lastOp.Key))
		}
	}

	return
}

func (l *DB) reset() {
	l.pendingOperations = nil
}

func (l *DB) Get(ctx context.Context, key string) (val []byte, err error) {
	val, err = l.store.Get(ctx, userKey(key))
	if err != nil && errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return
}

func (l *DB) GetMany(ctx context.Context, keys []string) (values [][]byte, err error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("%w: you must specify at least one key", ErrInvalidArguments)
	}
	userKeys := make([][]byte, len(keys))
	for i := range keys {
		userKeys[i] = userKey(keys[i])
	}

	itr := l.store.BatchGet(ctx, userKeys)
	for itr.Next() {
		values = append(values, itr.Item().Value)
	}
	if err := itr.Err(); err != nil {
		if err != nil && errors.Is(err, store.ErrNotFound) {
			return nil, ErrNotFound
		}
	}
	return values, nil
}

func (l *DB) GetByPrefix(ctx context.Context, prefix string, limit int) (values []*kvv1.KV, limitReached bool, err error) {
	if limit == 0 {
		limit = l.QueryRowsLimit
	}
	if limit < 0 || limit > l.QueryRowsLimit {
		return nil, false, fmt.Errorf("%w: request value for 'limit' must be between 1 and %d, but received %d", ErrInvalidArguments, l.QueryRowsLimit, limit)
	}
	if prefix == "" {
		return nil, false, fmt.Errorf("%w: request value for 'prefix' must not be empty", ErrInvalidArguments)
	}

	itr := l.store.Prefix(ctx, userKey(prefix), limit+1)
	for itr.Next() {
		if len(values) == limit {
			limitReached = true
			break
		}
		it := itr.Item()
		// it.Key must be userKey because it matches prefix userKey(...)
		values = append(values, &kvv1.KV{
			Key:   fromUserKey(it.Key),
			Value: it.Value,
		})
	}
	if err := itr.Err(); err != nil {
		return nil, false, err
	}
	if len(values) == 0 {
		return nil, false, ErrNotFound
	}
	return values, limitReached, nil
}

func (l *DB) Scan(ctx context.Context, begin, exclusiveEnd string, limit int) (values []*kvv1.KV, limitReached bool, err error) {
	if limit == 0 {
		limit = l.QueryRowsLimit
	}
	if limit < 0 || limit > l.QueryRowsLimit {
		return nil, false, fmt.Errorf("%w: request value for 'limit' must be between 1 and %d, but received %d", ErrInvalidArguments, l.QueryRowsLimit, limit)
	}

	endBytes := InfiniteEndBytes
	if exclusiveEnd != "" {
		endBytes = userKey(exclusiveEnd)
	}
	itr := l.store.Scan(ctx, userKey(begin), endBytes, limit+1)
	for itr.Next() {
		if !isUserKey(itr.Item().Key) {
			l.logger.Debug("skipping non-user-key", zap.String("key", string(itr.Item().Key)))
			continue // skip keys that are not valid user keys
		}
		if len(values) == limit {
			limitReached = true
			break
		}
		it := itr.Item()
		values = append(values, &kvv1.KV{
			Key:   fromUserKey(it.Key),
			Value: it.Value,
		})
	}
	if err := itr.Err(); err != nil {
		return nil, false, err
	}
	if len(values) == 0 {
		return nil, false, ErrNotFound
	}
	return values, limitReached, nil
}

func userKey(k string) []byte {
	return []byte(fmt.Sprintf("k%s", k))
}

func isUserKey(k []byte) bool {
	if len(k) > 1 && k[0] == 'k' {
		return true
	}
	return false
}

func fromUserKey(k []byte) string {
	return string(k[1:])
}
