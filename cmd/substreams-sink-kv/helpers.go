package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/streamingfast/substreams/manifest"
	pbsubstreams "github.com/streamingfast/substreams/pb/sf/substreams/v1"
	"github.com/streamingfast/substreams/pb/system"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
)

//func (p *addSinkConfigToSpkg()

//func LoadProtobufs(pkg *pbsubstreams.Package, manif *manifest.Manifest) ([]*desc.FileDescriptor, error) {
//	// System protos
//	systemFiles, err := readSystemProtobufs()
//	if err != nil {
//		return nil, err
//	}
//	seen := map[string]bool{}
//	for _, file := range systemFiles.File {
//		pkg.ProtoFiles = append(pkg.ProtoFiles, file)
//		seen[*file.Name] = true
//	}
//
//	var importPaths []string
//	for _, imp := range manif.Protobuf.ImportPaths {
//		importPaths = append(importPaths, manif.resolvePath(imp))
//	}
//
//	// The manifest's root directory is always added to the list of import paths so that
//	// files specified relative to the manifest's directory works properly. It is added last
//	// so that if user's specified import paths contains the file, it's picked from their
//	// import paths instead of the implicitly added folder.
//	if manif.Workdir != "" {
//		importPaths = append(importPaths, manif.Workdir)
//	}
//
//	// User-specified protos
//	parser := &protoparse.Parser{
//		ImportPaths:           importPaths,
//		IncludeSourceCodeInfo: true,
//	}
//
//	for _, file := range manif.Protobuf.Files {
//		if seen[file] {
//			return nil, fmt.Errorf("WARNING: proto file %s already exists in system protobufs, do not include it in your manifest", file)
//		}
//	}
//
//	customFiles, err := parser.ParseFiles(manif.Protobuf.Files...)
//	if err != nil {
//		return nil, fmt.Errorf("error parsing proto files %q (import paths: %q): %w", manif.Protobuf.Files, importPaths, err)
//	}
//	for _, fd := range customFiles {
//		pkg.ProtoFiles = append(pkg.ProtoFiles, fd.AsFileDescriptorProto())
//	}
//
//	return customFiles, nil
//}

func readSystemProtobufs() (*descriptorpb.FileDescriptorSet, error) {
	fds := &descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(system.ProtobufDescriptors, fds)
	if err != nil {
		return nil, err
	}

	return fds, nil
}

func LoadSinkConfigs(pkg *pbsubstreams.Package, m *manifest.Manifest, protoDescs []*desc.FileDescriptor) error {
	if m.Sink == nil {
		return nil
	}
	if m.Sink.Module == "" {
		return errors.New(`sink: "module" unspecified`)
	}
	if m.Sink.Type == "" {
		return errors.New(`sink: "type" unspecified`)
	}
	pkg.SinkModule = m.Sink.Module
	jsonConfig := convertYAMLtoJSONCompat(m.Sink.Config)
	jsonConfigBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		return fmt.Errorf("sink: config: error marshalling to json: %w", err)
	}

	var found bool
files:
	for _, file := range protoDescs {
		for _, msgDesc := range file.GetMessageTypes() {

			fmt.Println("Type found:", file.GetName(), msgDesc.GetFullyQualifiedName())
			if msgDesc.GetFullyQualifiedName() == m.Sink.Type {
				dynConf := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(msgDesc)
				if err := dynConf.UnmarshalJSON(jsonConfigBytes); err != nil {
					return fmt.Errorf("sink: config: encoding json into protobuf message: %w", err)
				}
				pbBytes, err := dynConf.Marshal()
				if err != nil {
					return fmt.Errorf("sink: config: encoding protobuf from dynamic message: %w", err)
				}
				pkg.SinkConfig = &anypb.Any{
					TypeUrl: m.Sink.Type,
					Value:   pbBytes,
				}
				found = true
				break files
			}
		}
	}
	if !found {
		return fmt.Errorf("sink: type: could not find protobuf message type %q in bundled protobuf descriptors", m.Sink.Type)
	}

	return nil
}

func convertYAMLtoJSONCompat(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convertYAMLtoJSONCompat(v)
		}
		return m2
	case map[string]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k] = convertYAMLtoJSONCompat(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convertYAMLtoJSONCompat(v)
		}
	}
	return i
}
