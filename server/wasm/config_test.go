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
		protoPath   string
		fqService   string
		expect      *Config
		expectError bool
	}{
		{
			protoPath: "./testdata/wasmquery/test.proto",
			fqService: "sf.test.v1.TestService",
			expect: &Config{
				FQGRPCServiceName: "sf.test.v1.TestService",
				Methods: []*MethodConfig{
					{
						Name:       "TestGet",
						FQGRPCName: "sf.test.v1.TestService.TestGet",
						ExportName: "sf_test_v1_testservice_testget",
					},
					{
						Name:       "TestGetMany",
						FQGRPCName: "sf.test.v1.TestService.TestGetMany",
						ExportName: "sf_test_v1_testservice_testgetmany",
					},
					{
						Name:       "TestPrefix",
						FQGRPCName: "sf.test.v1.TestService.TestPrefix",
						ExportName: "sf_test_v1_testservice_testprefix",
					},
					{
						Name:       "TestScan",
						FQGRPCName: "sf.test.v1.TestService.TestScan",
						ExportName: "sf_test_v1_testservice_testscan",
					},
				},
			},
		},
		{
			protoPath:   "./testdata/wasmquery/test.proto",
			fqService:   "sf.reader.v1.Unknown",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.protoPath, func(t *testing.T) {
			filedesc := protoFileToDescriptor(t, test.protoPath)
			config, err := NewConfig(filedesc, test.fqService)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expect, config)
			}

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
