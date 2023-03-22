package wasmquery

import (
	"github.com/second-state/WasmEdge-go/wasmedge"
)

type Intrinsics interface {
	WASMModule() *wasmedge.Module
}

type VM interface {
	SetPanic(error)
	Allocate()
}
