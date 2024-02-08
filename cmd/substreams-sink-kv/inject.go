package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	"inject <endpoint> <dsn> [<manifest>] [<start>:<stop>]",
	"Fills a KV store from a Substreams output and optionally runs a server",
	RangeArgs(2, 4),
	Flags(func(flags *pflag.FlagSet) {
		sink.AddFlagsToSet(flags)

		flags.Int("flush-interval", 100, "When in catch up mode, flush every N blocks")
		flags.String("module", "", "An explicit module to sink, if not provided, expecting the Substreams manifest to defined 'sink' configuration")
		flags.String("server-listen-addr", "", "Launch query server on this address")
		flags.Bool("server-listen-ssl-self-signed", false, "Listen with an HTTPS server (with self-signed certificate)")
		flags.String("server-api-prefix", "", "Launch query server with this API prefix so the URl to query is <server-listen-addr>/<server-api-prefix>")
		flags.Int("query-rows-limit", 5000, "Query rows limit when fetching from database if user specify an unlimited scan or if his limit is above this value")

		flags.String("listen-addr", "", "Launch query server on this address")
		flags.Lookup("listen-addr").Deprecated = "use --server-listen-addr instead"
	}),
	Description(`
		Fills a KV store from a Substreams output and optionally runs a server. Authentication
		with the remote endpoint can be done using the 'SUBSTREAMS_API_TOKEN' environment variable.

		The required arguments are:
		- <endpoint>: The Substreams endpoint to reach (e.g. 'mainnet.eth.streamingfast.io:443').
		- <dsn>: URL to connect to the KV store, see https://github.com/streamingfast/kvdb for more DSN details (e.g. 'badger3:///tmp/substreams-sink-kv-db').

		The optional arguments are:
		- <manifest>: URL or local path to a '.spkg' file (e.g. 'https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg').
		- <start>:<stop>: The range of block to sync, if not provided, will sync from the module's initial block and then forever.

		If the <manifest> is not provided, assume '.' contains a Substreams project to run. If
		<start>:<stop> is not provided, assumes the whole chain.

		Note that if you need to provide a block range, <manifest> needs to be provided. Use '.'
		which is the default when <manifest> is not provided.
	`),
	ExamplePrefixed("substreams-sink-kv inject", `
		# Inject key/values produced by kv_out for the whole chain
		mainnet.eth.streamingfast.io:443 badger3:///tmp/block-meta-db https://github.com/streamingfast/substreams-eth-block-meta/releases/download/v0.4.0/substreams-eth-block-meta-v0.4.0.spkg

		# Inject key/values produced by Substreams in curret directory for range 10 000 to 20 000
		mainnet.eth.streamingfast.io:443 badger3:///tmp/block-meta-db . 10_000:20_000
	`),
	OnCommandErrorLogAndExit(zlog),
)

func injectRunE(cmd *cobra.Command, args []string) error {
	app := shutter.New()
	ctx := cmd.Context()

	sink.RegisterMetrics()
	sinker.RegisterMetrics()

	endpoint, dsn, manifestPath, blockRange := extractInjectArgs(cmd, args)
	queryRowLimit := sflags.MustGetInt(cmd, "query-rows-limit")

	flushInterval := sflags.MustGetUint64(cmd, "flush-interval")
	module := sflags.MustGetString(cmd, "module")

	listenAddr, provided := sflags.MustGetStringProvided(cmd, "server-listen-addr")
	if !provided {
		// Fallback to deprecated flag
		listenAddr = sflags.MustGetString(cmd, "listen-addr")
	}

	apiPrefix := sflags.MustGetString(cmd, "server-api-prefix")
	listenSslSelfSigned := sflags.MustGetBool(cmd, "server-listen-ssl-self-signed")

	fields := []zap.Field{
		zap.String("dsn", dsn),
		zap.String("endpoint", endpoint),
		zap.String("manifest_path", manifestPath),
		zap.String("block_range", blockRange),
		zap.Uint64("flush_interval", flushInterval),
		zap.String("module", module),
	}

	if listenAddr != "" {
		fields = append(fields,
			zap.Bool("start_server", true),
			zap.String("listen_addr", listenAddr),
			zap.Bool("listen_ssl_self_signed", listenSslSelfSigned),
			zap.String("api_prefix", apiPrefix),
		)
	}

	zlog.Info("starting KV sinker", fields...)

	kvDB, err := db.New(dsn, queryRowLimit, zlog, tracer)
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

	kvSinker.OnTerminating(func(err error) {
		if err != nil {
			app.Shutdown(err)
			return
		}

		if listenAddr == "" {
			// If there is no server actively listening, we shut down right away.
			// The actual termination in this case will happen otherwise when the
			// SIGINT signal is received.
			app.Shutdown(nil)
		} else {
			zlog.Info("sinker terminating but server is still running, waiting for SIGINT to terminate")
		}
	})
	app.OnTerminating(func(err error) {
		kvSinker.Shutdown(err)
	})

	go func() {
		kvSinker.Run(ctx)
	}()

	if listenAddr != "" {
		zlog.Info("setting up query server")
		server, err := setupServer(cmd, sink.Package(), kvDB, apiPrefix, listenSslSelfSigned)
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
	if len(args) >= 3 {
		manifestPath = args[2]
	}

	if len(args) == 4 {
		blockRange = args[3]
	}

	return
}
