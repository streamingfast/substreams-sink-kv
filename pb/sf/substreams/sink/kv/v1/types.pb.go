// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: sf/substreams/sink/kv/v1/types.proto

package kvv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type KVPairs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Pairs []*KVPair `protobuf:"bytes,2,rep,name=pairs,proto3" json:"pairs,omitempty"`
}

func (x *KVPairs) Reset() {
	*x = KVPairs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sf_substreams_sink_kv_v1_types_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVPairs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVPairs) ProtoMessage() {}

func (x *KVPairs) ProtoReflect() protoreflect.Message {
	mi := &file_sf_substreams_sink_kv_v1_types_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVPairs.ProtoReflect.Descriptor instead.
func (*KVPairs) Descriptor() ([]byte, []int) {
	return file_sf_substreams_sink_kv_v1_types_proto_rawDescGZIP(), []int{0}
}

func (x *KVPairs) GetPairs() []*KVPair {
	if x != nil {
		return x.Pairs
	}
	return nil
}

type KVPair struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *KVPair) Reset() {
	*x = KVPair{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sf_substreams_sink_kv_v1_types_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVPair) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVPair) ProtoMessage() {}

func (x *KVPair) ProtoReflect() protoreflect.Message {
	mi := &file_sf_substreams_sink_kv_v1_types_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVPair.ProtoReflect.Descriptor instead.
func (*KVPair) Descriptor() ([]byte, []int) {
	return file_sf_substreams_sink_kv_v1_types_proto_rawDescGZIP(), []int{1}
}

func (x *KVPair) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *KVPair) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type KVKeys struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keys []string `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
}

func (x *KVKeys) Reset() {
	*x = KVKeys{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sf_substreams_sink_kv_v1_types_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KVKeys) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KVKeys) ProtoMessage() {}

func (x *KVKeys) ProtoReflect() protoreflect.Message {
	mi := &file_sf_substreams_sink_kv_v1_types_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KVKeys.ProtoReflect.Descriptor instead.
func (*KVKeys) Descriptor() ([]byte, []int) {
	return file_sf_substreams_sink_kv_v1_types_proto_rawDescGZIP(), []int{2}
}

func (x *KVKeys) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

var File_sf_substreams_sink_kv_v1_types_proto protoreflect.FileDescriptor

var file_sf_substreams_sink_kv_v1_types_proto_rawDesc = []byte{
	0x0a, 0x24, 0x73, 0x66, 0x2f, 0x73, 0x75, 0x62, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2f,
	0x73, 0x69, 0x6e, 0x6b, 0x2f, 0x6b, 0x76, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x73, 0x66, 0x2e, 0x73, 0x75, 0x62, 0x73, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x73, 0x2e, 0x73, 0x69, 0x6e, 0x6b, 0x2e, 0x6b, 0x76, 0x2e, 0x76, 0x31,
	0x22, 0x41, 0x0a, 0x07, 0x4b, 0x56, 0x50, 0x61, 0x69, 0x72, 0x73, 0x12, 0x36, 0x0a, 0x05, 0x70,
	0x61, 0x69, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x73, 0x66, 0x2e,
	0x73, 0x75, 0x62, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2e, 0x73, 0x69, 0x6e, 0x6b, 0x2e,
	0x6b, 0x76, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x56, 0x50, 0x61, 0x69, 0x72, 0x52, 0x05, 0x70, 0x61,
	0x69, 0x72, 0x73, 0x22, 0x30, 0x0a, 0x06, 0x4b, 0x56, 0x50, 0x61, 0x69, 0x72, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x1c, 0x0a, 0x06, 0x4b, 0x56, 0x4b, 0x65, 0x79, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6b,
	0x65, 0x79, 0x73, 0x42, 0xfd, 0x01, 0x0a, 0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x73, 0x66, 0x2e, 0x73,
	0x75, 0x62, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2e, 0x73, 0x69, 0x6e, 0x6b, 0x2e, 0x6b,
	0x76, 0x2e, 0x76, 0x31, 0x42, 0x0a, 0x54, 0x79, 0x70, 0x65, 0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x4c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x69, 0x6e, 0x67, 0x66, 0x61, 0x73, 0x74, 0x2f, 0x73, 0x75, 0x62,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2d, 0x73, 0x69, 0x6e, 0x6b, 0x2d, 0x6b, 0x76, 0x2f,
	0x70, 0x62, 0x2f, 0x73, 0x66, 0x2f, 0x73, 0x75, 0x62, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73,
	0x2f, 0x73, 0x69, 0x6e, 0x6b, 0x2f, 0x6b, 0x76, 0x2f, 0x76, 0x31, 0x3b, 0x6b, 0x76, 0x76, 0x31,
	0xa2, 0x02, 0x04, 0x53, 0x53, 0x53, 0x4b, 0xaa, 0x02, 0x18, 0x53, 0x66, 0x2e, 0x53, 0x75, 0x62,
	0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x2e, 0x53, 0x69, 0x6e, 0x6b, 0x2e, 0x4b, 0x76, 0x2e,
	0x56, 0x31, 0xca, 0x02, 0x18, 0x53, 0x66, 0x5c, 0x53, 0x75, 0x62, 0x73, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x73, 0x5c, 0x53, 0x69, 0x6e, 0x6b, 0x5c, 0x4b, 0x76, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x24,
	0x53, 0x66, 0x5c, 0x53, 0x75, 0x62, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x73, 0x5c, 0x53, 0x69,
	0x6e, 0x6b, 0x5c, 0x4b, 0x76, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x1c, 0x53, 0x66, 0x3a, 0x3a, 0x53, 0x75, 0x62, 0x73, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x73, 0x3a, 0x3a, 0x53, 0x69, 0x6e, 0x6b, 0x3a, 0x3a, 0x4b, 0x76, 0x3a,
	0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_sf_substreams_sink_kv_v1_types_proto_rawDescOnce sync.Once
	file_sf_substreams_sink_kv_v1_types_proto_rawDescData = file_sf_substreams_sink_kv_v1_types_proto_rawDesc
)

func file_sf_substreams_sink_kv_v1_types_proto_rawDescGZIP() []byte {
	file_sf_substreams_sink_kv_v1_types_proto_rawDescOnce.Do(func() {
		file_sf_substreams_sink_kv_v1_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_sf_substreams_sink_kv_v1_types_proto_rawDescData)
	})
	return file_sf_substreams_sink_kv_v1_types_proto_rawDescData
}

var file_sf_substreams_sink_kv_v1_types_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_sf_substreams_sink_kv_v1_types_proto_goTypes = []interface{}{
	(*KVPairs)(nil), // 0: sf.substreams.sink.kv.v1.KVPairs
	(*KVPair)(nil),  // 1: sf.substreams.sink.kv.v1.KVPair
	(*KVKeys)(nil),  // 2: sf.substreams.sink.kv.v1.KVKeys
}
var file_sf_substreams_sink_kv_v1_types_proto_depIdxs = []int32{
	1, // 0: sf.substreams.sink.kv.v1.KVPairs.pairs:type_name -> sf.substreams.sink.kv.v1.KVPair
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_sf_substreams_sink_kv_v1_types_proto_init() }
func file_sf_substreams_sink_kv_v1_types_proto_init() {
	if File_sf_substreams_sink_kv_v1_types_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sf_substreams_sink_kv_v1_types_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVPairs); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sf_substreams_sink_kv_v1_types_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVPair); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sf_substreams_sink_kv_v1_types_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KVKeys); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_sf_substreams_sink_kv_v1_types_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_sf_substreams_sink_kv_v1_types_proto_goTypes,
		DependencyIndexes: file_sf_substreams_sink_kv_v1_types_proto_depIdxs,
		MessageInfos:      file_sf_substreams_sink_kv_v1_types_proto_msgTypes,
	}.Build()
	File_sf_substreams_sink_kv_v1_types_proto = out.File
	file_sf_substreams_sink_kv_v1_types_proto_rawDesc = nil
	file_sf_substreams_sink_kv_v1_types_proto_goTypes = nil
	file_sf_substreams_sink_kv_v1_types_proto_depIdxs = nil
}
