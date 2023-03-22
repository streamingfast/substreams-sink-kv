package wasm

import (
	"encoding/binary"
	"errors"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (i *KVExtension) scan(_ interface{}, callFrame *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	request := i.vm.CurrentRequest()
	ctx := request.Context()
	logger := request.Logger()
	mem := callFrame.GetMemoryByIndex(0)

	startPtr := params[0].(int32)
	startSize := params[1].(int32)
	data, _ := mem.GetData(uint(startPtr), uint(startSize))
	start := make([]byte, startSize)
	copy(start, data)

	exclusiveEndPtr := params[2].(int32)
	exclusiveEndSize := params[3].(int32)
	data, _ = mem.GetData(uint(exclusiveEndPtr), uint(exclusiveEndSize))
	exclusiveEnd := make([]byte, exclusiveEndSize)
	copy(exclusiveEnd, data)

	limit := params[4].(int32)

	logger.Debug("kv WASM extension scan function",
		zap.String("start", string(start)),
		zap.String("exclusive_end", string(exclusiveEnd)),
		zap.Int("limit", int(limit)),
	)
	keyVals, _, err := i.kv.Scan(ctx, string(start), string(exclusiveEnd), int(limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			i.logger.Debug("no values found", zap.Error(err))
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}

		i.logger.Warn("scan search failed",
			zap.String("start", string(start)),
			zap.String("exclusive_end", string(exclusiveEnd)),
			zap.Error(err),
		)
		return nil, wasmedge.Result_Fail
	}
	i.logger.Debug("kv database scan",
		zap.String("start", string(start)),
		zap.String("exclusive_end", string(exclusiveEnd)),
		zap.Int32("limit", limit),
		zap.Int("key_value_count", len(keyVals)),
	)

	out := &pbkv.KVPairs{}
	for _, kv := range keyVals {
		out.Pairs = append(out.Pairs, &pbkv.KVPair{Key: kv.Key, Value: kv.Value})
	}
	outBytes, err := proto.Marshal(out)
	if err != nil {
		i.logger.Warn("failed to proto marshal kv pairs", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	protoPtr := i.vm.Allocate(int32(len(outBytes)))
	data, _ = mem.GetData(uint(protoPtr), uint(len(outBytes)))
	copy(data, outBytes)

	outputPtr := params[5].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(protoPtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(outBytes)))

	return []interface{}{1}, wasmedge.Result_Success
}
