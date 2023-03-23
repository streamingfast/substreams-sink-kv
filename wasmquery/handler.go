package wasmquery

import (
	"context"
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

func (h *Handler) handle(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {

	logger := logging.Logger(ctx, h.logger)

	t0 := time.Now()
	defer func() {
		logger.Debug("finished handler", zap.Duration("elapsed", time.Since(t0)))
	}()

	logger.Debug("handling wasm query call")
	m := h.protoCodec.NewMessage()
	if err := dec(m); err != nil {
		return nil, err
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
		if vmInstance.panic != nil {
			return nil, status.Error(codes.Internal, vmInstance.panic.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("unknown error executnig %q: %s", h.exportName, err))
	}
	if werr, ok := wasmErr.(string); ok && werr != "" {
		return nil, status.Error(codes.Internal, werr)
	}

	out := h.protoCodec.NewMessage()
	out.Set(res[0].([]byte))

	return out, nil
}
