package wasm

import (
	"context"
	"os"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/streamingfast/substreams-sink-kv/wasmquery"

	"github.com/streamingfast/dgrpc"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbtest "github.com/streamingfast/substreams-sink-kv/server/wasm/testdata/wasmquery/pb"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_WASMExtentionGet(t *testing.T) {
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
		req        *pbtest.TestGetRequest
		db         map[string][]byte
		expectResp *pbtest.Tuple
		expectErr  bool
	}{
		{
			name: "golden path",
			req:  &pbtest.TestGetRequest{Key: "key1"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectResp: &pbtest.Tuple{Key: "key1", Value: "julien"},
		},
		{
			name: "key not found",
			req:  &pbtest.TestGetRequest{Key: "key2"},
			db: map[string][]byte{
				"key1": []byte("julien"),
			},
			expectErr: true,
		},
	}

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbtest.NewTestServiceClient(conn)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			resp, err := cli.TestGet(context.Background(), test.req)
			if test.expectErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok, "error must be a GRPC error")
				assert.Equal(t, codes.Unknown, e.Code())
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}

		})
	}
}

func Test_WASMExtentionGetMany(t *testing.T) {
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
			expectErr: true,
		},
	}

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbtest.NewTestServiceClient(conn)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db.KV = test.db

			resp, err := cli.TestGetMany(context.Background(), test.req)
			if test.expectErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok, "error must be a GRPC error")
				assert.Equal(t, codes.Unknown, e.Code())
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}

		})
	}
}

func Test_WASMExtentionPrefix(t *testing.T) {
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
		expectErr  bool
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

			resp, err := cli.TestPrefix(context.Background(), test.req)
			if test.expectErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok, "error must be a GRPC error")
				assert.Equal(t, codes.Unknown, e.Code())
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}
		})
	}

}

func Test_WASMExtentionScan(t *testing.T) {
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

			resp, err := cli.TestScan(context.Background(), test.req)
			if test.expectErr {
				require.Error(t, err)
				e, ok := status.FromError(err)
				require.True(t, ok, "error must be a GRPC error")
				assert.Equal(t, codes.Unknown, e.Code())
			} else {
				require.NoError(t, err)
				assertProtoEqual(t, test.expectResp, resp)
			}
		})
	}

}

func Test_WASMExtensionSleep(t *testing.T) {
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

	time.Sleep(1 * time.Second)

	reqDurationA := 10000
	reqDurationB := 10
	reqAEnd := make(chan bool, 1)
	reqBEnd := make(chan bool, 1)

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbtest.NewTestServiceClient(conn)

	go func() {
		resp, err := cli.TestSleep(context.Background(), &pbtest.TestSleepRequest{Duration: int32(reqDurationA), RequestId: "A"})
		require.NoError(t, err)
		reqAEnd <- true
		assertProtoEqual(t, &pbtest.Response{Output: "A completed"}, resp)
	}()

	time.Sleep(3 * time.Millisecond)

	go func() {
		resp, err := cli.TestSleep(context.Background(), &pbtest.TestSleepRequest{Duration: int32(reqDurationB), RequestId: "B"})
		require.NoError(t, err)
		reqBEnd <- true
		assertProtoEqual(t, &pbtest.Response{Output: "B completed"}, resp)
	}()

	select {
	case _ = <-reqAEnd:
	case _ = <-reqBEnd:
		t.Fatal("expect request A to terminate first")
		return
	case <-time.After(10100 * time.Millisecond):
		t.Fatal("expect request A to terminate withing the specific sleep duration")
		return
	}

	select {
	case _ = <-reqBEnd:
	case <-time.After(20 * time.Millisecond):
		t.Fatal("expect request b to terminate withing the specific sleep duration")
		return
	}

}

func Test_WASMExtensionPanic(t *testing.T) {
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

	time.Sleep(1 * time.Second)

	conn, err := dgrpc.NewInternalClient(endpoint)
	require.NoError(t, err)
	cli := pbtest.NewTestServiceClient(conn)

	request := &pbtest.TestPanicRequest{ShouldPanic: true}
	_, err = cli.TestPanic(context.Background(), request)

	require.Error(t, err)
	assert.Error(t, err, status.Error(codes.Internal, "panic in wasm: \"panicking\" at src/lib.r:118:1"))

}

func getWasmService(t *testing.T, protoPath, wasmPath, fqServiceName string, mockDB *db.MockDB) *wasmquery.Engine {
	code, err := os.ReadFile(wasmPath)
	require.NoError(t, err)

	protoFileDesc := protoFileToDescriptor(t, protoPath)

	config, err := wasmquery.NewServiceConfig(protoFileDesc, fqServiceName)
	require.NoError(t, err)

	engineConfig := wasmquery.NewEngineConfig(1, code, config)

	engine, err := wasmquery.NewEngine(engineConfig, func(vm wasmquery.VM, logger *zap.Logger) wasmquery.WASMExtension {
		return NewKVExtension(mockDB, vm, logger)
	}, zlog, wasmquery.SkipProtoRegister())
	require.NoError(t, err)

	return engine

}

func i32(v int32) *int32 {
	return &v
}
