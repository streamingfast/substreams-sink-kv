package db

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/streamingfast/substreams-sink-kv/sinker"
	"math"
	"time"

	"github.com/streamingfast/kvdb/store"
	sink "github.com/streamingfast/substreams-sink"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var undoPrefix = []byte{'x', 'u'}

var ErrInvalidArguments = errors.New("invalid arguments")
var ErrNotFound = errors.New("not found")

// FIXME: open-ended scans need to be implemented in kvdb
var InfiniteEndBytes = []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

func (db *DB) AddOperations(ops *pbkv.KVOperations) {
	for _, op := range ops.Operations {
		db.AddOperation(op)
	}
}

func (db *DB) AddOperation(op *pbkv.KVOperation) {
	db.pendingOperations = append(db.pendingOperations, op)
}

func (db *DB) Flush(ctx context.Context, cursor *sink.Cursor) (count int, err error) {
	puts, deletes := lastOperationPerKey(db.pendingOperations)
	for _, put := range puts {
		if err := db.store.Put(ctx, userKey(put.Key), put.Value); err != nil {
			return 0, err
		}
	}

	if err := db.store.BatchDelete(ctx, deletes); err != nil {
		return 0, err
	}

	if err := db.WriteCursor(ctx, cursor); err != nil {
		return 0, err
	}
	db.reset()
	return len(puts) + len(deletes), nil
}

func (db *DB) HandleOperations(ctx context.Context, blockNumber uint64, kvOps *kvv1.KVOperations, cursor *sink.Cursor, batchModulo uint64, s *sinker.KVSinker) (flushDone bool, error) {
	err := db.StoreReverseOperations(ctx, blockNumber, kvOps.Operations)
	if err != nil {
		return false, fmt.Errorf("storing reverse operations: %w", err)
	}

	db.AddOperations(kvOps)

	blockRef := cursor.Block()
	if blockRef.Num()%batchModulo == 0 {
		flushStart := time.Now()
		count, err := db.Flush(ctx, cursor)
		if err != nil {
			return false, fmt.Errorf("failed to flush: %w", err)
		}

		sinker.FlushCount.Inc()
		sinker.FlushedEntriesCount.AddInt(count)
		sinker.FlushedEntriesCount.AddInt64(time.Since(flushStart).Nanoseconds())
	}
}
func (db *DB) StoreReverseOperations(ctx context.Context, blockNumber uint64, ops []*pbkv.KVOperation) error {
	reversedKVOperations := db.reverseOperations(ctx, ops)
	encodedReversedOperations, err := proto.Marshal(reversedKVOperations)
	if err != nil {
		return fmt.Errorf("unable to marshal reversed operations: %w", err)
	}

	err = db.store.Put(ctx, undoKey(blockNumber), encodedReversedOperations)
	if err != nil {
		return fmt.Errorf("unable to store reversed operations: %w", err)
	}
	return nil
}

func (db *DB) reverseOperations(ctx context.Context, ops []*pbkv.KVOperation) *pbkv.KVOperations {
	var reversedOperations []*pbkv.KVOperation
	for _, op := range ops {
		var reverseOperation *pbkv.KVOperation
		previousValue, errNotFound := db.store.Get(ctx, userKey(op.Key))

		switch op.Type {
		case pbkv.KVOperation_SET:
			if errNotFound != nil {
				reverseOperation = &pbkv.KVOperation{
					Type:  pbkv.KVOperation_DELETE,
					Key:   op.Key,
					Value: op.Value,
				}
				break
			}
			reverseOperation = &pbkv.KVOperation{
				Type:  pbkv.KVOperation_SET,
				Key:   op.Key,
				Value: previousValue,
			}
		case pbkv.KVOperation_DELETE:
			if errNotFound != nil {
				break
			}
			reverseOperation = &pbkv.KVOperation{
				Type:  pbkv.KVOperation_SET,
				Key:   op.Key,
				Value: op.Value,
			}

		case pbkv.KVOperation_UNSET:
			panic("Missing valid op")
		}

		reversedOperations = append([]*pbkv.KVOperation{reverseOperation}, reversedOperations...)
	}

	reversedKVOperations := &pbkv.KVOperations{Operations: reversedOperations}
	return reversedKVOperations
}

func (db *DB) HandleBlockUndo(ctx context.Context, lastValidBlock uint64) error {
	scanResult := db.store.Scan(ctx, undoKey(math.MaxUint64), undoKey(lastValidBlock), 0)
	if scanResult.Err() != nil {
		return fmt.Errorf("scanning undo operations for block %d: %w", lastValidBlock, scanResult.Err())
	}
	kvOperations := &pbkv.KVOperations{}
	var encodedOperations []byte
	for scanResult.Next() {
		encodedOperations = scanResult.Item().Value

		err := proto.Unmarshal(encodedOperations, kvOperations)
		if err != nil {
			return fmt.Errorf("unmarshaling undo operations: %w", err)
		}
		db.AddOperations(kvOperations)
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

func (db *DB) reset() {
	db.pendingOperations = nil
}

func (db *DB) Get(ctx context.Context, key string) (val []byte, err error) {
	val, err = db.store.Get(ctx, userKey(key))
	if err != nil && errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return
}

func (db *DB) GetMany(ctx context.Context, keys []string) (values [][]byte, err error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("%w: you must specify at least one key", ErrInvalidArguments)
	}
	userKeys := make([][]byte, len(keys))
	for i := range keys {
		userKeys[i] = userKey(keys[i])
	}

	itr := db.store.BatchGet(ctx, userKeys)
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

func (db *DB) GetByPrefix(ctx context.Context, prefix string, limit int) (values []*kvv1.KV, limitReached bool, err error) {
	if limit == 0 {
		limit = db.QueryRowsLimit
	}
	if limit < 0 || limit > db.QueryRowsLimit {
		return nil, false, fmt.Errorf("%w: request value for 'limit' must be between 1 and %d, but received %d", ErrInvalidArguments, db.QueryRowsLimit, limit)
	}
	if prefix == "" {
		return nil, false, fmt.Errorf("%w: request value for 'prefix' must not be empty", ErrInvalidArguments)
	}

	itr := db.store.Prefix(ctx, userKey(prefix), limit+1)
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

func (db *DB) Scan(ctx context.Context, begin, exclusiveEnd string, limit int) (values []*kvv1.KV, limitReached bool, err error) {
	if limit == 0 {
		limit = db.QueryRowsLimit
	}
	if limit < 0 || limit > db.QueryRowsLimit {
		return nil, false, fmt.Errorf("%w: request value for 'limit' must be between 1 and %d, but received %d", ErrInvalidArguments, db.QueryRowsLimit, limit)
	}

	endBytes := InfiniteEndBytes
	if exclusiveEnd != "" {
		endBytes = userKey(exclusiveEnd)
	}
	itr := db.store.Scan(ctx, userKey(begin), endBytes, limit+1)
	for itr.Next() {
		if !isUserKey(itr.Item().Key) {
			db.logger.Debug("skipping non-user-key", zap.String("key", string(itr.Item().Key)))
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
	return append(undoPrefix, numBytes...)
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
