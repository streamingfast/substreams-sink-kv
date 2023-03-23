package wasmquery

import (
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
	bindgen "github.com/second-state/wasmedge-bindgen/host/go"
	"go.uber.org/zap"
)

type vm struct {
	vm *wasmedge.VM
	id uint64

	instance *bindgen.Bindgen
	bg       *bindgen.Bindgen

	currentRequest *Request

	functionList     map[string]bool
	allocateFuncName string
	logger           *zap.Logger
	panic            *PanicErr
}

func newVMFromFile(id uint64, wasmFilepath string, logger *zap.Logger) (*vm, error) {
	return newVM(id, func(vm *wasmedge.VM) error {
		return vm.LoadWasmFile(wasmFilepath)
	}, logger)
}

func newVMFromBytes(id uint64, code []byte, logger *zap.Logger) (*vm, error) {
	return newVM(id, func(vm *wasmedge.VM) error {
		return vm.LoadWasmBuffer(code)
	}, logger)
}

func newVM(id uint64, loadCode func(vm *wasmedge.VM) error, logger *zap.Logger) (*vm, error) {

	conf := wasmedge.NewConfigure(wasmedge.WASI)

	vm := &vm{
		id:               id,
		allocateFuncName: "allocate",
		vm:               wasmedge.NewVMWithConfig(conf),
		functionList:     map[string]bool{},
		logger:           logger.Named("vm").With(zap.Uint64("vm_id", id)),
	}

	wasi := vm.vm.GetImportModule(wasmedge.WASI)
	wasi.InitWasi(nil, nil, nil)

	if err := loadCode(vm.vm); err != nil {
		return nil, fmt.Errorf("load wasm: %w", err)
	}

	if err := vm.vm.Validate(); err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	return vm, nil
}

func (v *vm) loggerFields() []zap.Field {
	return []zap.Field{
		zap.Uint64("vm_id", v.id),
	}
}
func (v *vm) getFunctionList() map[string]bool {
	functionNames, _ := v.vm.GetFunctionList()
	out := map[string]bool{}

	for _, functionName := range functionNames {
		out[functionName] = true
	}
	return out
}

func (v *vm) instantiate(requiredFunctionNames []string) error {
	bg := bindgen.New(v.vm)
	if err := bg.GetVm().Instantiate(); err != nil {
		return fmt.Errorf("error instantiating vm: %w", err)
	}

	functions := v.getFunctionList()

	for _, fname := range requiredFunctionNames {
		if _, found := functions[fname]; !found {
			return fmt.Errorf("wasm code missing function %q", fname)
		}
	}

	v.bg = bg
	return nil
}

func (v *vm) execute(request *Request, exportName string, input []byte) ([]interface{}, interface{}, error) {
	v.currentRequest = request
	v.panic = nil
	return v.bg.Execute(exportName, input)
}

func (v *vm) registerHost(i WASMExtension) error {
	hostModule := v.getHostModule()
	i.SetupWASMHostModule(hostModule)
	return v.vm.RegisterModule(hostModule)
}

func (v *vm) Allocate(size int32) int32 {
	allocateResult, err := v.vm.Execute(v.allocateFuncName, size)
	if err != nil {
		panic(err)
	}
	pointerOfPointers := allocateResult[0].(int32)
	return pointerOfPointers
}
