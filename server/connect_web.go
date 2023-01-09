package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	connect_go "github.com/bufbuild/connect-go"
	grpcreflect "github.com/bufbuild/connect-grpcreflect-go"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"github.com/streamingfast/kvdb/store"
	"github.com/streamingfast/substreams-sink-kv/db"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	kvconnect "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1/kvv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func ListenConnectWeb(addr string, dbReader db.DBReader, logger *zap.Logger) error {
	cs := &ConnectServer{
		DBReader: dbReader,
		logger:   logger,
	}

	reflector := grpcreflect.NewStaticReflector(
		"substreams.sink.kv.v1.Kv",
	)

	mux := http.NewServeMux()
	// The generated constructors return a path and a plain net/http
	// handler.
	mux.Handle(kvconnect.NewKvHandler(cs))
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	return http.ListenAndServe(
		addr,
		// For gRPC clients, it's convenient to support HTTP/2 without TLS. You can
		// avoid x/net/http2 by using http.ListenAndServeTLS.
		h2c.NewHandler(
			newCORS().Handler(mux),
			&http2.Server{}),
	)
}

type ConnectServer struct {
	kvconnect.UnimplementedKvHandler
	DBReader db.DBReader
	logger   *zap.Logger
}

func (cs *ConnectServer) Get(ctx context.Context, req *connect_go.Request[kvv1.GetRequest]) (*connect_go.Response[kvv1.GetResponse], error) {
	val, err := cs.DBReader.Get(ctx, req.Msg.Key)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			cs.logger.Debug("key not found", zap.String("key", req.Msg.Key), zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("requested key not found in database: %w", err))
		}
		cs.logger.Info("internal error", zap.String("key", req.Msg.Key), zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	resp := connect.NewResponse(&kvv1.GetResponse{
		Value: val,
	})
	return resp, nil
}

func (cs *ConnectServer) GetMany(ctx context.Context, req *connect_go.Request[kvv1.GetManyRequest]) (*connect_go.Response[kvv1.GetManyResponse], error) {
	vals, err := cs.DBReader.GetMany(ctx, req.Msg.Keys)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			cs.logger.Debug("key not found", zap.Strings("keys", req.Msg.Keys), zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("one of the requested keys was not found in database: %w", err))
		}
		cs.logger.Info("internal error", zap.Strings("keys", req.Msg.Keys), zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	resp := connect.NewResponse(&kvv1.GetManyResponse{
		Values: vals,
	})
	return resp, nil
}

func (cs *ConnectServer) GetByPrefix(ctx context.Context, req *connect_go.Request[kvv1.GetByPrefixRequest]) (*connect_go.Response[kvv1.GetByPrefixResponse], error) {
	keyVals, limitReached, err := cs.DBReader.GetByPrefix(ctx, req.Msg.Prefix, int(req.Msg.Limit))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			cs.logger.Debug("prefix not found", zap.String("prefix", req.Msg.Prefix), zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeNotFound, fmt.Errorf("one of the requested keys was not found in database: %w", err))
		}
		if errors.Is(err, db.ErrInvalidArguments) {
			cs.logger.Debug("invalid arguments", zap.String("prefix", req.Msg.Prefix), zap.Uint64("limit", req.Msg.Limit), zap.Error(err))
			return nil, connect_go.NewError(connect_go.CodeInvalidArgument, err)

		}
		cs.logger.Info("internal error", zap.String("prefix", req.Msg.Prefix), zap.Error(err))
		return nil, connect_go.NewError(connect_go.CodeInternal, errors.New("internal server error"))
	}
	protoKeyVals := make([]*kvv1.KV, len(keyVals))
	for i := range keyVals {
		protoKeyVals[i] = &kvv1.KV{
			Key:   keyVals[i].Key,
			Value: keyVals[i].Value,
		}
	}
	resp := connect.NewResponse(&kvv1.GetByPrefixResponse{
		KeyValues:    protoKeyVals,
		LimitReached: limitReached,
	})
	return resp, nil
}

func (cs *ConnectServer) Scan(context.Context, *connect_go.Request[kvv1.ScanRequest]) (*connect_go.Response[kvv1.ScanResponse], error) {
	return nil, connect_go.NewError(connect_go.CodeUnimplemented, errors.New("substreams.sink.kv.v1.Kv.Scan is not implemented"))
}

func newCORS() *cors.Cors {
	// To let web developers play with the demo service from browsers, we need a
	// very permissive CORS setup.
	return cors.New(cors.Options{
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowOriginFunc: func(origin string) bool {
			// Allow all origins, which effectively disables CORS.
			return true
		},
		AllowedHeaders: []string{"*"},
		ExposedHeaders: []string{
			// Content-Type is in the default safelist.
			"Accept",
			"Accept-Encoding",
			"Accept-Post",
			"Connect-Accept-Encoding",
			"Connect-Content-Encoding",
			"Content-Encoding",
			"Grpc-Accept-Encoding",
			"Grpc-Encoding",
			"Grpc-Message",
			"Grpc-Status",
			"Grpc-Status-Details-Bin",
		},
		// Let browsers cache CORS information for longer, which reduces the number
		// of preflight requests. Any changes to ExposedHeaders won't take effect
		// until the cached data expires. FF caps this value at 24h, and modern
		// Chrome caps it at 2h.
		MaxAge: int(2 * time.Hour / time.Second),
	})
}
