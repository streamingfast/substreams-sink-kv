package wasmquery

import (
	"testing"

	"github.com/streamingfast/substreams-sink-kv/db"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
)

func TestEngine_newEngineFromFile(t *testing.T) {
	wasmFile := "./testdata/wasm_query.wasm"
	kv := db.NewMockDB()
	eng, err := NewEngineFromFile(wasmFile, kv, zlog)
	require.NoError(t, err)
	assert.NotNil(t, eng)
}
