package wasmquery

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"

	"google.golang.org/protobuf/proto"
)

type JSONPassthroughCodec struct{}

func (c *JSONPassthroughCodec) Name() string { return "json" }
func (c *JSONPassthroughCodec) Marshal(a any) ([]byte, error) {
	el, ok := a.(*ConnectWebRequest)
	if ok {
		return el.data, nil
	}
	var options protojson.MarshalOptions
	return options.Marshal(a.(proto.Message))
}
func (c *JSONPassthroughCodec) Unmarshal(data []byte, a any) error {
	v, ok := a.(*ConnectWebRequest)
	if !ok {
		return fmt.Errorf("unpected object type")
	}
	v.contentType = ContentTypeJson
	v.data = data
	return nil
}

type ProtoPassthroughCodec struct{}

func (c *ProtoPassthroughCodec) Name() string { return "proto" }
func (c *ProtoPassthroughCodec) Marshal(a any) ([]byte, error) {
	el, ok := a.(*ConnectWebRequest)
	if ok {
		return el.data, nil
	}
	return proto.Marshal(a.(proto.Message))
}

func (c *ProtoPassthroughCodec) Unmarshal(data []byte, a any) error {
	v, ok := a.(*ConnectWebRequest)
	if !ok {
		return fmt.Errorf("unpected object type")
	}
	v.contentType = ContentTypeProto
	v.data = data
	return nil
}
