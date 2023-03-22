package wasmquery

import (
	"context"
	"fmt"
	"time"

	"github.com/streamingfast/dgrpc/server"
	"github.com/streamingfast/dgrpc/server/standard"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

type Engine struct {
	vmPool chan *vm
	srv    *standard.StandardServer
	logger *zap.Logger
}

func NewEngine(config *EngineConfig, extensionFactory WASMExtensionFactory, logger *zap.Logger) (*Engine, error) {
	logger.Info("initializing wasm query engine", zap.Uint64("vm_count", config.vmCount))

	eng := &Engine{
		vmPool: make(chan *vm, config.vmCount),
		logger: logger,
	}

	if err := eng.setupGRPCServer(config.codec, config.serviceConfig); err != nil {
		return nil, fmt.Errorf("failed to setup grpc server")
	}

	for i := uint64(0); i < config.vmCount; i++ {
		v, err := newVMFromBytes(i, config.code, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to create vm: %w", err)
		}

		wasmExtension := extensionFactory(v, eng.logger)
		if err := v.register(wasmExtension); err != nil {
			return nil, fmt.Errorf("failed to register intrinsic to module: %w", err)
		}

		if err := v.instantiate(config.serviceConfig.getWASMFunctionNames()); err != nil {
			return nil, fmt.Errorf("failed to instantiate vm: %w", err)
		}

		eng.vmPool <- v
	}

	return eng, nil
}

func (e *Engine) Shutdown() {
	e.logger.Info("wasn server received shutdown, shutting down server")
	e.srv.Shutdown(5 * time.Second)
}

func (e *Engine) Serve(listenAddr string) error {
	e.srv.Launch(listenAddr)
	return nil
}

func (e *Engine) setupGRPCServer(codec Codec, config *ServiceConfig) error {
	encoding.RegisterCodec(codec)
	srv := standard.NewServer(server.NewOptions())

	grpcServer := srv.GrpcServer()

	var v interface{}
	grpcService := &grpc.ServiceDesc{
		ServiceName: config.FQGRPCServiceName,
		//HandlerType: wasmEngine,
		Methods: []grpc.MethodDesc{},
	}

	for _, methodConfig := range config.Methods {
		handler, err := newHandler(methodConfig, e, codec, e.logger)
		if err != nil {
			return fmt.Errorf("failed to get handler: %w", err)
		}

		grpcService.Streams = append(grpcService.Streams, grpc.StreamDesc{
			StreamName:    methodConfig.Name,
			Handler:       handler.handle,
			ServerStreams: true,
		})
	}

	grpcServer.RegisterService(grpcService, v)
	e.srv = srv

	return nil
}

func (p *Engine) borrowVM(ctx context.Context) *vm {
	select {
	case <-ctx.Done():
		return nil
	case w := <-p.vmPool:
		return w
	}
}

func (p *Engine) returnVM(vm *vm) {
	p.vmPool <- vm
}
