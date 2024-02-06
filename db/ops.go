package db

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/streamingfast/kvdb/store"
	sink "github.com/streamingfast/substreams-sink"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var ErrInvalidArguments = errors.New("invalid arguments")
var ErrNotFound = errors.New("not found")

// FIXME: open-ended scans need to be implemented in kvdb
var InfiniteEndBytes = []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

func (l *DB) AddOperations(ops *pbkv.KVOperations) {
	for _, op := range ops.Operations {
		l.AddOperation(op)
	}
}

func (l *DB) AddOperation(op *pbkv.KVOperation) {
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

func (l *DB) StoreReverseOperations(ctx context.Context, blockNumber uint64, ops []*pbkv.KVOperation) error {
	reversedKVOperations := l.reverseOperations(ctx, ops)
	encodedReversedOperations, err := proto.Marshal(reversedKVOperations)
	if err != nil {
		return fmt.Errorf("unable to marshal reversed operations: %w", err)
	}

	err = l.store.Put(ctx, undoKey(blockNumber), encodedReversedOperations)
	if err != nil {
		return fmt.Errorf("unable to store reversed operations: %w", err)
	}
	return nil
}

func (l *DB) reverseOperations(ctx context.Context, ops []*pbkv.KVOperation) *pbkv.KVOperations {
	reversedOperations := make([]*pbkv.KVOperation, len(ops))
	for _, operation := range ops {
		var reverseOperation *pbkv.KVOperation
		previousValue, errNotFound := l.store.Get(ctx, userKey(operation.Key))

		switch operation.Type {
		case pbkv.KVOperation_SET:
			if errNotFound != nil {
				reverseOperation = &pbkv.KVOperation{
					Type:  pbkv.KVOperation_DELETE,
					Key:   operation.Key,
					Value: operation.Value,
				}
			} else {
				reverseOperation = &pbkv.KVOperation{
					Type:  pbkv.KVOperation_SET,
					Key:   operation.Key,
					Value: previousValue,
				}
			}
		case pbkv.KVOperation_DELETE:
			reverseOperation = &pbkv.KVOperation{
				Type:  pbkv.KVOperation_SET,
				Key:   operation.Key,
				Value: operation.Value,
			}

		case pbkv.KVOperation_UNSET:
			panic("Not implemented")
		}

		reversedOperations = append(reversedOperations, reverseOperation)
	}

	reversedKVOperations := &pbkv.KVOperations{Operations: reversedOperations}
	return reversedKVOperations
}

func (l *DB) HandleBlockUndo(ctx context.Context, lastValidBlock uint64, cursor *sink.Cursor) error {
	blockNumber := cursor.Block().Num() //Set the cursor to the current block number

	for blockNumber > lastValidBlock {
		encodedOperations, ErrNotFound := l.store.Get(ctx, undoKey(blockNumber))
		kvOperations := &pbkv.KVOperations{}

		if ErrNotFound != nil {
			return fmt.Errorf("unable to find undo operations for block %d: %w", blockNumber, ErrNotFound)
		}

		err := proto.Unmarshal(encodedOperations, kvOperations)
		if err != nil {
			return fmt.Errorf("unable to unmarshal undo operations for block %d: %w", blockNumber, err)
		}

		l.AddOperations(kvOperations)

		_, err = l.Flush(ctx, cursor)
		if err != nil {
			return fmt.Errorf("unable to flush undo operations for block %d: %w", blockNumber, err)
		}

		blockNumber -= 1
	}

	return nil
}

func lastOperationPerKey(ops []*pbkv.KVOperation) (puts []*pbkv.KVOperation, deletes [][]byte) {
	opsPerKey := make(map[string][]*pbkv.KVOperation)

	for _, op := range ops {
		opsPerKey[op.Key] = append(opsPerKey[op.Key], op)
	}

	for _, ops := range opsPerKey {
		//		sortByOrdinal(ops)
		lastOp := ops[len(ops)-1]
		switch lastOp.Type {
		case pbkv.KVOperation_SET:
			puts = append(puts, lastOp)
		case pbkv.KVOperation_DELETE:
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
	out := make([]byte, len(k)+1)
	out[0] = 'k'
	copy(out[1:], k)
	return out
}

func undoKey(num uint64) []byte {
	numBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(numBytes, math.MaxUint64-num)
	return append([]byte{'u'}, numBytes...)
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
