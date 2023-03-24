package wasmquery

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"

	"google.golang.org/protobuf/types/descriptorpb"
)

type EngineConfig struct {
	vmCount       uint64
	code          []byte
	serviceConfig *ServiceConfig
}

type ServiceConfig struct {
	fqn            string
	connectWebPath string
	Methods        []*MethodConfig
	protoFile      *descriptorpb.FileDescriptorProto
}

type MethodConfig struct {
	name           string
	fqn            string // fully qualified grpc name
	exportName     string // will match the name of the function in the wasm code
	connectWebPath string // represents the connect-web REST api path
	inputType      *desc.MessageDescriptor
	outputType     *desc.MessageDescriptor
}

func NewEngineConfig(count uint64, code []byte, serviceConfig *ServiceConfig) *EngineConfig {
	return &EngineConfig{
		vmCount:       count,
		code:          code,
		serviceConfig: serviceConfig,
	}
}

func NewServiceConfig(
	protoFile *descriptorpb.FileDescriptorProto,
	fqGRPCService string,
) (*ServiceConfig, error) {

	fd, err := desc.CreateFileDescriptor(protoFile)
	if err != nil {
		return nil, fmt.Errorf("file descriptor: %w", err)
	}

	for _, srv := range fd.GetServices() {
		servName := fmt.Sprintf("%s.%s", protoFile.GetPackage(), srv.GetName())
		if servName != fqGRPCService {
			continue
		}
		c := &ServiceConfig{
			fqn:            servName,
			connectWebPath: fmt.Sprintf("/%s/", servName),
			protoFile:      protoFile,
		}
		for _, mth := range srv.GetMethods() {
			fqGRPC := fmt.Sprintf("%s.%s", c.fqn, mth.GetName())
			connectWebPath := fmt.Sprintf("/%s/%s", c.fqn, mth.GetName())
			if mth.IsServerStreaming() {
				return nil, fmt.Errorf("unable to support GRPC stream %s", fqGRPC)
			}
			c.Methods = append(c.Methods, &MethodConfig{
				name:           mth.GetName(),
				fqn:            fqGRPC,
				exportName:     exportNameFromFQGrpcMethod(fqGRPC),
				connectWebPath: connectWebPath,
				inputType:      mth.GetInputType(),
				outputType:     mth.GetOutputType(),
			})
		}
		return c, nil
	}
	return nil, fmt.Errorf("unable to find grpc service %q in proto file", fqGRPCService)
}

func (s *ServiceConfig) getWASMFunctionNames() (out []string) {
	for _, method := range s.Methods {
		out = append(out, method.exportName)
	}
	return out
}

func exportNameFromFQGrpcMethod(fqGRPCMethod string) string {
	exportName := strings.Replace(fqGRPCMethod, ".", "_", -1)
	exportName = strings.Replace(exportName, "/", "_", -1)
	return strings.ToLower(exportName)
}
