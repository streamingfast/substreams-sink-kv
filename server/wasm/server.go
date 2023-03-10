package wasm

import (
	"fmt"
	"time"

	"google.golang.org/grpc"

	"google.golang.org/grpc/encoding"

	"github.com/streamingfast/dgrpc/server"
	"github.com/streamingfast/dgrpc/server/standard"
	sserver "github.com/streamingfast/substreams-sink-kv/server"
	"go.uber.org/zap"
)

var _ sserver.Serveable = (*Server)(nil)

func NewServer(config *Config, wasmEngine *Engine, logger *zap.Logger) (*Server, error) {
	encoding.RegisterCodec(passthroughCodec{})
	srv := standard.NewServer(server.NewOptions())

	grpcServer := srv.GrpcServer()

	var v interface{}
	for _, srvConfig := range config.Services {
		grpcService := &grpc.ServiceDesc{
			ServiceName: srvConfig.FQGRPCName,
			HandlerType: wasmEngine,
			Methods:     []grpc.MethodDesc{},
			//TODO: is this usefull?
			//Metadata: "sf/mycustomer/v1/eth.proto",
		}

		for _, methodConfig := range srvConfig.Methods {
			handler, err := wasmEngine.GetHandler(methodConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to get handler: %w", err)
			}
			grpcService.Streams = append(grpcService.Streams, grpc.StreamDesc{
				StreamName:    methodConfig.Name,
				Handler:       handler.handle,
				ServerStreams: true,
			})
		}

		grpcServer.RegisterService(grpcService, v)
	}

	s := &Server{
		logger: logger,
		srv:    srv,
	}

	return s, nil
}

type Server struct {
	logger *zap.Logger
	srv    *standard.StandardServer
}

func (s *Server) Shutdown() {
	s.logger.Info("wasn server received shutdown, shutting down server")
	s.srv.Shutdown(5 * time.Second)
}

func (s *Server) Serve(listenAddr string) error {
	s.srv.Launch(listenAddr)
	return nil
}
