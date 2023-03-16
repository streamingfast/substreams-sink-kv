package wasm

import (
	"context"
	"os"
	"testing"

	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbtest "github.com/streamingfast/substreams-sink-kv/server/wasm/testdata/wasmquery/pb"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_IntrinsicGet(t *testing.T) {
	endpoint := "localhost:7878"
	db := db.NewMockDB()
	server := getWasmService(
		t,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)
	go server.Serve(endpoint)
	defer func() {
		server.Shutdown()
	}()

	tests := []struct {
		name       string
		req        *pbtest.GetTestRequest
		db         map[string][]byte
		expectResp *pbtest.Tuple
		expectErr  error
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
			expectErr: status.Error(codes.Internal, "not found"),
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
			if test.expectErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectErr, err)
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
	server := getWasmService(
		t,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)
	go server.Serve(endpoint)
	defer func() {
		server.Shutdown()
	}()

	tests := []struct {
		name       string
		req        *pbtest.TestGetManyRequest
		db         map[string][]byte
		expectResp *pbtest.Tuples
		expectErr  error
	}{
		{
			name: "golden path",
			req:  &pbtest.TestGetManyRequest{Keys: []string{"key1", "key3"}},
			db: map[string][]byte{
				"key1": []byte("red"),
				"key2": []byte("black"),
				"key3": []byte("green"),
			},
			expectResp: &pbtest.Tuples{Pairs: []*pbtest.Tuple{
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
			expectErr: status.Error(codes.Internal, "not found"),
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
			if test.expectErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectErr, err)
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
	server := getWasmService(
		t,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)
	go server.Serve(endpoint)
	defer func() {
		server.Shutdown()
	}()

	tests := []struct {
		name       string
		req        *pbtest.TestPrefixRequest
		db         map[string][]byte
		expectResp *pbtest.Tuples
		expectErr  error
	}{
		{
			name: "with limit",
			req:  &pbtest.TestPrefixRequest{Prefix: "aa", Limit: i32(10)},
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
			name: "no limit",
			req:  &pbtest.TestPrefixRequest{Prefix: "aa"},
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
			name: "limit zero",
			req:  &pbtest.TestPrefixRequest{Prefix: "aa", Limit: i32(0)},
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
			name: "hit limit",
			req:  &pbtest.TestPrefixRequest{Prefix: "aa", Limit: i32(1)},
			db: map[string][]byte{
				"aa":  []byte("john"),
				"bb":  []byte("doe"),
				"aa1": []byte("coolio"),
				"ac":  []byte("paul"),
			},
			expectResp: &pbtest.Tuples{Pairs: []*pbtest.Tuple{
				{Key: "aa", Value: "john"},
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
			if test.expectErr != nil {
				require.Error(t, err)
				assert.Equal(t, test.expectErr, err)
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
	server := getWasmService(
		t,
		"./testdata/wasmquery/test.proto",
		"./testdata/wasmquery/wasm_query.wasm",
		"sf.test.v1.TestService",
		db,
	)
	go server.Serve(endpoint)
	defer func() {
		server.Shutdown()
	}()

	tests := []struct {
		name       string
		req        *pbtest.TestScanRequest
		db         map[string][]byte
		expectResp *pbtest.Tuples
		expectErr  bool
	}{
		{
			name: "with limit",
			req:  &pbtest.TestScanRequest{Start: "a1", ExclusiveEnd: "a4", Limit: i32(10)},
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
			name: "no limit",
			req:  &pbtest.TestScanRequest{Start: "a1", ExclusiveEnd: "a4"},
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
			name: "zero limit",
			req:  &pbtest.TestScanRequest{Start: "a1", ExclusiveEnd: "a4", Limit: i32(0)},
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
			name: "hit limit",
			req:  &pbtest.TestScanRequest{Start: "a1", ExclusiveEnd: "a4", Limit: i32(2)},
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
			}},
		},
		{
			name: "nothing found",
			req:  &pbtest.TestScanRequest{Start: "c1", ExclusiveEnd: "c8", Limit: i32(10)},
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

func getWasmService(t *testing.T, protoPath, wasmPath, fqServiceName string, mockDB *db.MockDB) *Server {
	code, err := os.ReadFile(wasmPath)
	require.NoError(t, err)

	protoFileDesc := protoFileToDescriptor(t, protoPath)

	wasmEngine, err := NewEngineFromBytes(code, mockDB, zlog)
	require.NoError(t, err)

	config, err := NewConfig(protoFileDesc, fqServiceName)
	require.NoError(t, err)

	server, err := NewServer(config, wasmEngine, TestPassthroughCodec{}, zlog)
	require.NoError(t, err)

	return server

}

func i32(v int32) *int32 {
	return &v
}
