package wasm

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type Handler struct {
	exportName string
	engine     *Engine
	logger     *zap.Logger
}

func (e *Engine) GetHandler(config *MethodConfig, logger *zap.Logger) (*Handler, error) {

	if _, found := e.functionList[config.ExportName]; !found {
		return nil, fmt.Errorf("unable to create handler for grpc method %q, export %q not found in wasm", config.FQGRPCName, config.ExportName)
	}

	return &Handler{
		exportName: config.ExportName,
		engine:     e,
		logger:     logger.With(zap.String("export_name", config.ExportName)),
	}, nil
}

func (h *Handler) handle(_ interface{}, stream grpc.ServerStream) error {
	t0 := time.Now()
	defer func() {
		h.logger.Debug("finished handler", zap.Duration("elapsed", time.Since(t0)))
	}()

	m := NewPassthroughBytes()
	if err := stream.RecvMsg(m); err != nil {
		return err
	}

	res, _, err := h.engine.bg.Execute(h.exportName, m.Bytes)
	if err != nil {
		return fmt.Errorf("executing func %q: %w", h.exportName, err)
	}

	out := NewPassthroughBytes()
	out.Set(res[0].([]byte))

	if err := stream.SendMsg(out); err != nil {
		return fmt.Errorf("send msg: %w", err)
	}
	return nil
}
