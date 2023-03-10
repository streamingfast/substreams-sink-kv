package main

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"

	pbsubstreams "github.com/streamingfast/substreams/pb/sf/substreams/v1"
	"google.golang.org/protobuf/types/descriptorpb"

	"time"

	"github.com/spf13/cobra"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/shutter"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"github.com/streamingfast/substreams-sink-kv/server"
	"github.com/streamingfast/substreams-sink-kv/server/standard"
	"github.com/streamingfast/substreams-sink-kv/server/wasm"
	"github.com/streamingfast/substreams/manifest"
	"go.uber.org/zap"
)

var serveCmd = Command(serveRunE,
	"serve <dsn> <spkg>",
	"Launches a query server to consume connected to a sinkd kv store",
	ExactArgs(2),
	Flags(func(flags *pflag.FlagSet) {
		flags.String("listen-addr", ":7878", "Listen via GRPC Connect-Web on this address")
		flags.Bool("listen-ssl-self-signed", false, "Listen with an HTTPS server (with self-signed certificate)")
	}),
	Description(`
		* dsn: URL to connect to the KV store. Supported schemes: 'badger3', 'badger', 'bigkv', 'tikv', 'netkv'. See https://github.com/streamingfast/kvdb for more details. (ex: 'badger3:///tmp/substreams-sink-kv-db')
 		* spkg: URL or local path to a '.spkg' file (ex: 'https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.3.0/substreams-eth-block-meta-v0.3.0.spkg')
	`),
)

func serveRunE(cmd *cobra.Command, args []string) error {
	app := shutter.New()

	ctx, cancelApp := context.WithCancel(cmd.Context())
	app.OnTerminating(func(_ error) {
		cancelApp()
	})

	// parse args
	dsn := args[0]
	manifestPath := args[1]
	listenAddr := mustGetString(cmd, "listen-addr")
	manifestReader := manifest.NewReader(manifestPath)

	zlog.Info("reading substreams spkg", zap.String("manifest_path", manifestPath))
	pkg, err := manifestReader.Read()
	if err != nil {
		return fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}

	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new kvdb: %w", err)
	}

	zlog.Info("setting up query server",
		zap.String("dsn", dsn),
		zap.String("listen_addr", listenAddr),
	)
	server, err := setupServer(cmd, pkg, kvDB)
	if err != nil {
		return fmt.Errorf("setup server: %w", err)

	}
	app.OnTerminating(func(err error) {
		zlog.Info("application terminating shutting down server")
		server.Shutdown()
	})

	go func() {
		if err := server.Serve(listenAddr); err != nil {
			app.Shutdown(err)
		}
	}()

	// Clean up and wait
	signalHandler := derr.SetupSignalHandler(0 * time.Second)
	zlog.Info("ready, waiting for signal to quit")
	select {
	case <-signalHandler:
		zlog.Info("received termination signal, quitting application")
		go app.Shutdown(nil)
	case <-app.Terminating():
		NoError(app.Err(), "application shutdown unexpectedly, quitting")
	}

	zlog.Info("waiting for app termination")
	select {
	case <-app.Terminated():
	case <-ctx.Done():
	case <-time.After(30 * time.Second):
		zlog.Error("application did not terminated within 30s, forcing exit")
	}

	return nil
}

func setupServer(cmd *cobra.Command, pkg *pbsubstreams.Package, kvDB *db.DB) (server.Serveable, error) {
	if pkg.SinkConfig == nil {
		return nil, fmt.Errorf("no sink config found in spkg")
	}
	switch pkg.SinkConfig.TypeUrl {
	case "sf.substreams.sink.kv.v1.WASMQueryService":

		wasmServ := &pbkv.WASMQueryService{}
		if err := pkg.SinkConfig.UnmarshalTo(wasmServ); err != nil {
			return nil, fmt.Errorf("failed to proto unmarshall: %w", err)
		}
		fileDesc, err := findProtoDef(pkg, wasmServ.GrpcService)
		if err != nil {
			return nil, fmt.Errorf("find proto file descriptor: %w", err)
		}

		wasmEngine, err := wasm.NewEngineFromBytes(wasmServ.GetWasmQueryModule(), kvDB, zlog)
		if err != nil {
			return nil, fmt.Errorf("failed to setup wasm engine: %w", err)
		}

		return wasm.NewServer(wasm.NewConfig(fileDesc), wasmEngine, zlog)
	case "sf.substreams.sink.kv.v1.GenericService":
		return standard.NewServer(kvDB, zlog, mustGetBool(cmd, "run-listen-ssl-self-signed")), nil
	default:
		return nil, fmt.Errorf("invalid sink_config type: %s", pkg.SinkConfig.TypeUrl)
	}
}

func findProtoDef(pkg *pbsubstreams.Package, fqGrpcService string) (*descriptorpb.FileDescriptorProto, error) {
	for _, f := range pkg.ProtoFiles {
		if f.GetPackage() == fqGrpcService {
			return f, nil
		}
	}

	return nil, fmt.Errorf("unable to find proto file descriptor with pakage %q in spkg", fqGrpcService)
}
