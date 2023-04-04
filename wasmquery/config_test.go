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
		name        string
		protoPath   string
		fqService   string
		apiPrefix   string
		expect      *ServiceConfig
		expectError bool
	}{
		{
			name:      "without api prefix",
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
			name:      "with api prefix",
			protoPath: "./testdata/test.proto",
			fqService: "sf.test.v1.TestService",
			apiPrefix: "/api",
			expect: &ServiceConfig{
				fqn:            "sf.test.v1.TestService",
				connectWebPath: "/api/sf.test.v1.TestService/",
				Methods: []*MethodConfig{
					{
						name:           "TestGet",
						fqn:            "sf.test.v1.TestService.TestGet",
						connectWebPath: "/api/sf.test.v1.TestService/TestGet",
						exportName:     "sf_test_v1_testservice_testget",
					},
					{
						name:           "TestGetMany",
						fqn:            "sf.test.v1.TestService.TestGetMany",
						connectWebPath: "/api/sf.test.v1.TestService/TestGetMany",
						exportName:     "sf_test_v1_testservice_testgetmany",
					},
					{
						name:           "TestPrefix",
						fqn:            "sf.test.v1.TestService.TestPrefix",
						connectWebPath: "/api/sf.test.v1.TestService/TestPrefix",
						exportName:     "sf_test_v1_testservice_testprefix",
					},
					{
						name:           "TestScan",
						fqn:            "sf.test.v1.TestService.TestScan",
						connectWebPath: "/api/sf.test.v1.TestService/TestScan",
						exportName:     "sf_test_v1_testservice_testscan",
					},
				},
			},
		},
		{
			name:        "service does not exists",
			protoPath:   "./testdata/test.proto",
			fqService:   "sf.reader.v1.Unknown",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filedesc := protoFileToDescriptor(t, test.protoPath)
			config, err := NewServiceConfig(filedesc, test.fqService, test.apiPrefix)
			if test.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expect.fqn, config.fqn)
				assert.Equal(t, test.expect.connectWebPath, config.connectWebPath)
				assertEqualMethods(t, test.expect.Methods, config.Methods)
			}
		})
	}
}
func assertEqualMethods(t *testing.T, expect, actual []*MethodConfig) {
	require.Equal(t, len(expect), len(actual))
	for i := 0; i < len(expect); i++ {
		assertEqualMethod(t, expect[i], actual[i])
	}
}

func assertEqualMethod(t *testing.T, expect, actual *MethodConfig) {
	assert.Equal(t, expect.name, actual.name)
	assert.Equal(t, expect.fqn, actual.fqn)
	assert.Equal(t, expect.exportName, actual.exportName)
	assert.Equal(t, expect.connectWebPath, actual.connectWebPath)

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

func Test_validatePrefix(t *testing.T) {
	tests := []struct {
		prefix      string
		expectError bool
	}{
		{"api", true},
		{"/api", false},
		{"/api/", true},
	}

	for _, test := range tests {
		t.Run(test.prefix, func(t *testing.T) {
			err := validatePrefix(test.prefix)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
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
