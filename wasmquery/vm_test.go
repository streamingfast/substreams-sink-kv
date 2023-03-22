package wasmquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
)

func TestEngine_newEngineFromFile(t *testing.T) {
	wasmFile := "./testdata/wasm_query.wasm"
	eng, err := newVMFromFile(0, wasmFile, zlog)
	require.NoError(t, err)
	assert.NotNil(t, eng)
}
