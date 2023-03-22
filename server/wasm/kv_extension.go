package wasm

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
	"github.com/streamingfast/substreams-sink-kv/wasmquery"

	"github.com/streamingfast/substreams-sink-kv/db"
	"go.uber.org/zap"
)

var _ wasmquery.WASMExtension = (*KVExtension)(nil)

type KVExtension struct {
	kv             db.Reader
	vm             wasmquery.VM
	currentRequest *wasmquery.Request
	logger         *zap.Logger
}

func NewKVExtension(kv db.Reader, vm wasmquery.VM, logger *zap.Logger) *KVExtension {
	return &KVExtension{
		kv:     kv,
		vm:     vm,
		logger: logger,
	}
}

func (i *KVExtension) SetRequest(r *wasmquery.Request) {
	i.currentRequest = r
}

func (i *KVExtension) SetupWASMHostModule(module *wasmedge.Module) {
	hostGetKeyType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostGetKeyFunc := wasmedge.NewFunction(hostGetKeyType, i.getKey, nil, 0)
	hostGetKeyType.Release()
	module.AddFunction("get_key", hostGetKeyFunc)

	hostGetManyKeysType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostGetMenyKeysFunc := wasmedge.NewFunction(hostGetManyKeysType, i.getManyKeys, nil, 0)
	hostGetManyKeysType.Release()
	module.AddFunction("get_many_keys", hostGetMenyKeysFunc)

	hostGetByPrefixType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostGetByPrefixFunc := wasmedge.NewFunction(hostGetByPrefixType, i.getByPrefix, nil, 0)
	hostGetByPrefixType.Release()
	module.AddFunction("get_by_prefix", hostGetByPrefixFunc)

	hostScanType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostScanFunc := wasmedge.NewFunction(hostScanType, i.scan, nil, 0)
	hostScanType.Release()
	module.AddFunction("scan", hostScanFunc)
}
