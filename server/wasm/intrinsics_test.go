package wasm

import (
	"encoding/hex"
	"fmt"
	"testing"

	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"github.com/test-go/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewEngineFromBytes(t *testing.T) {
	obj := &pbkv.KVPairs{
		Pairs: []*pbkv.KVPair{
			{
				Key:   "key1",
				Value: []byte{0xaa, 0xbb},
			},
			{
				Key:   "key2",
				Value: []byte{0xaa, 0xcc},
			},
		},
	}

	cnt, err := proto.Marshal(obj)
	require.NoError(t, err)
	fmt.Println("OK", hex.EncodeToString(cnt))

	obj = &pbkv.KVPairs{}

	cnt, err = proto.Marshal(obj)
	require.NoError(t, err)
	fmt.Println("OK:", hex.EncodeToString(cnt))
}
