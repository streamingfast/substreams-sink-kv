package db

import (
	"testing"

	"github.com/streamingfast/logging"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
)

func TestDB_HandleBlockUndo(t *testing.T) {
	_, tracer := logging.PackageLogger("db", "github.com/streamingfast/substreams-sink-kv/db.test")

	//todo: handle dsn with local db

	db, err := New("", 0, zap.NewNop(), tracer)
	if err != nil {
		t.Fatal(err)
	}
	opsFirstBlock := []*pbkv.KVOperation{}

	opsFirstBlock = append(opsFirstBlock, &pbkv.KVOperation{
		Key:     "k",
		Value:   []byte("1"),
		Ordinal: 0,
		Type:    1,
	})

	opsSecondBlock := []*pbkv.KVOperation{}
	opsThirdBlock := []*pbkv.KVOperation{}

	err = db.storeUndoOperations(nil, 1, opsFirstBlock)
	if err != nil {
		t.Fatal(err)
	}
	err = db.storeUndoOperations(nil, 2, opsSecondBlock)
	if err != nil {
		t.Fatal(err)
	}
	err = db.storeUndoOperations(nil, 3, opsThirdBlock)
	if err != nil {
		t.Fatal(err)
	}

	//todo: populate my db with some undo data

	err = db.HandleBlockUndo(nil, 0)
	if err != nil {
		t.Fatal(err)
	}

}
