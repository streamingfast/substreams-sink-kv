package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/dgrpc/server"
	"github.com/streamingfast/dgrpc/server/standard"
)

var ServeCmd = Command(serveE,
	`serve <dsn> <spkg>`,
	"Serves the contents of an spkg",
	ExactArgs(2),
	Flags(func(flags *pflag.FlagSet) {
		flags.String("listen-addr", "", "Listen via GRPC Connect-Web on this address")
	}),
)

func serveE(cmd *cobra.Command, args []string) error {
	s := standard.NewServer(server.NewOptions())
	s.Launch(":7878")

	//packagePath := args[1]
	//
	//zlog.Info("reading substreams manifest", zap.String("manifest_path", packagePath))
	//pkg, err := manifest.NewReader(packagePath).Read()
	//if err != nil {
	//	return fmt.Errorf("read package: %w", err)
	//}
	//
	//targetNetwork := pkg.TargetNetwork
	//
	//s := standard.NewServer()
	return nil
}
