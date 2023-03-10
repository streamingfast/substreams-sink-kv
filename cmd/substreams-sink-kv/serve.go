package main

import (
	"fmt"

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

// in the proto file that is given to us, the sinkConfig > grpcService will point to a proto file contained within the "REPEATED"
// I have to go into that file, a service will be defined, and from that service there are a few different rpc messages that we will "serve"
// I have to serve these rpc messages.

var serveCmd = &cobra.Command{
	Use:   `serve <dsn> <spkg>`,
	Short: "Serves the contents of an spkg",
	Args:  cobra.ExactArgs(2),
	RunE:  serveE,
}

func init() {
	serveCmd.Flags().String("listen-addr", ":7878", "Listen via GRPC Connect-Web on this address")
	serveCmd.Flags().Bool("listen-ssl-self-signed", false, "Listen with an HTTPS server (with self-signed certificate)")

	rootCmd.AddCommand(serveCmd)
}

func mustGetString(cmd *cobra.Command, flagName string) string {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func mustGetBool(cmd *cobra.Command, flagName string) bool {
	val, err := cmd.Flags().GetBool(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func serveE(cmd *cobra.Command, args []string) error {
	app := shutter.New()

	// parse args
	dsn := args[0]
	manifestPath := args[1]
	listenAddr := mustGetString(cmd, "listen-addr")
	manifestReader := manifest.NewReader(manifestPath)

	// get config type and check it matches either option
	pkg, err := manifestReader.Read()
	if err != nil {
		return fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}
	if pkg.SinkConfig == nil {
		return fmt.Errorf("no sink config found in spkg")
	}

	// init db
	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new psql loader: %w", err)
	}

	zlog.Info("sink to kv",
		zap.String("dsn", dsn),
		zap.String("manifest_path", manifestPath),
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

	//if sinkConfigType == "sf.substreams.sink.kv.v1.GenericService" {
	//	go func() {
	//		zlog.Info("starting to listen on: ", zap.String("addr", listenAddr))
	//		ListenConnectWeb(listenAddr, kvDB, zlog, mustGetBool(cmd, "run-listen-ssl-self-signed"))
	//	}()
	//} else {
	//	go func() {
	//		err := wasmQueryServiceServe(kvDB, pkg, manifestPath)
	//		if err != nil {
	//			fmt.Errorf("serving wasm query service: %w", err)
	//		}
	//	}()
	//}

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
	case <-time.After(30 * time.Second):
		zlog.Error("application did not terminated within 30s, forcing exit")
	}

	return nil
}

func setupServer(cmd *cobra.Command, pkg *pbsubstreams.Package, kvDB *db.DB) (server.Serveable, error) {
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
