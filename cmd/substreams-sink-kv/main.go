package main

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
	_ "github.com/streamingfast/kvdb/store/badger"
	_ "github.com/streamingfast/kvdb/store/badger3"
	_ "github.com/streamingfast/kvdb/store/bigkv"
	_ "github.com/streamingfast/kvdb/store/netkv"
	_ "github.com/streamingfast/kvdb/store/tikv"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
)

// Commit sha1 value, injected via go build `ldflags` at build time
var commit = ""

// Version value, injected via go build `ldflags` at build time
var version = "dev"

// Date value, injected via go build `ldflags` at build time
var date = ""

var zlog, tracer = logging.RootLogger("sink-kv", "github.com/streamingfast/substreams-sink-kv/cmd/substreams-sink-kv")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zap.InfoLevel))
}

func main() {
	Run("substreams-sink-kv", "Substreams KV Sink",
		ConfigureViper("SINK"),
		ConfigureVersion(),

		injectCmd,
		serveCmd,

		PersistentFlags(
			func(flags *pflag.FlagSet) {
				flags.Duration("delay-before-start", 0, "[OPERATOR] Amount of time to wait before starting any internal processes, can be used to perform to maintenance on the pod before actually letting it starts")
				flags.String("metrics-listen-addr", "localhost:9102", "[OPERATOR] If non-empty, the process will listen on this address for Prometheus metrics request(s)")
				flags.String("pprof-listen-addr", "localhost:6060", "[OPERATOR] If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")
			},
		),
		AfterAllHook(func(cmd *cobra.Command) {
			cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
				if err := setupCmd(cmd); err != nil {
					return err
				}
				return nil
			}
		}),
	)
}

func ConfigureVersion() CommandOption {
	return CommandOptionFunc(func(cmd *cobra.Command) {
		cmd.Version = versionString(version)
	})
}

func versionString(version string) string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		panic("we should have been able to retrieve info from 'runtime/debug#ReadBuildInfo'")
	}

	commit := findSetting("vcs.revision", info.Settings)
	date := findSetting("vcs.time", info.Settings)

	var labels []string
	if len(commit) >= 7 {
		labels = append(labels, fmt.Sprintf("Commit %s", commit[0:7]))
	}

	if date != "" {
		labels = append(labels, fmt.Sprintf("Built %s", date))
	}

	if len(labels) == 0 {
		return version
	}

	return fmt.Sprintf("%s (%s)", version, strings.Join(labels, ", "))
}

func findSetting(key string, settings []debug.BuildSetting) (value string) {
	for _, setting := range settings {
		if setting.Key == key {
			return setting.Value
		}
	}

	return ""
}
