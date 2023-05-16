package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/streamingfast/cli"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/cli/sflags"
	"github.com/streamingfast/shutter"
	sink "github.com/streamingfast/substreams-sink"
	"github.com/streamingfast/substreams-sink-kv/db"
	"github.com/streamingfast/substreams-sink-kv/sinker"
	"go.uber.org/zap"
)

var injectCmd = Command(injectRunE,
	"inject <endpoint> <dsn> <manifest> [<start>:<stop>]",
	"Fills a KV store from a Substreams output and optionally runs a server",
	RangeArgs(3, 4),
	Flags(func(flags *pflag.FlagSet) {
		sink.AddFlagsToSet(flags)

		flags.Int("flush-interval", 1000, "When in catch up mode, flush every N blocks")
		flags.String("listen-addr", "", "Launch query server on this address")
		flags.String("module", "", "An explicit module to sink, if not provided, expecting the Substreams manifest to defined 'sink' configuration")
	}),
	Description(`
		Fills a KV store from a Substreams output and optionally runs a server. Authentication
		with the remote endpoint can be done using the 'SUBSTREAMS_API_TOKEN' environment variable.

		The required arguments are:
		- <endpoint>: The Substreams endpoint to reach (e.g. 'mainnet.eth.streamingfast.io:443').
		- <dsn>: URL to connect to the KV store, see https://github.com/streamingfast/kvdb for more DSN details (e.g. 'badger3:///tmp/substreams-sink-kv-db').
		- <manifest>: URL or local path to a '.spkg' file (e.g. 'https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg').

		The optional arguments are:
		- <start>:<stop>: The range of block to sync, if not provided, will sync from the module's initial block and then forever.
	`),
	ExamplePrefixed("substreams-sink-kv inject", `
		# Inject key/values produced by kv_out for the whole chain
		mainnet.eth.streamingfast.io:443 badger3:///tmp/block-meta-db https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg
	`),
	OnCommandErrorLogAndExit(zlog),
)

func injectRunE(cmd *cobra.Command, args []string) error {
	app := shutter.New()
	ctx := cmd.Context()

	sink.RegisterMetrics()
	sinker.RegisterMetrics()

	endpoint, dsn, manifestPath, blockRange := extractInjectArgs(cmd, args)

	flushInterval := sflags.MustGetDuration(cmd, "flush-interval")
	module := sflags.MustGetString(cmd, "module")

	zlog.Info("starting KV sinker",
		zap.String("dsn", dsn),
		zap.String("endpoint", endpoint),
		zap.String("manifest_path", manifestPath),
		zap.String("block_range", blockRange),
		zap.Duration("flush_interval", flushInterval),
		zap.String("module", module),
	)

	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new psql loader: %w", err)
	}

	outputModuleName := sink.InferOutputModuleFromPackage
	if module != "" {
		outputModuleName = module
	}

	sink, err := sink.NewFromViper(
		cmd,
		"sf.substreams.sink.kv.v1.KVOperations",
		endpoint, manifestPath, outputModuleName, blockRange,
		zlog, tracer,
	)
	if err != nil {
		return fmt.Errorf("unable to setup sinker: %w", err)
	}

	kvSinker, err := sinker.New(sink, kvDB, flushInterval, zlog, tracer)
	if err != nil {
		return fmt.Errorf("unable to setup sinker: %w", err)
	}

	kvSinker.OnTerminating(app.Shutdown)
	app.OnTerminating(func(err error) {
		kvSinker.Shutdown(err)
	})

	go func() {
		kvSinker.Run(ctx)
	}()

	if listenAddr := viper.GetString("inject-listen-addr"); listenAddr != "" {
		zlog.Info("setting up query server", zap.String("dsn", dsn), zap.String("listen_addr", listenAddr))
		server, err := setupServer(cmd, sink.Package(), kvDB)
		if err != nil {
			return fmt.Errorf("setup server: %w", err)

		}
		app.OnTerminating(func(_ error) {
			zlog.Info("inject terminating shutting down server")
			server.Shutdown()
		})

		go func() {
			if err := server.Serve(listenAddr); err != nil {
				app.Shutdown(err)
			}
		}()
	}

	zlog.Info("ready, waiting for signal to quit")

	signalHandler, isSignaled, _ := cli.SetupSignalHandler(0*time.Second, zlog)
	select {
	case <-signalHandler:
		go app.Shutdown(nil)
		break
	case <-app.Terminating():
		zlog.Info("inject terminating", zap.Bool("from_signal", isSignaled.Load()), zap.Bool("with_error", app.Err() != nil))
		break
	}

	zlog.Info("inject for run termination")
	select {
	case <-app.Terminated():
	case <-time.After(30 * time.Second):
		zlog.Warn("inject did not terminate within 30s")
	}

	if err := app.Err(); err != nil {
		return err
	}

	zlog.Info("inject terminated gracefully")
	return nil
}

func extractInjectArgs(_ *cobra.Command, args []string) (endpoint, dsn, manifestPath, blockRange string) {
	endpoint = args[0]
	dsn = args[1]
	manifestPath = args[2]
	if len(args) == 4 {
		blockRange = args[3]
	}

	return
}
