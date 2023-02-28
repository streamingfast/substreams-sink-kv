package main

import (
	"github.com/spf13/cobra"
)

var injectCmd = &cobra.Command{
	Use:          `inject <dsn> <substreams-endpoint> <spkg>`,
	Short:        "Injects the contents of an spkg",
	Args:         cobra.ExactArgs(3),
	RunE:         injectE,
	SilenceUsage: true,
}

func init() {
	injectCmd.Flags().String("listen-addr", "", "Listen via GRPC Connect-Web on this address")

	rootCmd.AddCommand(injectCmd)
}

func injectE(cmd *cobra.Command, args []string) error {
	return nil
}
