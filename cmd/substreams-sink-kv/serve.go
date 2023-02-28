package main

import (
	"fmt"
	"github.com/spf13/cobra"
	. "github.com/streamingfast/cli"
	"github.com/streamingfast/substreams/manifest"
)

var serveCmd = &cobra.Command{
	Use:   `serve <spkg>`,
	Short: "Serves the contents of an spkg",
	Args:  cobra.ExactArgs(1),
	RunE:  serveE,
}

func init() {
	serveCmd.Flags().String("listen-addr", "", "Listen via GRPC Connect-Web on this address")

	rootCmd.AddCommand(serveCmd)
}

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
