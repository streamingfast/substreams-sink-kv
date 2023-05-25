package standard

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	pbkv "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1"

	"github.com/bufbuild/connect-go"
	connect_go "github.com/bufbuild/connect-go"
	"github.com/streamingfast/dgrpc/server"
	connectweb "github.com/streamingfast/dgrpc/server/connect-web"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbkvconnect "github.com/streamingfast/substreams-sink-kv/pb/sf/substreams/sink/kv/v1/pbkvconnect"
	sserver "github.com/streamingfast/substreams-sink-kv/server"
	"go.uber.org/zap"
)

var _ sserver.Serveable = (*ConnectServer)(nil)

func NewServer(dbReader db.Reader, logger *zap.Logger, encrypted bool) *ConnectServer {
	cs := &ConnectServer{
		DBReader: dbReader,
		logger:   logger,
	}

	location, handler := pbkvconnect.NewKvHandler(cs)
	mappings := map[string]http.Handler{
		location: handler,
	}

	opts := []server.Option{
		server.WithReflection("sf.substreams.sink.kv.v1.Kv"),
		server.WithLogger(logger),
		server.WithPermissiveCORS(),
		server.WithHealthCheck(server.HealthCheckOverHTTP, func(_ context.Context) (isReady bool, out interface{}, err error) { return true, nil, nil }),
	}

	if encrypted {
		opts = append(opts, server.WithInsecureServer())
	} else {
		opts = append(opts, server.WithPlainTextServer())
	}
	cs.srv = connectweb.New(mappings, opts...)
	return cs
}

type ConnectServer struct {
	pbkvconnect.UnimplementedKvHandler
	srv      *connectweb.ConnectWebServer
	DBReader db.Reader
	logger   *zap.Logger
}

func (cs *ConnectServer) Shutdown() {
	cs.logger.Info("connect server received shutdown, shutting down server")
	cs.srv.Shutdown(nil)
}

func (cs *ConnectServer) Serve(listenAddr string) error {
	go cs.srv.Launch(listenAddr)
	<-cs.srv.Terminated()

	return cs.srv.Err()
}

func (cs *ConnectServer) Get(ctx context.Context, req *connect_go.Request[pbkv.GetRequest]) (*connect_go.Response[pbkv.GetResponse], error) {
	logger := cs.logger.With(zap.String("key", req.Msg.Key))
	val, err := cs.DBReader.Get(ctx, req.Msg.Key)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			logger.Debug("key not found", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("requested key not found in database: %w", err))
		}
		logger.Info("internal error", zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	resp := connect.NewResponse(&pbkv.GetResponse{
		Value: val,
	})
	return resp, nil
}

func (cs *ConnectServer) GetMany(ctx context.Context, req *connect_go.Request[pbkv.GetManyRequest]) (*connect_go.Response[pbkv.GetManyResponse], error) {
	logger := cs.logger.With(zap.Strings("keys", req.Msg.Keys))
	vals, err := cs.DBReader.GetMany(ctx, req.Msg.Keys)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			logger.Debug("key not found", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("one of the requested keys was not found in database: %w", err))
		}
		if errors.Is(err, db.ErrInvalidArguments) {
			logger.Debug("invalid arguments", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeInvalidArgument, err)
		}
		logger.Info("internal error", zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	resp := connect.NewResponse(&pbkv.GetManyResponse{
		Values: vals,
	})
	return resp, nil
}

func (cs *ConnectServer) GetByPrefix(ctx context.Context, req *connect_go.Request[pbkv.GetByPrefixRequest]) (*connect_go.Response[pbkv.GetByPrefixResponse], error) {
	logger := cs.logger.With(zap.String("prefix", req.Msg.Prefix), zap.Uint64("limit", req.Msg.Limit))
	keyVals, limitReached, err := cs.DBReader.GetByPrefix(ctx, req.Msg.Prefix, int(req.Msg.Limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			logger.Debug("prefix not found", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("one of the requested keys was not found in database: %w", err))
		}
		if errors.Is(err, db.ErrInvalidArguments) {
			logger.Debug("invalid arguments", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeInvalidArgument, err)
		}
		logger.Info("internal error", zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	protoKeyVals := make([]*pbkv.KV, len(keyVals))
	for i := range keyVals {
		protoKeyVals[i] = &pbkv.KV{
			Key:   keyVals[i].Key,
			Value: keyVals[i].Value,
		}
	}
	resp := connect.NewResponse(&pbkv.GetByPrefixResponse{
		KeyValues:    protoKeyVals,
		LimitReached: limitReached,
	})
	return resp, nil
}

func (cs *ConnectServer) Scan(ctx context.Context, req *connect_go.Request[pbkv.ScanRequest]) (*connect_go.Response[pbkv.ScanResponse], error) {
	logger := cs.logger.With(zap.String("begin", req.Msg.Begin), zap.Uint64("limit", req.Msg.Limit))
	exclusiveEnd := ""
	if req.Msg.ExclusiveEnd != nil {
		exclusiveEnd = *req.Msg.ExclusiveEnd
	}
	keyVals, limitReached, err := cs.DBReader.Scan(ctx, req.Msg.Begin, exclusiveEnd, int(req.Msg.Limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			logger.Debug("no values found", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("one of the requested keys was not found in database: %w", err))
		}
		if errors.Is(err, db.ErrInvalidArguments) {
			logger.Debug("invalid arguments", zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeInvalidArgument, err)
		}
		logger.Info("internal error", zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	protoKeyVals := make([]*pbkv.KV, len(keyVals))
	for i := range keyVals {
		protoKeyVals[i] = &pbkv.KV{
			Key:   keyVals[i].Key,
			Value: keyVals[i].Value,
		}
	}
	resp := connect.NewResponse(&pbkv.ScanResponse{
		KeyValues:    protoKeyVals,
		LimitReached: limitReached,
	})
	return resp, nil
}
