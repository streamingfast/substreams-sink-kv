package wasm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbreader "github.com/streamingfast/substreams-sink-kv/server/wasm/testdata/wasmquery/pb"
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

func Test_IntrinsicGet(t *testing.T) {
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	go launchWasmService(t, endpoint, "./testdata/wasmquery/reader.proto", "./testdata/wasmquery/wasm_query.wasm", db)
	tests := []struct {
		name       string
		req        *pbreader.GetRequest
		db         map[string][]byte
		expectResp *pbreader.Tuple
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbreader.GetRequest{Key: "key1"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectResp: &pbreader.Tuple{Key: "key1", Value: "julien"},
		},
		{
			name: "key not found",
			req:  &pbreader.GetRequest{Key: "key2"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectResp: &pbreader.Tuple{Key: "key2", Value: "not found"},
		},
	}

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbreader.NewEthClient(conn)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			stream, err := cli.Get(context.Background(), test.req)
			require.NoError(t, err)

			resp, err := stream.Recv()
			if test.expectErr {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}

		})
	}
}

func Test_IntrinsicPrefix(t *testing.T) {
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	go launchWasmService(t, endpoint, "./testdata/wasmquery/reader.proto", "./testdata/wasmquery/wasm_query.wasm", db)

	tests := []struct {
		name       string
		req        *pbreader.PrefixRequest
		db         map[string][]byte
		expectResp *pbreader.Tuples
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbreader.PrefixRequest{Prefix: "aa", Limit: 10},
			db: map[string][]byte{
				"aa":  []byte("john"),
				"bb":  []byte("doe"),
				"aa1": []byte("coolio"),
				"ac":  []byte("paul"),
			},
			expectResp: &pbreader.Tuples{Pairs: []*pbreader.Tuple{
				{Key: "aa", Value: "john"},
				{Key: "aa1", Value: "coolio"},
			}},
		},
		{
			name: "nothing found",
			req:  &pbreader.PrefixRequest{Prefix: "zz"},
			db: map[string][]byte{
				"aa":  []byte("john"),
				"bb":  []byte("doe"),
				"aa1": []byte("coolio"),
				"ac":  []byte("paul"),
			},
			expectResp: &pbreader.Tuples{Pairs: []*pbreader.Tuple{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			conn, err := dgrpc.NewInternalClient(endpoint)
			require.NoError(t, err)
			cli := pbreader.NewEthClient(conn)

			stream, err := cli.Prefix(context.Background(), test.req)
			require.NoError(t, err)

			resp, err := stream.Recv()
			if test.expectErr {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}
		})
	}

}

func Test_IntrinsicScan(t *testing.T) {
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	go launchWasmService(t, endpoint, "./testdata/wasmquery/reader.proto", "./testdata/wasmquery/wasm_query.wasm", db)

	tests := []struct {
		name       string
		req        *pbreader.ScanRequest
		db         map[string][]byte
		expectResp *pbreader.Tuples
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbreader.ScanRequest{Start: "a1", ExclusiveEnd: "a4", Limit: 10},
			db: map[string][]byte{
				"a1": []byte("blue"),
				"a2": []byte("yellow"),
				"a3": []byte("amber"),
				"a4": []byte("green"),
				"aa": []byte("red"),
				"bb": []byte("black"),
			},
			expectResp: &pbreader.Tuples{Pairs: []*pbreader.Tuple{
				{Key: "a1", Value: "blue"},
				{Key: "a2", Value: "yellow"},
				{Key: "a3", Value: "amber"},
			}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			conn, err := dgrpc.NewInternalClient(endpoint)
			require.NoError(t, err)
			cli := pbreader.NewEthClient(conn)

			stream, err := cli.Scan(context.Background(), test.req)
			require.NoError(t, err)

			resp, err := stream.Recv()
			if test.expectErr {
				require.NoError(t, err)
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}
		})
	}

}

func launchWasmService(t *testing.T, endpoint, protoPath, wasmPath string, mockDB *db.MockDB) {
	code, err := os.ReadFile(wasmPath)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	protoFileDesc := protoFileToDescriptor(t, protoPath)

	wasmEngine, err := NewEngineFromBytes(code, mockDB, zlog)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	server, err := NewServer(NewConfig(protoFileDesc), wasmEngine, TestPassthroughCodec{}, zlog)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	server.Serve(endpoint)
}
