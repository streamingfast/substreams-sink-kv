package wasm

import (
	"context"
	"os"
	"testing"

	"github.com/streamingfast/dgrpc"
	pbreader "github.com/streamingfast/substreams-sink-kv/server/wasm/testdata/wasmquery/pb"
	"github.com/stretchr/testify/assert"

	"github.com/streamingfast/substreams-sink-kv/db"
	"github.com/test-go/testify/require"
)

func Test_LaunchServer(t *testing.T) {
	t.Skip("use to test server")
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	db.KV["key1"] = []byte("value1")
	launchWasmService(t, endpoint, "./testdata/wasmquery/reader.proto", "./testdata/wasmquery/wasm_query.wasm", db)

	// runt test
	// then, in a tab, run:
	//   grpcurl -plaintext -proto ./reader.proto -d '{"key": "key1"}' localhost:7878 sf.reader.v1.Eth.Get
	// yields:
	//    {"output": "value1"}
	// then:
	//   grpcurl -plaintext -proto ./reader.proto -d '{"key": "key2"}' localhost:7878 sf.reader.v1.Eth.Get
	// yields:
	//    {"output": "not found"}

}

func Test_Integration(t *testing.T) {
	//t.Skip("fix broken test")
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	go launchWasmService(t, endpoint, "./testdata/wasmquery/reader.proto", "./testdata/wasmquery/wasm_query.wasm", db)

	tests := []struct {
		name       string
		req        *pbreader.Request
		db         map[string][]byte
		expectResp *pbreader.Response
		expectErr  bool
	}{
		{
			req: &pbreader.Request{Key: "key1"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectResp: &pbreader.Response{Value: "julien"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			conn, err := dgrpc.NewInternalClient(endpoint)
			require.NoError(t, err)
			cli := pbreader.NewEthClient(conn)

			stream, err := cli.Get(context.Background(), test.req)
			require.NoError(t, err)

			resp, err := stream.Recv()
			if test.expectErr {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectResp.Value, resp.Value)
			}

		})
	}

}

func launchWasmService(t *testing.T, endpoint, protoPath, wasmPath string, mockDB *db.MockDB) {
	code, err := os.ReadFile(wasmPath)
	require.NoError(t, err)

	protoFileDesc := protoFileToDescriptor(t, protoPath)

	wasmEngine, err := NewEngineFromBytes(code, mockDB, zlog)
	require.NoError(t, err)

	server, err := NewServer(NewConfig(protoFileDesc), wasmEngine, zlog)
	require.NoError(t, err)

	server.Serve(endpoint)
}
