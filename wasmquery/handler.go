package wasmquery

import (
	"context"
	"fmt"
	connect_go "github.com/bufbuild/connect-go"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

type handler struct {
	method *MethodConfig
	engine *Engine
	logger *zap.Logger
}

func newHandler(engine *Engine, method *MethodConfig, logger *zap.Logger) *handler {
	return &handler{
		engine: engine,
		method: method,
		logger: logger,
	}
}

func (h *handler) Handler(ctx context.Context, req *connect_go.Request[ConnectWebRequest]) (*connect_go.Response[ConnectWebRequest], error) {
	logger := logging.Logger(ctx, h.logger).With(zap.String("export_name", h.method.exportName))
	fromRestAPI := req.Msg.contentType == ContentTypeJson
	logger = logger.With(h.method.loggerFields(fromRestAPI)...)
	t0 := time.Now()
	defer func() {
		logger.Info("finished connect-web handler", zap.Duration("elapsed", time.Since(t0)))
	}()

	data := req.Msg.data
	if fromRestAPI {
		dynMsg := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(h.method.inputType)
		if err := dynMsg.UnmarshalJSON(data); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to unmarshal json  for request: %w", err).Error())
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

	exportName := h.method.exportName
	res, wasmErr, err := vmInstance.execute(request, exportName, data)
	if err != nil {
		if vmInstance.panic != nil {
			return nil, status.Error(codes.Internal, vmInstance.panic.Error())
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("unknown error executnig %q: %s", exportName, err))
	}
	if wErr, ok := wasmErr.(string); ok && wErr != "" {
		return nil, status.Error(codes.Internal, wErr)
	}

	respBytes := res[0].([]byte)
	if fromRestAPI {
		dynMsg := dynamic.NewMessageFactoryWithDefaults().NewDynamicMessage(h.method.outputType)
		if err := dynMsg.Unmarshal(respBytes); err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to unmarshal proto for response: %w", err).Error())
		}

		var err error
		respBytes, err = dynMsg.MarshalJSON()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Errorf("failed to get json bytes response: %w", err).Error())
		}

	}

	out := &ConnectWebRequest{data: respBytes, contentType: ContentTypeProto}
	return connect_go.NewResponse(out), nil
}
