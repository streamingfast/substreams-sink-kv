package wasm

import (
	"encoding/json"
	"testing"

	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var zlog, _ = logging.PackageLogger("sink-kv", "github.com/streamingfast/substreams-sink-kv/server/wasm.test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zapcore.DebugLevel))
}

func assertProtoEqual(t *testing.T, expected proto.Message, actual proto.Message) {
	t.Helper()

	if !proto.Equal(expected, actual) {
		expectedAsJSON, err := protojson.Marshal(expected)
		require.NoError(t, err)

		actualAsJSON, err := protojson.Marshal(actual)
		require.NoError(t, err)

		expectedAsMap := map[string]interface{}{}
		err = json.Unmarshal(expectedAsJSON, &expectedAsMap)
		require.NoError(t, err)

		actualAsMap := map[string]interface{}{}
		err = json.Unmarshal(actualAsJSON, &actualAsMap)
		require.NoError(t, err)

		// We use equal is not equal above so we get a good diff, if the first condition failed, the second will also always
		// fail which is what we want here
		assert.Equal(t, expectedAsMap, actualAsMap)
	}
}

// The TestPassthroughCodec is very similar to the  PassthroughCodec used in production, except in handles the case where
// the object passed around is *NOT* a testPassthroughBytes object. The reason for this be due to the nature of how the
// proto library handles Codec. The proto library keeps a global registry of Codecs. In our test cases we are creating a
// GRPC server and a GRPC client, since both use the same package, they share the *same* global registry of codec (not this is
// not the case in a production environment since the server is run independently for the clients). With that in mind, both the
// client and server need to have a `proto` Codec, hence why we create a Codec that can handle generic proto message (utilized
// by the GRPC client) and the `bytesPassthrough` message utilized by the GRPC server
type TestPassthroughCodec struct{}

var _ Codec = TestPassthroughCodec{}

func (TestPassthroughCodec) Marshal(v interface{}) ([]byte, error) {
	el, ok := v.(*testPassthroughBytes)
	if ok {
		return el.bytes, nil
	}
	return proto.Marshal(v.(proto.Message))
}

func (TestPassthroughCodec) Unmarshal(data []byte, v interface{}) error {
	el, ok := v.(*testPassthroughBytes)
	if ok {
		el.bytes = data
		return nil
	}
	return proto.Unmarshal(data, v.(proto.Message))
}

func (TestPassthroughCodec) Name() string { return "proto" }

func (TestPassthroughCodec) NewMessage() Byteable { return NewTestPassthroughCodec() }

// Passing bytes around
type testPassthroughBytes struct {
	bytes []byte
}

func NewTestPassthroughCodec() *testPassthroughBytes {
	return &testPassthroughBytes{}
}

func (b *testPassthroughBytes) Set(in []byte) {
	b.bytes = in
}

func (b *testPassthroughBytes) Bytes() []byte {
	return b.bytes
}
