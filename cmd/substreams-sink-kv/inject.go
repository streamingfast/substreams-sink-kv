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
	"inject <dsn> <manifest> [<start>:<stop>]",
	"Fills a KV store from a Substreams output and optionally runs a server",
	RangeArgs(1, 3),
	Flags(func(flags *pflag.FlagSet) {
		sink.AddFlagsToSet(flags)
		flags.Int("flush-interval", 1000, "When in catch up mode, flush every N blocks")

		flags.String("endpoint", "mainnet.eth.stramingfast.io:443", "The endpoint where to contact the Substreams server")
		flags.String("listen-addr", "", "Launch query server on this address")
	}),
	Description(`
		* dsn: URL to connect to the KV store. Supported schemes: 'badger3', 'badger', 'bigkv', 'tikv', 'netkv'. See https://github.com/streamingfast/kvdb for more details. (ex: 'badger3:///tmp/substreams-sink-kv-db')
  		* spkg: URL or local path to a '.spkg' file (ex: 'https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.3.0/substreams-eth-block-meta-v0.3.0.spkg')
  		* module: FQGRPCName of the output module (declared in the manifest), (ex: 'kv_out')

		Environment Variables:
		* SUBSTREAMS_API_TOKEN: Your authentication token (JWt) to the substreams endpoint
	`),
	OnCommandErrorLogAndExit(zlog),
)

func injectRunE(cmd *cobra.Command, args []string) error {
	app := shutter.New()
	ctx := cmd.Context()

	sink.RegisterMetrics()
	sinker.RegisterMetrics()

	dsn, manifestPath, blockRange := extractInjectArgs(cmd, args)

	flushInterval := sflags.MustGetDuration(cmd, "flush-interval")
	endpoint := sflags.MustGetString(cmd, "endpoint")

	zlog.Info("starting KV sinker",
		zap.String("dsn", dsn),
		zap.String("endpoint", endpoint),
		zap.String("manifest_path", manifestPath),
		zap.String("block_range", blockRange),
		zap.Duration("flush_interval", flushInterval),
	)

	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new psql loader: %w", err)
	}

	sink, err := sink.NewFromViper(
		cmd,
		"proto:sf.substreams.sink.kv.v1.KVOperations",
		endpoint, manifestPath, sink.InferOutputModuleFromPackage, blockRange,
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

func extractInjectArgs(_ *cobra.Command, args []string) (dsn, manifestPath, blockRange string) {
	manifestPath = "substreams.yaml"

	dsn = args[0]
	if len(args) >= 2 {
		manifestPath = args[1]
	}
	if len(args) >= 3 {
		blockRange = args[2]
	}

	return
}
