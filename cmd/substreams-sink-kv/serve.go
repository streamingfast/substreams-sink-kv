package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
)

var ServeCmd = Command(serveE,
	`serve <dsn> <spkg>`,
	"Serves the contents of an spkg",
	ExactArgs(2),
	Flags(func(flags *pflag.FlagSet) {
		flags.String("listen-addr", "", "Listen via GRPC Connect-Web on this address")
	}),
	//AfterAllHook(func(_ *cobra.Command) {
	//	sinker.RegisterMetrics()
	//}),
)

func serveE(cmd *cobra.Command, args []string) error {
	return nil
}
