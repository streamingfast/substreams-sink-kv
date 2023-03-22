package wasmquery

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

	wasi := vm.GetImportModule(wasmedge.WASI)
	wasi.InitWasi(nil, nil, nil)

	if err := loadCode(vm); err != nil {
		return nil, fmt.Errorf("load wasm: %w", err)
	}

	if err := vm.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	bg := bindgen.New(vm)
	if err := bg.GetVm().Instantiate(); err != nil {
		return nil, fmt.Errorf("error instantiating VM: %w", err)
	}

	// storing this for validation
	fnames, _ := vm.GetFunctionList()
	for _, fname := range fnames {
		fmt.Println(fname)
		e.functionList[fname] = true
	}

	//e.bg = bg
	e.vm = vm

	return e, nil
}

func (e *Engine) RegisterIntrinsic() {



}

func (e *Engine) Instantiate(exportName string) (*Instance, error) {
	bg := bindgen.New(e.vm)
	if err := bg.GetVm().Instantiate(); err != nil {
		return nil, fmt.Errorf("error instantiating VM: %w", err)
	}
	return &Instance{
		bg: bg,
	}, nil

	//
	//
	//
	//// storing this for validation
	//fnames, _ := e.vm.GetFunctionList()
	//for _, fname := range fnames {
	//	e.functionList[fname] = true
	//}
}
