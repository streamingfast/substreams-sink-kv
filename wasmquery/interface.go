package wasmquery

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
	"go.uber.org/zap"
)

type MemAllocateFunc = func(size int32) int32
type WASMExtensionFactory func(VM, *zap.Logger) WASMExtension

type WASMExtension interface {
	SetupWASMHostModule(module *wasmedge.Module)
}

type VM interface {
	CurrentRequest() *Request
	SetPanic(error)
	Allocate(size int32) int32
}
