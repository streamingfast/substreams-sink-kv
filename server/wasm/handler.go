package wasm

import (
	"fmt"
	"github.com/streamingfast/substreams-sink-kv/wasmquery"
	"time"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type Handler struct {
	exportName string
	enginePool *wasmquery.EnginePool
	protoCodec Codec
	logger     *zap.Logger
}

func NewHandler(config *MethodConfig, enginePool *wasmquery.EnginePool, protoCodec Codec, logger *zap.Logger) (*Handler, error) {

	return &Handler{
		exportName: config.ExportName,
		protoCodec: protoCodec,
		enginePool: enginePool,
		logger:     logger.With(zap.String("export_name", config.ExportName)),
	}, nil
}

//func (e *Engine) GetHandler(config *MethodConfig, protoCodec Codec, logger *zap.Logger) (*Handler, error) {
//
//	if _, found := e.functionList[config.ExportName]; !found {
//		return nil, fmt.Errorf("unable to create handler for grpc method %q, export %q not found in wasm", config.FQGRPCName, config.ExportName)
//	}
//
//}

func (h *Handler) handle(_ interface{}, stream grpc.ServerStream) error {
	ctx := stream.Context()

	t0 := time.Now()
	defer func() {
		h.logger.Debug("finished handler", zap.Duration("elapsed", time.Since(t0)))
	}()

	h.logger.Debug("handling wasm query call")
	m := h.protoCodec.NewMessage()
	if err := stream.RecvMsg(m); err != nil {
		return err
	}


	engine := h.enginePool.Borrow(ctx)
	vm, err := engine.Instantiate(h.exportName)
	if err != nil {
		return err
	}


	res, wasmErr, err := vm.Execute(h.exportName, m.Bytes())
	if err != nil {
		return fmt.Errorf("executing func %q: %w", h.exportName, err)
	}
	if err, ok := wasmErr.(string); ok && err != "" {
		return status.Error(codes.Internal, err)
	}

	out := h.protoCodec.NewMessage()
	out.Set(res[0].([]byte))

	if err := stream.SendMsg(out); err != nil {
		return fmt.Errorf("send msg: %w", err)
	}
	return nil
}
