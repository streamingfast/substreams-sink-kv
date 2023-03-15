package wasm

import (
	"context"
	"encoding/binary"
	"errors"
	"google.golang.org/protobuf/proto"

	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/streamingfast/substreams-sink-kv/db"
	"go.uber.org/zap"
)

func (e *Engine) getKey(_ interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callframe.GetMemoryByIndex(0)

	keyPtr := params[0].(int32)
	keySize := params[1].(int32)
	data, _ := mem.GetData(uint(keyPtr), uint(keySize))
	key := make([]byte, keySize)

	copy(key, data)

	// TODO: ctx is probably incorrect
	val, err := e.kv.Get(context.Background(), string(key))
	if err != nil {
		if err == db.ErrNotFound {
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}
		e.logger.Warn("get key failed", zap.String("key", string(key)), zap.Error(err))
		return []interface{}{int32(0)}, wasmedge.Result_Fail
	}

	valuePtr := e.allocate(int32(len(val)))
	data, _ = mem.GetData(uint(valuePtr), uint(len(val)))

	copy(data, val)

	outputPtr := params[2].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(valuePtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(val)))

	return []interface{}{1}, wasmedge.Result_Success
}

func (e *Engine) getManyKeys(_ interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callframe.GetMemoryByIndex(0)

	keyPtr := params[0].(int32)
	keySize := params[1].(int32)
	data, _ := mem.GetData(uint(keyPtr), uint(keySize))
	key := make([]byte, keySize)

	copy(key, data)

	keys := &pbkv.KVKeys{}
	if err := proto.Unmarshal(data, keys); err != nil {
		e.logger.Warn("failed to proto unmarhal proto keys", zap.Error(err))
		return nil, wasmedge.Result_Fail

	}

	// TODO: ctx is probably incorrect
	values, err := e.kv.GetMany(context.Background(), keys.Keys)
	if err != nil {
		if err == db.ErrNotFound {
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}
		e.logger.Warn("get key failed", zap.String("key", string(key)), zap.Error(err))
		return []interface{}{int32(0)}, wasmedge.Result_Fail
	}
	e.logger.Debug("kv database prefix",
		zap.Strings("keys", keys.Keys),
		zap.Int("resp", len(values)),
	)
	out := &pbkv.KVPairs{}
	for idx, value := range values {
		out.Pairs = append(out.Pairs, &pbkv.KVPair{Key: keys.Keys[idx], Value: value})
	}
	outBytes, err := proto.Marshal(out)
	if err != nil {
		e.logger.Warn("failed to proto marshal kv pairs", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	protoPtr := e.allocate(int32(len(outBytes)))
	data, _ = mem.GetData(uint(protoPtr), uint(len(outBytes)))
	copy(data, outBytes)

	outputPtr := params[2].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(protoPtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(outBytes)))

	return []interface{}{1}, wasmedge.Result_Success
}

func (e *Engine) getByPrefix(_ interface{}, callFrame *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callFrame.GetMemoryByIndex(0)

	prefixPtr := params[0].(int32)
	prefixSize := params[1].(int32)
	data, _ := mem.GetData(uint(prefixPtr), uint(prefixSize))
	prefix := make([]byte, prefixSize)
	copy(prefix, data)

	limit := params[2].(int32)

	keyVals, _, err := e.kv.GetByPrefix(context.Background(), string(prefix), int(limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			e.logger.Debug("no values found", zap.Error(err))
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}

		e.logger.Warn("prefix search failed", zap.String("prefix", string(prefix)), zap.Error(err))
		return nil, wasmedge.Result_Fail
	}
	e.logger.Debug("kv database prefix",
		zap.String("prefix", string(prefix)),
		zap.Int32("limit", limit),
		zap.Int("key_value_count", len(keyVals)),
	)

	out := &pbkv.KVPairs{}
	for _, kv := range keyVals {
		out.Pairs = append(out.Pairs, &pbkv.KVPair{Key: kv.Key, Value: kv.Value})
	}
	outBytes, err := proto.Marshal(out)
	if err != nil {
		e.logger.Warn("failed to proto marshal kv pairs", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	protoPtr := e.allocate(int32(len(outBytes)))
	data, _ = mem.GetData(uint(protoPtr), uint(len(outBytes)))
	copy(data, outBytes)

	outputPtr := params[3].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(protoPtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(outBytes)))

	return []interface{}{1}, wasmedge.Result_Success
}

func (e *Engine) scan(_ interface{}, callFrame *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
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

	keyVals, _, err := e.kv.Scan(context.Background(), string(start), string(exclusiveEnd), int(limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			e.logger.Debug("no values found", zap.Error(err))
			return []interface{}{int32(0)}, wasmedge.Result_Success
		}

		e.logger.Warn("scan search failed",
			zap.String("start", string(start)),
			zap.String("exclusive_end", string(exclusiveEnd)),
			zap.Error(err),
		)
		return nil, wasmedge.Result_Fail
	}
	e.logger.Debug("kv database scan",
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
		e.logger.Warn("failed to proto marshal kv pairs", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	protoPtr := e.allocate(int32(len(outBytes)))
	data, _ = mem.GetData(uint(protoPtr), uint(len(outBytes)))
	copy(data, outBytes)

	outputPtr := params[5].(int32)
	data, _ = mem.GetData(uint(outputPtr), uint(8))
	binary.LittleEndian.PutUint32(data[0:4], uint32(protoPtr))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(outBytes)))

	return []interface{}{1}, wasmedge.Result_Success
}
