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
