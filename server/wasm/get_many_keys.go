package wasm

import (
	"encoding/binary"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (i *KVExtension) getManyKeys(_ interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callframe.GetMemoryByIndex(0)
	request := i.vm.CurrentRequest()
	logger := request.Logger()
	ctx := request.Context()

	keyPtr := params[0].(int32)
	keySize := params[1].(int32)
	data, _ := mem.GetData(uint(keyPtr), uint(keySize))
	key := make([]byte, keySize)

	copy(key, data)

	keys := &pbkv.KVKeys{}
	if err := proto.Unmarshal(data, keys); err != nil {
		i.logger.Warn("failed to proto unmarshal proto keys", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	logger.Debug("kv wasm extension get many keys function", zap.Reflect("keys", keys))
	// TODO: ctx is probably incorrect
	values, err := i.kv.GetMany(ctx, keys.Keys)
	if err != nil {
		if err == db.ErrNotFound {
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}
		i.logger.Warn("get key failed", zap.String("key", string(key)), zap.Error(err))
		return []interface{}{int32(0)}, wasmedge.Result_Fail
	}
	i.logger.Debug("kv database prefix",
		zap.Strings("keys", keys.Keys),
		zap.Int("resp", len(values)),
	)
	out := &pbkv.KVPairs{}
	for idx, value := range values {
		out.Pairs = append(out.Pairs, &pbkv.KVPair{Key: keys.Keys[idx], Value: value})
	}
	outBytes, err := proto.Marshal(out)
	if err != nil {
		i.logger.Warn("failed to proto marshal kv pairs", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	protoPtr := i.vm.Allocate(int32(len(outBytes)))
	data, _ = mem.GetData(uint(protoPtr), uint(len(outBytes)))
	copy(data, outBytes)

	outputPtr := params[2].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(protoPtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(outBytes)))

	return []interface{}{1}, wasmedge.Result_Success
}
