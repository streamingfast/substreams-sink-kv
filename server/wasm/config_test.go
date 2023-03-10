package wasm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/test-go/testify/require"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		protoPath string
		expect    *Config
	}{
		{
			protoPath: "./testdata/wasmquery/proto/reader.proto",
			expect: &Config{
				Services: []*ServiceConfig{
					{
						FQGRPCName: "sf.reader.v1.Eth",
						Methods: []*MethodConfig{
							{
								Name:       "Get",
								FQGRPCName: "sf.reader.v1.Eth.Get",
								ExportName: "sf_reader_v1_eth_get",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.protoPath, func(t *testing.T) {
			filedesc := protoFileToDescriptor(t, test.protoPath)
			assert.Equal(t, test.expect, NewConfig(filedesc))
		})
	}
}

// TODO: add more testing
// TODO: do validation that there are only letters and digits left
func Test_exportNameFromFQGrpcMethod(t *testing.T) {
	tests := []struct {
		grpcPath string
		expect   string
	}{
		{"sf.mycustomer.v1.Eth.Transfers", "sf_mycustomer_v1_eth_transfers"},
	}

	for _, test := range tests {
		t.Run(test.grpcPath, func(t *testing.T) {
			assert.Equal(t, test.expect, exportNameFromFQGrpcMethod(test.grpcPath))
		})
	}
}

func protoFileToDescriptor(t *testing.T, protoPath string) *descriptorpb.FileDescriptorProto {
	parser := &protoparse.Parser{
		ImportPaths:           []string{},
		IncludeSourceCodeInfo: true,
	}

	customFiles, err := parser.ParseFiles(protoPath)
	require.NoError(t, err)
	require.Equal(t, 1, len(customFiles))
	return customFiles[0].AsFileDescriptorProto()
}
