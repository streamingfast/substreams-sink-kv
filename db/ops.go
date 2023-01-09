package db

import (
	"context"
	"fmt"

	sink "github.com/streamingfast/substreams-sink"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
)

func (l *DB) AddOperations(ops *pbkv.KVOperations) {
	for _, op := range ops.Operations {
		l.AddOperation(op)
	}
}

func (l *DB) AddOperation(op *pbkv.KVOperation) {
	l.pendingOperations = append(l.pendingOperations, op)
}

func (l *DB) Flush(ctx context.Context, moduleHash string, cursor *sink.Cursor) (count int, err error) {
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

func userKey(k string) []byte {
	return []byte(fmt.Sprintf("k%s", k))
}

/*

map[key][]operation


var batchPut []operation
var batchDelete []operation

for k, v := range map {
	sort(v)

}

batchPUT

*/
