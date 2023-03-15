package wasm

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

type Config struct {
	FQGRPCServiceName string
	Methods           []*MethodConfig
}

type MethodConfig struct {
	Name       string
	FQGRPCName string // fully qualified grpc name
	ExportName string // will match the name of the function in the wasm code
}

func NewConfig(protoFile *descriptorpb.FileDescriptorProto, fqGRPCService string) (*Config, error) {
	for _, srv := range protoFile.Service {
		servName := fmt.Sprintf("%s.%s", protoFile.GetPackage(), srv.GetName())
		if servName != fqGRPCService {
			continue
		}
		c := &Config{
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

func exportNameFromFQGrpcMethod(fqGRPCMethod string) string {
	exportName := strings.Replace(fqGRPCMethod, ".", "_", -1)
	exportName = strings.Replace(exportName, "/", "_", -1)
	return strings.ToLower(exportName)
}
