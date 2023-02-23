package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
)

var InjectCmd = Command(injectE,
	`inject <dsn> <substreams-endpoint> <spkg>`,
	"Injects the contents of an spkg",
	ExactArgs(3),
	Flags(func(flags *pflag.FlagSet) {
		flags.String("listen-addr", "", "Listen via GRPC Connect-Web on this address")
	}),
	//AfterAllHook(func(_ *cobra.Command) {
	//	sinker.RegisterMetrics()
	//}),
)

func injectE(cmd *cobra.Command, args []string) error {
	return nil
}
