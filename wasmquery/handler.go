package wasmquery

import (
	"context"
	"fmt"
	connect_go "github.com/bufbuild/connect-go"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type handler struct {
	msgDesc    *desc.MessageDescriptor
	exportName string
	engine     *Engine
	logger     *zap.Logger
}

func newHandler(engine *Engine, exportName string, msgDesc *desc.MessageDescriptor, logger *zap.Logger) *handler {
	return &handler{
		msgDesc:    msgDesc,
		engine:     engine,
		exportName: exportName,
		logger:     logger,
	}
}

func (h *handler) Handler(ctx context.Context, req *connect_go.Request[ConnectWebRequest]) (*connect_go.Response[ConnectWebRequest], error) {
	logger := logging.Logger(ctx, h.logger)
	t0 := time.Now()
	defer func() {
		logger.Debug("finished connect-web handler", zap.Duration("elapsed", time.Since(t0)))
	}()

	dynMsg := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(h.msgDesc)

	data := req.Msg.data

	if req.Msg.contentType == ContentTypeJson {
		if err := dynMsg.UnmarshalJSON(data); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to unmarshal request: %w", err).Error())
		}

		var err error
		data, err = dynMsg.Marshal()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to get byte data: %w", err).Error())
		}
	}

	vmInstance := h.engine.borrowVM(ctx)
	defer func() {
		h.engine.returnVM(vmInstance)
	}()
	logger = logger.With(vmInstance.loggerFields()...)

	request := &Request{
		ctx:    ctx,
		logger: logger,
	}

	res, wasmErr, err := vmInstance.execute(request, h.exportName, data)
	if err != nil {
		if vmInstance.panic != nil {
			return nil, status.Error(codes.Internal, vmInstance.panic.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("unknown error executnig %q: %s", h.exportName, err))
	}
	if werr, ok := wasmErr.(string); ok && werr != "" {
		return nil, status.Error(codes.Internal, werr)
	}

	out := &ConnectWebRequest{data: res[0].([]byte), contentType: ContentTypeProto}
	return connect_go.NewResponse(out), nil
}
