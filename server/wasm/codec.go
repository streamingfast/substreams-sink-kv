package wasm

// Codec

type passthroughCodec struct{}

func (passthroughCodec) Marshal(v interface{}) ([]byte, error) {
	return v.(*passthroughBytes).Bytes, nil
}

func (passthroughCodec) Unmarshal(data []byte, v interface{}) error {
	el := v.(*passthroughBytes)
	el.Bytes = data
	return nil
}

func (passthroughCodec) Name() string { return "proto" }

// Passing bytes around

type passthroughBytes struct {
	Bytes []byte
}

func NewPassthroughBytes() *passthroughBytes {
	return &passthroughBytes{}
}
func (b *passthroughBytes) Set(in []byte) {
	b.Bytes = in
}
