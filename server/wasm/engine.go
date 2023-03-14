package wasm

import (
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
	bindgen "github.com/second-state/wasmedge-bindgen/host/go"
	"github.com/streamingfast/substreams-sink-kv/db"
	"go.uber.org/zap"
)

type Engine struct {
	bg *bindgen.Bindgen
	vm *wasmedge.VM
	kv db.Reader

	functionList     map[string]bool
	allocateFuncName string
	logger           *zap.Logger
}

func NewEngineFromFile(wasmFilepath string, dbReader db.Reader, logger *zap.Logger) (*Engine, error) {
	return newEngine(dbReader, func(vm *wasmedge.VM) error {
		return vm.LoadWasmFile(wasmFilepath)
	}, logger)
}
func NewEngineFromBytes(code []byte, dbReader db.Reader, logger *zap.Logger) (*Engine, error) {
	return newEngine(dbReader, func(vm *wasmedge.VM) error {
		return vm.LoadWasmBuffer(code)
	}, logger)
}

func newEngine(dbReader db.Reader, loadCode func(vm *wasmedge.VM) error, logger *zap.Logger) (*Engine, error) {
	e := &Engine{
		kv:               dbReader,
		allocateFuncName: "allocate",
		functionList:     map[string]bool{},
		logger:           logger,
	}
	wasmedge.SetLogErrorLevel()
	conf := wasmedge.NewConfigure(wasmedge.WASI)
	vm := wasmedge.NewVMWithConfig(conf)

	registerIntrinsics(vm, e)

	wasi := vm.GetImportModule(wasmedge.WASI)
	wasi.InitWasi(nil, nil, nil)

	if err := loadCode(vm); err != nil {
		return nil, fmt.Errorf("load wasm: %w", err)
	}

	if err := vm.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	bg := bindgen.New(vm)
	//bg.SetAllocateExport(srv.allocateFuncName)
	if err := bg.GetVm().Instantiate(); err != nil {
		return nil, fmt.Errorf("error instantiating VM: %w", err)
	}

	// storing this for validation
	fnames, _ := vm.GetFunctionList()
	for _, fname := range fnames {
		e.functionList[fname] = true
	}

	e.bg = bg
	e.vm = vm

	return e, nil
}

func registerIntrinsics(vm *wasmedge.VM, e *Engine) {
	impobj := wasmedge.NewModule("host")

	hostGetKeyType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostGetKeyFunc := wasmedge.NewFunction(hostGetKeyType, e.getKey, nil, 0)
	hostGetKeyType.Release()
	impobj.AddFunction("get_key", hostGetKeyFunc)

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

	hostGetByPrefixFunc := wasmedge.NewFunction(hostGetByPrefixType, e.getByPrefix, nil, 0)
	hostGetByPrefixType.Release()
	impobj.AddFunction("get_by_prefix", hostGetByPrefixFunc)

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

	hostScanFunc := wasmedge.NewFunction(hostScanType, e.scan, nil, 0)
	hostScanType.Release()
	impobj.AddFunction("scan", hostScanFunc)

	// TODO: add scan, and prefix support
	vm.RegisterModule(impobj)

}

func (e *Engine) allocate(size int32) int32 {
	allocateResult, err := e.vm.Execute(e.allocateFuncName, size)
	if err != nil {
		panic(err)
	}
	pointerOfPointers := allocateResult[0].(int32)
	return pointerOfPointers
}
