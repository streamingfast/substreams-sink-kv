package wasmquery

import (
	"fmt"
	"strings"

	"go.uber.org/zap"

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

func (m *MethodConfig) loggerFields(asRESTApi bool) []zap.Field {
	apiType := "gRPC"
	path := m.fqn
	if asRESTApi {
		apiType = "REST"
		path = m.connectWebPath
	}
	return []zap.Field{
		zap.String("api_type", apiType),
		zap.String("path", path),
	}
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
	apiPrefix string,
) (*ServiceConfig, error) {
	if err := validatePrefix(apiPrefix); err != nil {
		return nil, fmt.Errorf("invalid api prefix %q: %w", apiPrefix, err)
	}

	fd, err := desc.CreateFileDescriptor(protoFile)
	if err != nil {
		return nil, fmt.Errorf("file descriptor: %w", err)
	}

	for _, srv := range fd.GetServices() {
		servName := fmt.Sprintf("%s.%s", protoFile.GetPackage(), srv.GetName())
		if servName != fqGRPCService {
			continue
		}

		connectWebPath := fmt.Sprintf("/%s/", servName)
		if apiPrefix != "" {
			connectWebPath = fmt.Sprintf("%s%s", apiPrefix, connectWebPath)
		}
		c := &ServiceConfig{
			fqn:            servName,
			connectWebPath: connectWebPath,
			protoFile:      protoFile,
		}
		for _, mth := range srv.GetMethods() {
			fqGRPC := fmt.Sprintf("%s.%s", c.fqn, mth.GetName())
			connectWebPath := fmt.Sprintf("/%s/%s", c.fqn, mth.GetName())
			if apiPrefix != "" {
				connectWebPath = fmt.Sprintf("%s%s", apiPrefix, connectWebPath)
			}
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

func validatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	if !strings.HasPrefix(prefix, "/") {
		return fmt.Errorf("api prefix must start with a '/' and cannot end with a '/")
	}
	if strings.HasSuffix(prefix, "/") {
		return fmt.Errorf("api prefix must start with a '/' and cannot end with a '/")
	}
	return nil
}
