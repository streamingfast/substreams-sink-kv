package db

import (
	"context"
	"testing"

	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
)

func TestMockDB_Scan(t *testing.T) {
	tests := []struct {
		keys           map[string][]byte
		start          string
		exclusivelyEnd string
		limit          int

		expectValue []*kvv1.KV
		expectLimit bool
		expectErr   bool
	}{
		{
			keys: map[string][]byte{
				"a1": []byte("blue"),
				"a2": []byte("yellow"),
				"a3": []byte("amber"),
				"a4": []byte("green"),
				"aa": []byte("red"),
				"bb": []byte("black"),
			},
			start:          "a1",
			exclusivelyEnd: "a4",
			expectValue: []*kvv1.KV{
				{Key: "a1", Value: []byte("blue")},
				{Key: "a2", Value: []byte("yellow")},
				{Key: "a3", Value: []byte("amber")},
			},
		},
		{
			keys: map[string][]byte{
				"a1": []byte("blue"),
				"a2": []byte("yellow"),
				"a3": []byte("amber"),
				"a4": []byte("green"),
				"aa": []byte("red"),
				"bb": []byte("black"),
			},
			start:          "b1",
			exclusivelyEnd: "b4",
			expectErr:      true,
		},
		{
			keys: map[string][]byte{
				"a1": []byte("blue"),
				"a2": []byte("yellow"),
				"a3": []byte("amber"),
				"a4": []byte("green"),
				"aa": []byte("red"),
				"bb": []byte("black"),
			},
			start:          "a",
			exclusivelyEnd: "z",
			expectValue: []*kvv1.KV{
				{Key: "a1", Value: []byte("blue")},
				{Key: "a2", Value: []byte("yellow")},
				{Key: "a3", Value: []byte("amber")},
				{Key: "a4", Value: []byte("green")},
				{Key: "aa", Value: []byte("red")},
				{Key: "bb", Value: []byte("black")},
			},
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			db := &MockDB{KV: test.keys}
			values, limitReached, err := db.Scan(context.Background(), test.start, test.exclusivelyEnd, test.limit)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectLimit, limitReached)
				require.Equal(t, len(test.expectValue), len(values))
				for i, kv := range values {
					assertProtoEqual(t, test.expectValue[i], kv)
				}

			}
		})
	}
}
