package wasm

import (
	"encoding/binary"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/streamingfast/substreams-sink-kv/db"
	"go.uber.org/zap"
)

func (i *KVExtension) getKey(_ interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	request := i.vm.CurrentRequest()
	logger := request.Logger()
	ctx := request.Context()

	mem := callframe.GetMemoryByIndex(0)

	keyPtr := params[0].(int32)
	keySize := params[1].(int32)
	data, _ := mem.GetData(uint(keyPtr), uint(keySize))
	key := make([]byte, keySize)

	copy(key, data)

	logger.Debug("kv wasm extension get key function", zap.String("key", string(key)))
	// TODO: ctx is probably incorrect
	val, err := i.kv.Get(ctx, string(key))
	if err != nil {
		if err == db.ErrNotFound {
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}
		i.logger.Warn("get key failed", zap.String("key", string(key)), zap.Error(err))
		return []interface{}{int32(0)}, wasmedge.Result_Fail
	}

	valuePtr := i.vm.Allocate(int32(len(val)))
	data, _ = mem.GetData(uint(valuePtr), uint(len(val)))

	copy(data, val)

	outputPtr := params[2].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(valuePtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(val)))

	return []interface{}{1}, wasmedge.Result_Success
}
