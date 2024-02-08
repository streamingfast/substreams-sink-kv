package db

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/kvdb"
	"github.com/streamingfast/kvdb/store"
	"github.com/streamingfast/logging"
	sink "github.com/streamingfast/substreams-sink"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"
)

type CursorError struct {
	error
}

var _ Reader = (*OperationDB)(nil)

type OperationDB struct {
	store store.KVStore

	QueryRowsLimit    int
	pendingOperations []*pbkv.KVOperation
	logger            *zap.Logger
	tracer            logging.Tracer
}

func New(dsn string, queryRowsLimit int, logger *zap.Logger, tracer logging.Tracer) (*OperationDB, error) {
	s, err := store.New(dsn)
	if err != nil {
		return nil, err
	}
	return &OperationDB{
		QueryRowsLimit: queryRowsLimit,
		store:          s,
		logger:         logger,
		tracer:         tracer,
	}, nil
}

func (db *OperationDB) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	//TODO implement me
	//encoder.AddUint64("entries_count", db.EntriesCount)
	return nil
}

var undoPrefix = [2]byte{'x', 'u'}
var userKeyPrefix byte = 'k'

var ErrInvalidArguments = errors.New("invalid arguments")
var ErrNotFound = errors.New("not found")

// FIXME: open-ended scans need to be implemented in kvdb
var InfiniteEndBytes = []byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}

func (db *OperationDB) AddOperations(ops *pbkv.KVOperations) {
	for _, op := range ops.Operations {
		db.AddOperation(op)
	}
}

func (db *OperationDB) AddOperation(op *pbkv.KVOperation) {
	// Make this a MAP, and keep always the latest KVOperation.
	db.pendingOperations = append(db.pendingOperations, op)
}

func (db *OperationDB) Flush(ctx context.Context, cursor *sink.Cursor) (count int, err error) {
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

func (db *OperationDB) HandleOperations(ctx context.Context, blockNumber, finalBlockHeight uint64, step bstream.StepType, kvOps *pbkv.KVOperations) error {
	if step == bstream.StepNew {
		err := db.DeleteLIBUndoOperations(ctx, finalBlockHeight)
		if err != nil {
			return fmt.Errorf("deleting LIB undo operations: %w", err)
		}
		err = db.storeUndoOperations(ctx, blockNumber, kvOps.Operations)
		if err != nil {
			return fmt.Errorf("storing reverse operations: %w", err)
		}
	}

	db.AddOperations(kvOps)

	return nil
}

func (db *OperationDB) DeleteLIBUndoOperations(ctx context.Context, finalBlockHeight uint64) error {
	keys := make([][]byte, 0)

	scanOutput := db.store.Scan(ctx, undoKey(finalBlockHeight), undoKey(0), 0)

	if scanOutput.Err() != nil {
		return fmt.Errorf("scanning undo operations for block %d: %w", finalBlockHeight, scanOutput.Err())
	}

	for scanOutput.Next() {
		keys = append(keys, scanOutput.Item().Key)
	}

	return db.store.BatchDelete(ctx, keys)
}

func (db *OperationDB) storeUndoOperations(ctx context.Context, blockNumber uint64, ops []*pbkv.KVOperation) error {
	undoOperations := generateUndoOperations(ctx, ops, db.store)
	data, err := proto.Marshal(undoOperations)
	if err != nil {
		return fmt.Errorf("unable to marshal reversed operations: %w", err)
	}

	err = db.store.Put(ctx, undoKey(blockNumber), data)
	if err != nil {
		return fmt.Errorf("storing reversed operations: %w", err)
	}

	err = db.store.FlushPuts(ctx)
	if err != nil {
		return fmt.Errorf("flushing undo put: %w", err)
	}

	return nil
}

func generateUndoOperations(ctx context.Context, ops []*pbkv.KVOperation, kvStore store.KVStore) *pbkv.KVOperations {
	var undoOperations []*pbkv.KVOperation
	for _, op := range ops {
		previousValue, err := kvStore.Get(ctx, userKey(op.Key))
		previousKeyExists := err == kvdb.ErrNotFound
		if err != nil {
			return nil, err
		}
		undoOp := undoOperation(op, previousValue, previousKeyExists)
		undoOperations = append([]*pbkv.KVOperation{undoOp}, undoOperations...)
	}
	reversedKVOperations := &pbkv.KVOperations{Operations: undoOperations}
	return reversedKVOperations
}

func undoOperation(op *pbkv.KVOperation, previousValue []byte, previousKeyExists bool) *pbkv.KVOperation {
	switch op.Type {
	case pbkv.KVOperation_SET:
		if previousKeyExists { // if previous value was not notFound, we need to delete the key when applying the undo
			return &pbkv.KVOperation{
				Type:  pbkv.KVOperation_DELETE,
				Key:   op.Key,
				Value: op.Value,
			}
		}
		return &pbkv.KVOperation{
			Type:  pbkv.KVOperation_SET,
			Key:   op.Key,
			Value: previousValue,
		}
	case pbkv.KVOperation_DELETE:
		if previousKeyExists {
			return nil
		}
		return &pbkv.KVOperation{
			Type:  pbkv.KVOperation_SET,
			Key:   op.Key,
			Value: op.Value,
		}
	default:
		panic(fmt.Sprintf("invalid operation type %d", op.Type))
	}
}

func (db *OperationDB) HandleBlockUndo(ctx context.Context, lastValidBlock uint64) error {
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

func (db *OperationDB) DeleteUndoKeys(ctx context.Context, keys [][]byte) error {
	return db.store.BatchDelete(ctx, keys)
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

func (db *OperationDB) reset() {
	db.pendingOperations = nil
}

func (db *OperationDB) Get(ctx context.Context, key string) (val []byte, err error) {
	val, err = db.store.Get(ctx, userKey(key))
	if err != nil && errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return
}

func (db *OperationDB) GetMany(ctx context.Context, keys []string) (values [][]byte, err error) {
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

func (db *OperationDB) GetByPrefix(ctx context.Context, prefix string, limit int) (values []*pbkv.KV, limitReached bool, err error) {
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
		values = append(values, &pbkv.KV{
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

func (db *OperationDB) Scan(ctx context.Context, begin, exclusiveEnd string, limit int) (values []*pbkv.KV, limitReached bool, err error) {
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
		values = append(values, &pbkv.KV{
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
	out[0] = userKeyPrefix
	copy(out[1:], k)
	return out
}

func undoKey(num uint64) []byte {
	numBytes := make([]byte, 8+len(undoPrefix))
	copy(numBytes, undoPrefix[:])
	binary.BigEndian.PutUint64(numBytes[2:], math.MaxUint64-num)
	return numBytes
}

func isUserKey(k []byte) bool {
	if len(k) > 1 && k[0] == userKeyPrefix {
		return true
	}
	return false
}

func fromUserKey(k []byte) string {
	// skip the `userKeyPrefix`
	return string(k[1:])
}
