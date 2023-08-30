package wasmquery

import (
	"context"
	"fmt"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"
	grpcservers "github.com/streamingfast/dgrpc/server"
	connectweb "github.com/streamingfast/dgrpc/server/connect-web"
	"go.uber.org/zap"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type server struct {
	srv    *connectweb.ConnectWebServer
	logger *zap.Logger
}

func (e *Engine) newServer(registerProto bool, config *ServiceConfig) (*server, error) {
	c := &server{logger: e.logger}

	opts := []connect_go.HandlerOption{
		connect_go.WithCodec(&JSONPassthroughCodec{}),
		connect_go.WithCodec(&ProtoPassthroughCodec{}),
	}
	mux := http.NewServeMux()

	for _, method := range config.Methods {
		h := newHandler(e, method, c.logger)
		mux.Handle(method.connectWebPath, connect_go.NewUnaryHandler(
			method.connectWebPath,
			h.Handler,
			opts...,
		))
	}

	handlerGetter := func(opts ...connect_go.HandlerOption) (string, http.Handler) {
		return config.connectWebPath, mux
	}

	// we need to add the protoFile to the global proto registry or else
	// connect-web over gRPC will not work. it uses the registry to
	// find the service/method called in question.
	// In testing, we would to disable this, since the Proto file is AUTOMATICALLY registered when we import the file
	if registerProto {
		registerProtoFile(config.protoFile)
	}

	servOpts := []grpcservers.Option{
		grpcservers.WithReflection(config.fqn),
		grpcservers.WithLogger(c.logger),
		grpcservers.WithPermissiveCORS(),
		grpcservers.WithHealthCheck(grpcservers.HealthCheckOverHTTP, func(_ context.Context) (isReady bool, out interface{}, err error) { return true, nil, nil }),
	}
	servOpts = append(servOpts, grpcservers.WithPlainTextServer())

	c.srv = connectweb.New([]connectweb.HandlerGetter{handlerGetter}, servOpts...)
	return c, nil
}

func (c *server) Launch(listenAddr string) {
	c.srv.Launch(listenAddr)
}

func (c *server) Shutdown() {
	c.srv.Shutdown(nil)
}

const (
	ContentTypeProto = iota
	ContentTypeJson
)

type ConnectWebRequest struct {
	data        []byte
	contentType int
}

func registerProtoFile(fd *descriptorpb.FileDescriptorProto) error {
	desc, err := protodesc.NewFile(fd, nil)
	if err != nil {
		return fmt.Errorf("invalid FileDescriptorProto %q: %v", fd.GetName(), err)
	}
	if err := protoregistry.GlobalFiles.RegisterFile(desc); err != nil {
		return fmt.Errorf("cannot register descriptor %q: %v", fd.GetName(), err)
	}
	return nil
}
