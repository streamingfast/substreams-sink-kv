package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

var version = "dev"
var zlog, tracer = logging.RootLogger("sink-kv", "github.com/streamingfast/substreams-sink-kv/cmd/substreams-sink-kv")

var rootCmd = &cobra.Command{
	Use:          "substreams-sink-kv",
	Short:        "Substreams KV Sink",
	SilenceUsage: true,
}

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zap.InfoLevel))
	ConfigureViper("SINK")
	ConfigureVersion()

	rootCmd.PersistentFlags().Duration("delay-before-start", 0, "[OPERATOR] Amount of time to wait before starting any internal processes, can be used to perform to maintenance on the pod before actually letting it starts")
	rootCmd.PersistentFlags().String("metrics-listen-addr", "localhost:9102", "[OPERATOR] If non-empty, the process will listen on this address for Prometheus metrics request(s)")
	rootCmd.PersistentFlags().String("pprof-listen-addr", "localhost:6060", "[OPERATOR] If non-empty, the process will listen on this address for pprof analysis (see https://golang.org/pkg/net/http/pprof/)")

	rootCmd.PersistentPreRun = func(_ *cobra.Command, _ []string) {
		delay := viper.GetDuration("global-delay-before-start")
		if delay > 0 {
			zlog.Info("sleeping to respect delay before start setting", zap.Duration("delay", delay))
			time.Sleep(delay)
		}

		if v := viper.GetString("global-metrics-listen-addr"); v != "" {
			zlog.Info("starting prometheus metrics server", zap.String("listen_addr", v))
			go dmetrics.Serve(v)
		}

		if v := viper.GetString("global-pprof-listen-addr"); v != "" {
			go func() {
				zlog.Info("starting pprof server", zap.String("listen_addr", v))
				err := http.ListenAndServe(v, nil)
				if err != nil {
					zlog.Debug("unable to start profiling server", zap.Error(err), zap.String("listen_addr", v))
				}
			}()
		}
	}
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
