package wasm

import (
	"encoding/binary"
	"errors"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/streamingfast/substreams-sink-kv/db"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (i *KVExtension) getByPrefix(_ interface{}, callFrame *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callFrame.GetMemoryByIndex(0)
	request := i.vm.CurrentRequest()
	logger := request.Logger()
	ctx := request.Context()

	prefixPtr := params[0].(int32)
	prefixSize := params[1].(int32)
	data, _ := mem.GetData(uint(prefixPtr), uint(prefixSize))
	prefix := make([]byte, prefixSize)
	copy(prefix, data)

	limit := params[2].(int32)

	logger.Debug("kv wasm extension prefix function",
		zap.String("prefix", string(prefix)),
		zap.Int("limit", int(limit)),
	)
	keyVals, _, err := i.kv.GetByPrefix(ctx, string(prefix), int(limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			i.logger.Debug("no values found", zap.Error(err))
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}

		i.logger.Warn("prefix search failed", zap.String("prefix", string(prefix)), zap.Error(err))
		return nil, wasmedge.Result_Fail
	}
	i.logger.Debug("kv database prefix",
		zap.String("prefix", string(prefix)),
		zap.Int32("limit", limit),
		zap.Int("key_value_count", len(keyVals)),
	)

	out := &kvv1.KVPairs{}
	for _, kv := range keyVals {
		out.Pairs = append(out.Pairs, &kvv1.KVPair{Key: kv.Key, Value: kv.Value})
	}
	outBytes, err := proto.Marshal(out)
	if err != nil {
		i.logger.Warn("failed to proto marshal kv pairs", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	protoPtr := i.vm.Allocate(int32(len(outBytes)))
	data, _ = mem.GetData(uint(protoPtr), uint(len(outBytes)))
	copy(data, outBytes)

	outputPtr := params[3].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(protoPtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(outBytes)))

	return []interface{}{1}, wasmedge.Result_Success
}
