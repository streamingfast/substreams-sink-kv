package wasmquery

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

type EngineConfig struct {
	vmCount       uint64
	code          []byte
	codec         Codec
	serviceConfig *ServiceConfig
}

type ServiceConfig struct {
	FQGRPCServiceName string
	Methods           []*MethodConfig
}

type MethodConfig struct {
	Name       string
	FQGRPCName string // fully qualified grpc name
	ExportName string // will match the name of the function in the wasm code
}

func NewEngineConfig(count uint64, code []byte, serviceConfig *ServiceConfig) *EngineConfig {
	return NewEngineConfigWithCodec(count, code, serviceConfig, PassthroughCodec{})
}

func NewEngineConfigWithCodec(count uint64, code []byte, serviceConfig *ServiceConfig, codec Codec) *EngineConfig {
	return &EngineConfig{
		vmCount:       count,
		code:          code,
		codec:         codec,
		serviceConfig: serviceConfig,
	}
}

func NewServiceConfig(
	protoFile *descriptorpb.FileDescriptorProto,
	fqGRPCService string,
) (*ServiceConfig, error) {
	for _, srv := range protoFile.Service {
		servName := fmt.Sprintf("%s.%s", protoFile.GetPackage(), srv.GetName())
		if servName != fqGRPCService {
			continue
		}
		c := &ServiceConfig{
			FQGRPCServiceName: servName,
		}
		for _, mth := range srv.Method {
			fqGRPC := fmt.Sprintf("%s.%s", c.FQGRPCServiceName, mth.GetName())
			c.Methods = append(c.Methods, &MethodConfig{
				Name:       mth.GetName(),
				FQGRPCName: fqGRPC,
				ExportName: exportNameFromFQGrpcMethod(fqGRPC),
			})
		}
		return c, nil
	}
	return nil, fmt.Errorf("unable to find grpc service %q in proto file", fqGRPCService)
}

func (s *ServiceConfig) getWASMFunctionNames() (out []string) {
	for _, method := range s.Methods {
		out = append(out, method.ExportName)
	}
	return out
}

func exportNameFromFQGrpcMethod(fqGRPCMethod string) string {
	exportName := strings.Replace(fqGRPCMethod, ".", "_", -1)
	exportName = strings.Replace(exportName, "/", "_", -1)
	return strings.ToLower(exportName)
}
