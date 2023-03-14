package db

import (
	"context"
	"strings"

	sink "github.com/streamingfast/substreams-sink"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
)

var _ Loader = (*MockDB)(nil)
var _ Reader = (*MockDB)(nil)

type MockDB struct {
	KV map[string][]byte
}

func NewMockDB() *MockDB {
	return &MockDB{
		KV: map[string][]byte{},
	}
}

func (m *MockDB) AddOperations(ops *kvv1.KVOperations) {
	panic("implement me")
}

func (m *MockDB) AddOperation(op *kvv1.KVOperation) {
	panic("implement me")
}

func (m *MockDB) Flush(ctx context.Context, cursor *sink.Cursor) (count int, err error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) GetCursor(ctx context.Context) (*sink.Cursor, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) WriteCursor(ctx context.Context, c *sink.Cursor) error {
	//TODO implement me
	panic("implement me")
}

func (m *MockDB) Get(ctx context.Context, key string) (value []byte, err error) {
	if val, found := m.KV[key]; found {
		return val, nil
	}
	return nil, ErrNotFound
}

func (m *MockDB) GetMany(ctx context.Context, keys []string) (values [][]byte, err error) {
	for _, k := range keys {
		v, err := m.Get(ctx, k)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return
}

func (m *MockDB) GetByPrefix(ctx context.Context, prefix string, limit int) (values []*kvv1.KV, limitReached bool, err error) {
	for k, v := range m.KV {
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		values = append(values, &kvv1.KV{Key: k, Value: v})

		if len(values) == limit {
			limitReached = true
			break
		}
	}
	if len(values) == 0 {
		return nil, false, ErrNotFound
	}
	return values, limitReached, nil
}

func (m *MockDB) Scan(ctx context.Context, start string, exclusiveEnd string, limit int) (values []*kvv1.KV, limitReached bool, err error) {
	//TODO implement me
	panic("implement me")
}
