package main

import (
	"fmt"
	"github.com/second-state/WasmEdge-go/wasmedge"
	bindgen "github.com/second-state/wasmedge-bindgen/host/go"
	"github.com/spf13/cobra"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dgrpc/server"
	"github.com/streamingfast/dgrpc/server/standard"
	"github.com/streamingfast/substreams-sink-kv/db"
	kvv1 "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	. "github.com/streamingfast/substreams-sink-kv/server"
	"github.com/streamingfast/substreams/manifest"
	pbsubstreams "github.com/streamingfast/substreams/pb/sf/substreams/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
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

type Service struct {
	bg *bindgen.Bindgen
	vm *wasmedge.VM

	kv *db.DB
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

	sinkConfigType := pkg.SinkConfig.TypeUrl
	if sinkConfigType != "sf.substreams.sink.kv.v1.WASMQueryService" && sinkConfigType != "sf.substreams.sink.kv.v1.GenericService" {
		return fmt.Errorf("invalid sink_config type: %s", sinkConfigType)
	}

	zlog.Info("sink to kv",
		zap.String("dsn", dsn),
		zap.String("manifest_path", manifestPath),
	)

	// init db
	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new psql loader: %w", err)
	}

	if sinkConfigType == "sf.substreams.sink.kv.v1.GenericService" {
		go func() {
			zlog.Info("starting to listen on: ", zap.String("addr", listenAddr))
			ListenConnectWeb(listenAddr, kvDB, zlog, mustGetBool(cmd, "run-listen-ssl-self-signed"))
		}()
	} else {
		go func() {
			err := wasmQueryServiceServe(kvDB, pkg, manifestPath)
			if err != nil {
				fmt.Errorf("serving wasm query service: %w", err)
			}
		}()
	}

	// Clean up and wait
	signalHandler := derr.SetupSignalHandler(0 * time.Second)
	zlog.Info("ready, waiting for signal to quit")
	select {
	case <-signalHandler:
		zlog.Info("received termination signal, quitting application")
	}

	return nil
}

func wasmQueryServiceServe(db *db.DB, pkg *pbsubstreams.Package, manifestPath string) error {
	fmt.Println("Started")

	var config kvv1.WASMQueryService
	s := standard.NewServer(server.NewOptions())
	_ = s.GrpcServer()

	fmt.Println("marshaling")
	// Unmarshal the kvConfig
	opts := proto.UnmarshalOptions{}
	err := anypb.UnmarshalTo(pkg.SinkConfig, &config, opts)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("marshaling kv config: %w", err)
	}

	fmt.Println("getting wasm buffer")
	// Get wasm buffer from kvConfig
	wasmedge.SetLogErrorLevel()
	conf := wasmedge.NewConfigure(wasmedge.WASI)
	_ = wasmedge.NewVMWithConfig(conf)

	fmt.Printf("config: %v\n", config)
	fmt.Println("loading wasm")

	// Load vm from wasm buffer
	//err = vm.LoadWasmBuffer(config.WasmQueryModule)
	//if err != nil {
	//	return fmt.Errorf("loading wasm buffer: %w", err)
	//}
	//
	//fmt.Println("loading manifest from file")
	//_, err = manifest.LoadManifestFile(manifestPath)
	//if err != nil {
	//	return fmt.Errorf("loading manifest: %w", err)
	//}

	//protoDefinitions, err := LoadProtobufs(pkg, manifest)
	//if err != nil {
	//	return fmt.Errorf("error loading protobuf: %w", err)
	//}
	//
	//if err := LoadSinkConfigs(pkg, manifest, protoDefinitions); err != nil {
	//	return fmt.Errorf("error parsing sink configuration: %w", err)
	//}

	fmt.Println("Listening on :7878")
	s.Launch(":7878")

	return nil
}
