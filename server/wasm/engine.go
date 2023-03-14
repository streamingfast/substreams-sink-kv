package wasm

import (
	"context"
	"encoding/binary"
	"errors"
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

	// registering getKey
	hostgetftype := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	// registering prefixScan
	hostprefixftype := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_ExternRef,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostgetprint := wasmedge.NewFunction(hostgetftype, e.getKey, nil, 0)
	hostgetftype.Release()

	hostprefixprint := wasmedge.NewFunction(hostprefixftype, e.prefixScan, nil, 0)
	hostprefixftype.Release()

	impobj.AddFunction("get_key", hostgetprint)
	impobj.AddFunction("prefix_scan", hostprefixprint)

	// TODO: add scan, and prefix support
	vm.RegisterModule(impobj)

}

func (e *Engine) prefixScan(_ interface{}, callFrame *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callFrame.GetMemoryByIndex(0)

	prefix := params[0].(string)
	//keyPtr := params[1].(int32)
	//keySize := params[2].(int32)

	//data, _ := mem.GetData(uint(keyPtr), uint(keySize))
	//
	//keys := make([][]byte, keySize)
	//
	//copy(keys, data)

	keyVals, _, err := e.kv.Scan(context.Background(), prefix, "", 20)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			e.logger.Debug("no values found", zap.Error(err))
			return nil, wasmedge.Result_Fail
		}
		if errors.Is(err, db.ErrInvalidArguments) {
			e.logger.Debug("invalid arguments", zap.Error(err))
			return nil, wasmedge.Result_Fail
		}
		e.logger.Info("internal error", zap.Error(err))
		return nil, wasmedge.Result_Fail
	}

	var valsPtr []int32
	var dataArray [][]byte
	for i, val := range keyVals {
		valsPtr = append(valsPtr, e.allocate(int32(len(val.Value))))

		data, err := mem.GetData(uint(valsPtr[i]), uint(len(val.Value)))
		if err != nil {
			e.logger.Debug(fmt.Sprintf("getting data from key: %s", val.Key), zap.Error(err))
		}
		dataArray = append(dataArray, data)

		copy(dataArray[i], val.Value)
	}

	return []interface{}{1}, wasmedge.Result_Success
}

func (e *Engine) getKey(_ interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	// As in: https://github.com/second-state/WasmEdge-go-examples/blob/master/go_HostFunc/hostfunc.go
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

func (e *Engine) allocate(size int32) int32 {
	allocateResult, err := e.vm.Execute(e.allocateFuncName, size)
	if err != nil {
		panic(err)
	}
	pointerOfPointers := allocateResult[0].(int32)
	return pointerOfPointers
}
