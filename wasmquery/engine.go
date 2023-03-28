package wasmquery

import "C"
import (
	"context"
	"fmt"

	"github.com/second-state/WasmEdge-go/wasmedge"
	"go.uber.org/zap"
)

type Engine struct {
	vmPool        chan *vm
	srv           *server
	logger        *zap.Logger
	registerProto bool
}

type Option func(*Engine) *Engine

func WithVMDebugLogLevel() Option {
	return func(engine *Engine) *Engine {
		wasmedge.SetLogDebugLevel()
		return engine
	}
}

func WithVMErrorLogLevel() Option {
	return func(engine *Engine) *Engine {
		wasmedge.SetLogErrorLevel()
		return engine
	}
}
func SkipProtoRegister() Option {
	return func(engine *Engine) *Engine {
		engine.registerProto = false
		return engine
	}
}
func NewEngine(config *EngineConfig, extensionFactory WASMExtensionFactory, logger *zap.Logger, opts ...Option) (*Engine, error) {
	logger.Info("initializing wasm query engine", zap.Uint64("vm_count", config.vmCount))

	eng := &Engine{
		vmPool:        make(chan *vm, config.vmCount),
		registerProto: true,
		logger:        logger,
	}

	for _, opt := range opts {
		eng = opt(eng)
	}

	srv, err := eng.newServer(eng.registerProto, config.serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to setup connect-web server: %w", err)
	}
	eng.srv = srv

	for i := uint64(0); i < config.vmCount; i++ {
		v, err := newVMFromBytes(i, config.code, logger)
		if err != nil {
			return nil, fmt.Errorf("unable to create vm: %w", err)
		}

		wasmExtension := extensionFactory(v, eng.logger)
		if err := v.registerHost(wasmExtension); err != nil {
			return nil, fmt.Errorf("failed to registerHost intrinsic to module: %w", err)
		}

		if err := v.instantiate(config.serviceConfig.getWASMFunctionNames()); err != nil {
			return nil, fmt.Errorf("failed to instantiate vm: %w", err)
		}

		eng.vmPool <- v
	}

	return eng, nil
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

func (e *Engine) Serve(listenAddr string) error {
	go e.srv.Launch(listenAddr)
	return nil
}
func (e *Engine) Shutdown() {
	e.srv.Shutdown()
}
