package wasmquery

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
		expect      *ServiceConfig
		expectError bool
	}{
		{
			protoPath: "./testdata/test.proto",
			fqService: "sf.test.v1.TestService",
			expect: &ServiceConfig{
				fqn:            "sf.test.v1.TestService",
				connectWebPath: "/sf.test.v1.TestService/",
				Methods: []*MethodConfig{
					{
						name:           "TestGet",
						fqn:            "sf.test.v1.TestService.TestGet",
						connectWebPath: "/sf.test.v1.TestService/TestGet",
						exportName:     "sf_test_v1_testservice_testget",
					},
					{
						name:           "TestGetMany",
						fqn:            "sf.test.v1.TestService.TestGetMany",
						connectWebPath: "/sf.test.v1.TestService/TestGetMany",
						exportName:     "sf_test_v1_testservice_testgetmany",
					},
					{
						name:           "TestPrefix",
						fqn:            "sf.test.v1.TestService.TestPrefix",
						connectWebPath: "/sf.test.v1.TestService/TestPrefix",
						exportName:     "sf_test_v1_testservice_testprefix",
					},
					{
						name:           "TestScan",
						fqn:            "sf.test.v1.TestService.TestScan",
						connectWebPath: "/sf.test.v1.TestService/TestScan",
						exportName:     "sf_test_v1_testservice_testscan",
					},
				},
			},
		},
		{
			protoPath:   "./testdata/test.proto",
			fqService:   "sf.reader.v1.Unknown",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.protoPath, func(t *testing.T) {
			filedesc := protoFileToDescriptor(t, test.protoPath)
			config, err := NewServiceConfig(filedesc, test.fqService)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expect, config)
			}

		})
	}
}

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
