package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/shutter"
	"github.com/streamingfast/substreams-sink-kv/db"
	"github.com/streamingfast/substreams-sink-kv/sinker"
	"github.com/streamingfast/substreams/client"
	"github.com/streamingfast/substreams/manifest"
	"go.uber.org/zap"
)

var runCmd = &cobra.Command{
	Use:   `run <dsn> <manifest> <module> [<start>:<stop>]`,
	Short: "Fills a KV store from a substreams output and runs a Connect-Web listener",
	Long: Dedent(`
		* dsn: URL to connect to the KV store. Supported schemes: 'badger3', 'badger', 'bigkv', 'tikv', 'netkv'. See https://github.com/streamingfast/kvdb for more details. (ex: 'badger3:///tmp/substreams-sink-kv-db')
		* endpoint: URL to the substreams endpoint (ex: mainnet.eth.streamingfast.io:443)
  		* manifest_path: URL or local path to a 'substreams.yaml' or '.spkg' file (ex: 'https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.3.0/substreams-eth-block-meta-v0.3.0.spkg')
  		* module: FQGRPCName of the output module (declared in the manifest), (ex: 'kv_out')
		* dsn: URL to connect to the KV store. Supported schemes: 'badger3', 'badger', 'bigkv', 'tikv', 'netkv'. See https://github.com/streamingfast/kvdb for more details. (ex: 'badger3:///tmp/substreams-sink-kv-db')
		
		Environment Variables:
		* SUBSTREAMS_API_TOKEN: Your authentication token (JWt) to the substreams endpoint
	`),
	Args: cobra.RangeArgs(1, 4),
	RunE: runE,
}

func init() {
	runCmd.Flags().BoolP("insecure", "k", false, "Skip certificate validation on GRPC connection")
	runCmd.Flags().BoolP("plaintext", "p", false, "Establish GRPC connection in plaintext")
	runCmd.Flags().String("listen-addr", "localhost:8000", "Listen via GRPC Connect-Web on this address")
	runCmd.Flags().Bool("listen-ssl-self-signed", false, "Listen with an HTTPS server (with self-signed certificate)")
	runCmd.Flags().StringP("endpoint", "e", "mainnet.eth.streamingfast.io:443", "URL to the substreams endpoint")

	sinker.RegisterMetrics()
	rootCmd.AddCommand(runCmd)
}

func runE(cmd *cobra.Command, args []string) error {
	app := shutter.New()

	ctx, cancelApp := context.WithCancel(cmd.Context())
	app.OnTerminating(func(_ error) {
		cancelApp()
	})

	dsn, manifestPath, outputModuleName, blockRange := parseArgs(cmd, args)
	endpoint := viper.GetString("run-endpoint")

	zlog.Info("sink to kv",
		zap.String("dsn", dsn),
		zap.String("endpoint", endpoint),
		zap.String("manifest_path", manifestPath),
		zap.String("output_module_name", outputModuleName),
		zap.String("block_range", blockRange),
	)

	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new psql loader: %w", err)
	}

	zlog.Info("reading substreams manifest", zap.String("manifest_path", manifestPath))
	pkg, err := manifest.NewReader(manifestPath).Read()
	if err != nil {
		return fmt.Errorf("read manifest: %w", err)
	}

	graph, err := manifest.NewModuleGraph(pkg.Modules.Modules)
	if err != nil {
		return fmt.Errorf("create substreams moduel graph: %w", err)
	}

	zlog.Info("validating output store", zap.String("output_store", outputModuleName))
	module, err := graph.Module(outputModuleName)
	if err != nil {
		return fmt.Errorf("get output module %q: %w", outputModuleName, err)
	}
	if module.GetKindMap() == nil {
		return fmt.Errorf("ouput module %q is *not* of  type 'Mapper'", outputModuleName)
	}

	if module.Output.Type != "proto:sf.substreams.sink.kv.v1.KVOperations" {
		return fmt.Errorf("kv sync only supports maps with output type 'proto:sf.substreams.sink.kv.v1.KVOperations'")
	}
	hashes := manifest.NewModuleHashes()
	outputModuleHash := hashes.HashModule(pkg.Modules, module, graph)

	resolvedStartBlock, resolvedStopBlock, err := readBlockRange(module, blockRange)
	if err != nil {
		return fmt.Errorf("resolve block range: %w", err)
	}
	zlog.Info("resolved block range",
		zap.Int64("start_block", resolvedStartBlock),
		zap.Uint64("stop_block", resolvedStopBlock),
	)

	apiToken := readAPIToken()
	config := &sinker.Config{
		DBLoader:         kvDB,
		BlockRange:       blockRange,
		Pkg:              pkg,
		OutputModule:     module,
		OutputModuleName: outputModuleName,
		OutputModuleHash: outputModuleHash,
		ClientConfig: client.NewSubstreamsClientConfig(
			endpoint,
			apiToken,
			viper.GetBool("run-insecure"),
			viper.GetBool("run-plaintext"),
		),
	}

	kvSinker, err := sinker.New(
		config,
		zlog,
		tracer,
	)
	if err != nil {
		return fmt.Errorf("unable to setup sinker: %w", err)
	}
	kvSinker.OnTerminating(app.Shutdown)

	app.OnTerminating(func(err error) {
		zlog.Info("application terminating shutting down sinker")
		kvSinker.Shutdown(err)
	})

	go func() {
		if err := kvSinker.Start(ctx); err != nil {
			zlog.Error("sinker failed", zap.Error(err))
			kvSinker.Shutdown(err)
		}
	}()

	if cwListen := viper.GetString("run-listen-addr"); cwListen != "" {
		//go func() {
		//	zlog.Info("starting to listen on", zap.String("addr", cwListen))
		//	standard.ListenConnectWeb(cwListen, kvDB, zlog, viper.GetBool("run-listen-ssl-self-signed"))
		//}()
	} else {
		fmt.Println("oh shit its empty")
	}

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

	zlog.Info("app terminated")
	return nil
}

func parseArgs(cmd *cobra.Command, args []string) (dsn, manifestPath, outputModuleName, blockRange string) {
	switch len(args) {
	case 4:
		blockRange = args[3]
		fallthrough
	case 3:
		dsn = args[0]
		manifestPath = args[1]
		outputModuleName = args[2]
	case 2:
		dsn = args[0]
		manifestPath = args[1]
		outputModuleName = "kv_out"
	case 1:
		dsn = args[0]
		manifestPath = "substreams.yaml"
		outputModuleName = "kv_out"
	default:
	}
	return
}
