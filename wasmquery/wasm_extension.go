package wasmquery

import "github.com/second-state/WasmEdge-go/wasmedge"

func (v *vm) registerPanic(_ interface{}, callframe *wasmedge.CallingFrame, params []interface{}) ([]interface{}, wasmedge.Result) {
	mem := callframe.GetMemoryByIndex(0)

	msgPtr := params[0].(int32)
	msgSize := params[1].(int32)
	data, _ := mem.GetData(uint(msgPtr), uint(msgSize))
	errorMessage := make([]byte, msgSize)
	copy(errorMessage, data)

	filePtr := params[2].(int32)
	fileSize := params[3].(int32)
	data, _ = mem.GetData(uint(filePtr), uint(fileSize))
	filename := make([]byte, msgSize)
	copy(filename, data)

	v.panic = &PanicErr{
		message:      string(errorMessage),
		filename:     string(filename),
		lineNumber:   params[4].(int32),
		columnNumber: params[5].(int32),
	}
	return []interface{}{1}, wasmedge.Result_Fail
}

func (v *vm) getHostModule() *wasmedge.Module {
	hostModule := wasmedge.NewModule("host")

	registerPanicType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{},
	)

	registerPanicFunc := wasmedge.NewFunction(registerPanicType, v.registerPanic, nil, 0)
	registerPanicType.Release()
	hostModule.AddFunction("register_panic", registerPanicFunc)
	return hostModule
}
