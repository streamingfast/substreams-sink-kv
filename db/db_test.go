package db

import (
	"context"
	"fmt"
	"github.com/test-go/testify/require"
	"os"
	"testing"

	_ "github.com/streamingfast/kvdb/store/badger3"
	"github.com/streamingfast/logging"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
)

func TestDB_HandleOperations(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/substreams-sink-kv-db"

	//delete dbPath if it exists
	err := os.RemoveAll(dbPath)
	require.NoError(t, err)

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
			for _, block := range c.blocks {
				err = db.HandleOperations(ctx, block.blockNumber, block.finalBlockHeight, block.operations)
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

func TestDB_StoreUndoOperations(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/substreams-sink-kv-db"

	//delete dbPath if it exists
	err := os.RemoveAll(dbPath)
	require.NoError(t, err)

	_, tracer := logging.PackageLogger("db", "github.com/streamingfast/substreams-sink-kv/db.test")

	db, err := New(fmt.Sprintf("badger3://%s", dbPath), 0, zap.NewNop(), tracer)
	require.NoError(t, err)

	type blockOperation struct {
		blockNumber uint64
		operations  *pbkv.KVOperations
	}

	cases := []struct {
		name                     string
		block                    blockOperation
		operationsPreviouslyDone *pbkv.KVOperations
		expectedUndoValue        []byte
	}{
		{
			name:                     "sunny path",
			operationsPreviouslyDone: &pbkv.KVOperations{},
			block: blockOperation{
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
			},
			expectedUndoValue: []byte{10, 18, 10, 5, 107, 101, 121, 46, 51, 18, 7, 118, 97, 108, 117, 101, 46, 51, 32, 2}, //Marshal for KVOperations [{key:key.3, value: []bytes(value.3), type:KVOperation_DELETE]
		},

		{
			name: "set operation for previously key set",
			operationsPreviouslyDone: &pbkv.KVOperations{
				Operations: []*pbkv.KVOperation{
					{
						Key:   "key.3",
						Value: []byte("value.4"),
						Type:  pbkv.KVOperation_SET,
					},
				}},

			block: blockOperation{
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
			},
			expectedUndoValue: []byte{10, 18, 10, 5, 107, 101, 121, 46, 51, 18, 7, 118, 97, 108, 117, 101, 46, 52, 32, 1},
		},
		{
			name:                     "delete operation",
			operationsPreviouslyDone: &pbkv.KVOperations{},
			block: blockOperation{
				blockNumber: 1,
				operations: &pbkv.KVOperations{
					Operations: []*pbkv.KVOperation{
						{
							Key:   "key.2",
							Value: []byte("value.2"),
							Type:  pbkv.KVOperation_DELETE,
						},
					},
				},
			},
			expectedUndoValue: []byte{10, 0},
		},
		{
			name: "delete operation for previously set key",
			operationsPreviouslyDone: &pbkv.KVOperations{
				Operations: []*pbkv.KVOperation{
					{
						Key:   "key.3",
						Value: []byte("value.3"),
						Type:  pbkv.KVOperation_SET,
					},
				},
			},
			block: blockOperation{
				blockNumber: 1,
				operations: &pbkv.KVOperations{
					Operations: []*pbkv.KVOperation{
						{
							Key:   "key.3",
							Value: []byte("value.3"),
							Type:  pbkv.KVOperation_DELETE,
						},
					},
				},
			},
			expectedUndoValue: []byte{10, 18, 10, 5, 107, 101, 121, 46, 51, 18, 7, 118, 97, 108, 117, 101, 46, 51, 32, 1},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			db.AddOperations(c.operationsPreviouslyDone)
			_, err = db.Flush(ctx, nil)
			require.NoError(t, err)

			err = db.storeUndoOperations(ctx, c.block.blockNumber, c.block.operations.Operations)
			require.NoError(t, err)

			var undoValue []byte
			undoValue, err = db.store.Get(ctx, undoKey(c.block.blockNumber))
			require.NoError(t, err)
			require.Equal(t, c.expectedUndoValue, undoValue)
		})
	}

}

func TestDB_HandleBlockUndo(t *testing.T) {
	//ctx := context.Background()

}
