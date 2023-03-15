package wasm

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbtest "github.com/streamingfast/substreams-sink-kv/server/wasm/testdata/wasmquery/pb"
	"github.com/test-go/testify/require"
)

func Test_LaunchServer(t *testing.T) {
	t.Skip("use to test server")
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	db.KV["key1"] = []byte("value1")
	launchWasmService(
		t,
		endpoint,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)

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
	go launchWasmService(
		t,
		endpoint,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)

	tests := []struct {
		name       string
		req        *pbtest.GetTestRequest
		db         map[string][]byte
		expectResp *pbtest.Tuple
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbtest.GetTestRequest{Key: "key1"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectResp: &pbtest.Tuple{Key: "key1", Value: "julien"},
		},
		{
			name: "key not found",
			req:  &pbtest.GetTestRequest{Key: "key2"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectResp: &pbtest.Tuple{Key: "key2", Value: "not found"},
		},
	}

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbtest.NewTestServiceClient(conn)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			stream, err := cli.TestGet(context.Background(), test.req)
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

func Test_IntrinsicGetMany(t *testing.T) {
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	go launchWasmService(
		t,
		endpoint,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",

		db,
	)

	tests := []struct {
		name       string
		req        *pbtest.TestGetManyRequest
		db         map[string][]byte
		expectResp *pbtest.OptionalTuples
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbtest.TestGetManyRequest{Keys: []string{"key1", "key3"}},
			db: map[string][]byte{
				"key1": []byte("red"),
				"key2": []byte("black"),
				"key3": []byte("green"),
			},
			expectResp: &pbtest.OptionalTuples{Pairs: []*pbtest.Tuple{
				{Key: "key1", Value: "red"},
				{Key: "key3", Value: "green"},
			}},
		},
		{
			name: "Not Found",
			req:  &pbtest.TestGetManyRequest{Keys: []string{"key1", "key4"}},
			db: map[string][]byte{
				"key1": []byte("red"),
				"key2": []byte("black"),
				"key3": []byte("green"),
			},
			expectResp: &pbtest.OptionalTuples{Pairs: []*pbtest.Tuple{}, Error: "Not Found"},
		},
	}

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbtest.NewTestServiceClient(conn)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			stream, err := cli.TestGetMany(context.Background(), test.req)
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
	go launchWasmService(
		t,
		endpoint,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",

		db,
	)

	tests := []struct {
		name       string
		req        *pbtest.TestPrefixRequest
		db         map[string][]byte
		expectResp *pbtest.Tuples
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbtest.TestPrefixRequest{Prefix: "aa", Limit: 10},
			db: map[string][]byte{
				"aa":  []byte("john"),
				"bb":  []byte("doe"),
				"aa1": []byte("coolio"),
				"ac":  []byte("paul"),
			},
			expectResp: &pbtest.Tuples{Pairs: []*pbtest.Tuple{
				{Key: "aa", Value: "john"},
				{Key: "aa1", Value: "coolio"},
			}},
		},
		{
			name: "nothing found",
			req:  &pbtest.TestPrefixRequest{Prefix: "zz"},
			db: map[string][]byte{
				"aa":  []byte("john"),
				"bb":  []byte("doe"),
				"aa1": []byte("coolio"),
				"ac":  []byte("paul"),
			},
			expectResp: &pbtest.Tuples{Pairs: []*pbtest.Tuple{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			conn, err := dgrpc.NewInternalClient(endpoint)
			require.NoError(t, err)
			cli := pbtest.NewTestServiceClient(conn)

			stream, err := cli.TestPrefix(context.Background(), test.req)
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
	go launchWasmService(
		t,
		endpoint,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)

	tests := []struct {
		name       string
		req        *pbtest.TestScanRequest
		db         map[string][]byte
		expectResp *pbtest.Tuples
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbtest.TestScanRequest{Start: "a1", ExclusiveEnd: "a4", Limit: 10},
			db: map[string][]byte{
				"a1": []byte("blue"),
				"a2": []byte("yellow"),
				"a3": []byte("amber"),
				"a4": []byte("green"),
				"aa": []byte("red"),
				"bb": []byte("black"),
			},
			expectResp: &pbtest.Tuples{Pairs: []*pbtest.Tuple{
				{Key: "a1", Value: "blue"},
				{Key: "a2", Value: "yellow"},
				{Key: "a3", Value: "amber"},
			}},
		},
		{
			name: "nothing found",
			req:  &pbtest.TestScanRequest{Start: "c1", ExclusiveEnd: "c8", Limit: 10},
			db: map[string][]byte{
				"a1": []byte("blue"),
				"a2": []byte("yellow"),
				"a3": []byte("amber"),
				"a4": []byte("green"),
				"aa": []byte("red"),
				"bb": []byte("black"),
			},
			expectResp: &pbtest.Tuples{Pairs: []*pbtest.Tuple{}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			conn, err := dgrpc.NewInternalClient(endpoint)
			require.NoError(t, err)
			cli := pbtest.NewTestServiceClient(conn)

			stream, err := cli.TestScan(context.Background(), test.req)
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

func launchWasmService(t *testing.T, endpoint, protoPath, wasmPath, fqServiceName string, mockDB *db.MockDB) {
	code, err := os.ReadFile(wasmPath)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	protoFileDesc := protoFileToDescriptor(t, protoPath)

	wasmEngine, err := NewEngineFromBytes(code, mockDB, zlog)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	config, err := NewConfig(protoFileDesc, fqServiceName)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}
	server, err := NewServer(config, wasmEngine, TestPassthroughCodec{}, zlog)
	if err != nil {
		panic(fmt.Errorf("%w", err))
	}

	server.Serve(endpoint)
}
