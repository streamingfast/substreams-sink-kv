package wasmquery

import bindgen "github.com/second-state/wasmedge-bindgen/host/go"

type Instance struct {
	bg *bindgen.Bindgen
}

func (i* Instance) Execute(exportName string, input []byte) ([]interface{}, interface{}, error){
	return i.bg.Execute(exportName, input)
}


func (i* Instance) SetPanic(err error) {
	i.panic = err
}

