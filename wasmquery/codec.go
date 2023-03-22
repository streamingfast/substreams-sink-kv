package wasmquery

import "google.golang.org/grpc/encoding"

// codec

type Codec interface {
	encoding.Codec

	NewMessage() Byteable
}
type Byteable interface {
	Set(in []byte)
	Bytes() []byte
}

var _ Codec = PassthroughCodec{}

type PassthroughCodec struct{}

func (PassthroughCodec) Marshal(v interface{}) ([]byte, error) {
	return v.(*passthroughBytes).bytes, nil
}

func (PassthroughCodec) Unmarshal(data []byte, v interface{}) error {
	el := v.(*passthroughBytes)
	el.bytes = data
	return nil
}

func (PassthroughCodec) Name() string { return "proto" }

func (c PassthroughCodec) NewMessage() Byteable {
	return &passthroughBytes{}
}

// Passing bytes around

type passthroughBytes struct {
	bytes []byte
}

func (b *passthroughBytes) Set(in []byte) {
	b.bytes = in
}

func (b *passthroughBytes) Bytes() []byte {
	return b.bytes
}
