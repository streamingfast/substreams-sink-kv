package wasm

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/descriptorpb"
)

type Config struct {
	Services []*ServiceConfig
}

type ServiceConfig struct {
	FQGRPCName string
	Methods    []*MethodConfig
}

type MethodConfig struct {
	Name       string
	FQGRPCName string // fully qualified grpc name
	ExportName string // will match the name of the function in the wasm code
}

func NewConfig(protoFile *descriptorpb.FileDescriptorProto) *Config {
	c := &Config{}
	for _, srv := range protoFile.Service {
		s := &ServiceConfig{
			FQGRPCName: fmt.Sprintf("%s.%s", protoFile.GetPackage(), srv.GetName()),
		}
		for _, mth := range srv.Method {
			fqGRPC := fmt.Sprintf("%s.%s", s.FQGRPCName, mth.GetName())
			s.Methods = append(s.Methods, &MethodConfig{
				Name:       mth.GetName(),
				FQGRPCName: fqGRPC,
				ExportName: exportNameFromFQGrpcMethod(fqGRPC),
			})
		}
		c.Services = append(c.Services, s)
	}
	return c
}

func exportNameFromFQGrpcMethod(fqGRPCMethod string) string {
	exportName := strings.Replace(fqGRPCMethod, ".", "_", -1)
	exportName = strings.Replace(exportName, "/", "_", -1)
	return strings.ToLower(exportName)
}
