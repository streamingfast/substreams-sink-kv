package main

import (
	"net/http"
	"time"

	"github.com/streamingfast/logging"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/streamingfast/dmetrics"
	"go.uber.org/zap"
)

func setupCmd(cmd *cobra.Command) error {
	cmd.SilenceUsage = true

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
	setupLogger(viper.GetString("global-log-format"))
	return nil
}

func setupLogger(logFormat string) {
	options := []logging.InstantiateOption{
		logging.WithSwitcherServerAutoStart(),
		logging.WithDefaultLevel(zap.InfoLevel),
		logging.WithConsoleToStdout(),
	}
	options = append(options, logging.WithConsoleToStderr())
	if logFormat == "stackdriver" || logFormat == "json" {
		options = append(options, logging.WithProductionLogger())
	}

	logging.InstantiateLoggers(options...)
}
