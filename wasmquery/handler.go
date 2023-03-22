package wasmquery

import (
	"fmt"
	"time"

	"github.com/streamingfast/logging"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type Handler struct {
	exportName string
	enginePool *Engine
	protoCodec Codec
	logger     *zap.Logger
}

func newHandler(config *MethodConfig, enginePool *Engine, protoCodec Codec, logger *zap.Logger) (*Handler, error) {

	return &Handler{
		exportName: config.ExportName,
		protoCodec: protoCodec,
		enginePool: enginePool,
		logger:     logger.With(zap.String("export_name", config.ExportName)),
	}, nil
}

func (h *Handler) handle(_ interface{}, stream grpc.ServerStream) error {
	ctx := stream.Context()

	logger := logging.Logger(ctx, h.logger)

	t0 := time.Now()
	defer func() {
		logger.Debug("finished handler", zap.Duration("elapsed", time.Since(t0)))
	}()

	logger.Debug("handling wasm query call")
	m := h.protoCodec.NewMessage()
	if err := stream.RecvMsg(m); err != nil {
		return err
	}

	vmInstance := h.enginePool.borrowVM(ctx)
	defer func() {
		h.enginePool.returnVM(vmInstance)
	}()
	logger = logger.With(vmInstance.loggerFields()...)

	request := &Request{
		ctx:    ctx,
		logger: logger,
	}

	res, wasmErr, err := vmInstance.execute(request, h.exportName, m.Bytes())
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
