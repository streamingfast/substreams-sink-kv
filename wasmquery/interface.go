package wasmquery

import (
	"fmt"

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
	Allocate(size int32) int32
}

type PanicErr struct {
	message      string
	filename     string
	lineNumber   int32
	columnNumber int32
}

func (e *PanicErr) Error() string {
	return fmt.Sprintf("panic in wasm: %q at %s:%d:%d", e.message, e.filename, e.lineNumber, e.columnNumber)
}
