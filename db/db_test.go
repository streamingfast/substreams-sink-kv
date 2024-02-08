package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/streamingfast/bstream"
	_ "github.com/streamingfast/kvdb/store/badger3"
	"github.com/streamingfast/logging"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"github.com/test-go/testify/require"
	"go.uber.org/zap"
)

func TestDB_HandleOperations(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/substreams-sink-kv-db"

	_, tracer := logging.PackageLogger("db", "github.com/streamingfast/substreams-sink-kv/db.test")

	db, err := New(fmt.Sprintf("badger3://%s", dbPath), 0, zap.NewNop(), tracer)
	require.NoError(t, err)

	type blockOperations struct {
		blockNumber      uint64
		operations       *pbkv.KVOperations
		finalBlockHeight uint64
	}
	cases := []struct {
		name                 string
		blocks               []blockOperations
		expectedRemainingKey [][]byte
	}{
		{
			name: "sunny path",
			blocks: []blockOperations{
				{
					blockNumber: 3,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.1",
								Value: []byte("value.1"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
			},
			expectedRemainingKey: [][]byte{userKey("key.1"), []byte("xc"), undoKey(3)},
		},

		{
			name: "undo deletion",
			blocks: []blockOperations{
				{
					blockNumber: 1,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.3",
								Value: []byte("value.3"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
				{
					blockNumber: 2,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.2",
								Value: []byte("value.2"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
				{
					blockNumber: 3,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.1",
								Value: []byte("value.1"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
			},
			expectedRemainingKey: [][]byte{userKey("key.1"), userKey("key.2"), userKey("key.3"), []byte("xc"), undoKey(3), undoKey(2)},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			//delete dbPath if it exists
			err = os.RemoveAll(dbPath)
			require.NoError(t, err)

			for _, block := range c.blocks {
				err = db.HandleOperations(ctx, block.blockNumber, block.finalBlockHeight, bstream.StepNew, block.operations)
				require.NoError(t, err)
				_, err = db.Flush(ctx, nil)
				require.NoError(t, err)
			}

			require.NoError(t, err)
			scanOutput := db.store.Scan(ctx, userKey(""), []byte{'y'}, 0)

			if scanOutput.Err() != nil {
				t.Fatal(scanOutput.Err())
			}

			var availableKeys [][]byte
			for scanOutput.Next() {
				availableKeys = append(availableKeys, scanOutput.Item().Key)
			}

			expectedStringKeys := make([]string, len(c.expectedRemainingKey))
			for i, key := range c.expectedRemainingKey {
				expectedStringKeys[i] = string(key)
			}

			availableKeysString := make([]string, len(availableKeys))
			for i, key := range availableKeys {
				availableKeysString[i] = string(key)
			}
			require.Equal(t, expectedStringKeys, availableKeysString)
		})
	}
}
func TestDB_HandleUndo(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/substreams-sink-kv-db"

	_, tracer := logging.PackageLogger("db", "github.com/streamingfast/substreams-sink-kv/db.test2")

	db, err := New(fmt.Sprintf("badger3://%s", dbPath), 0, zap.NewNop(), tracer)
	require.NoError(t, err)

	type blockOperations struct {
		blockNumber      uint64
		operations       *pbkv.KVOperations
		finalBlockHeight uint64
	}
	cases := []struct {
		name           string
		blocks         []blockOperations
		lastValidBlock uint64
		expectedKV     map[string]string
	}{
		{
			name: "sunny path",
			blocks: []blockOperations{
				{
					blockNumber: 3,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.1",
								Value: []byte("value.1"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
			},
			lastValidBlock: 2,
			expectedKV:     map[string]string{},
		},

		{
			name: "deletion until last valid block",
			blocks: []blockOperations{
				{
					blockNumber: 4,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.4",
								Value: []byte("value.4"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
				{
					blockNumber: 3,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.3",
								Value: []byte("value.3"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
				{
					blockNumber: 2,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.2",
								Value: []byte("value.2"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
				{
					blockNumber: 1,
					operations: &pbkv.KVOperations{
						Operations: []*pbkv.KVOperation{
							{
								Key:   "key.1",
								Value: []byte("value.1"),
								Type:  pbkv.KVOperation_SET,
							},
						},
					},
					finalBlockHeight: 1,
				},
			},
			lastValidBlock: 2,
			expectedKV:     map[string]string{"kkey.1": "value.1", "kkey.2": "value.2"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			//delete dbPath if it exists
			err = os.RemoveAll(dbPath)
			require.NoError(t, err)

			for _, block := range c.blocks {
				err = db.HandleOperations(ctx, block.blockNumber, block.finalBlockHeight, bstream.StepNew, block.operations)
				require.NoError(t, err)
				_, err = db.Flush(ctx, nil)
				require.NoError(t, err)
			}

			err = db.HandleBlockUndo(ctx, c.lastValidBlock)
			require.NoError(t, err)

			_, err = db.Flush(ctx, nil)
			require.NoError(t, err)

			scanOutput := db.store.Scan(ctx, userKey(""), []byte{'l'}, 0)
			require.NoError(t, scanOutput.Err())

			currentState := map[string]string{}
			for scanOutput.Next() {
				currentState[string(scanOutput.Item().Key)] = string(scanOutput.Item().Value)
			}

			require.Equal(t, c.expectedKV, currentState)
		})
	}
}

func TestDB_UndoOperation(t *testing.T) {
	type foundValue struct {
		previousKeyExists bool
		previousValue     []byte
	}
	cases := []struct {
		name                  string
		operation             *pbkv.KVOperation
		foundValue            foundValue
		expectedUndoOperation *pbkv.KVOperation
	}{
		{
			name:       "sunny path",
			foundValue: foundValue{false, nil},
			operation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.3"),
				Type:  pbkv.KVOperation_SET,
			},
			expectedUndoOperation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.3"),
				Type:  pbkv.KVOperation_DELETE,
			},
		},

		{
			name:       "set operation for previously key set",
			foundValue: foundValue{true, []byte("value.4")},
			operation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.3"),
				Type:  pbkv.KVOperation_SET,
			},
			expectedUndoOperation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.4"),
				Type:  pbkv.KVOperation_SET,
			},
		},
		{
			name:       "delete operation",
			foundValue: foundValue{false, nil},
			operation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.3"),
				Type:  pbkv.KVOperation_DELETE,
			},
			expectedUndoOperation: nil,
		},
		{
			name:       "delete operation for previously set key",
			foundValue: foundValue{true, []byte("value.3")},
			operation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.3"),
				Type:  pbkv.KVOperation_DELETE,
			},
			expectedUndoOperation: &pbkv.KVOperation{
				Key:   "key.3",
				Value: []byte("value.3"),
				Type:  pbkv.KVOperation_SET,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			undoOperationResult := undoOperation(c.operation, c.foundValue.previousValue, c.foundValue.previousKeyExists)
			require.Equal(t, c.expectedUndoOperation, undoOperationResult)
		})
	}

}
