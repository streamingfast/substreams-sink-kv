package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/streamingfast/derr"
	"github.com/streamingfast/substreams-sink-kv/db"
	. "github.com/streamingfast/substreams-sink-kv/server"
	"github.com/streamingfast/substreams/manifest"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
)

var serveCmd = &cobra.Command{
	Use:   `serve <dsn> <spkg>`,
	Short: "Serves the contents of an spkg",
	Args:  cobra.ExactArgs(2),
	RunE:  serveE,
}

func init() {
	serveCmd.Flags().String("listen-addr", "", "Listen via GRPC Connect-Web on this address")
	serveCmd.Flags().Bool("listen-ssl-self-signed", false, "Listen with an HTTPS server (with self-signed certificate)")

	rootCmd.AddCommand(serveCmd)
}

func serveE(cmd *cobra.Command, args []string) error {

	// parse args
	dsn := args[0]
	manifestPath := args[1]
	listenAddr := mustGetString(cmd, "listen-addr")
	manifestReader := manifest.NewReader(manifestPath)

	// get config type and check it matches either option
	pkg, err := manifestReader.Read()
	if err != nil {
		return fmt.Errorf("read manifest %q: %w", manifestPath, err)
	}
	sinkConfigType := pkg.SinkConfig.TypeUrl
	if sinkConfigType != "pcs.services.v1.WASMQueryService" || sinkConfigType != "pcs.services.v1.GenericService" {
		return fmt.Errorf("%s", "invalid sink_config type")
	}

	zlog.Info("sink to kv",
		zap.String("dsn", dsn),
		zap.String("manifest_path", manifestPath),
	)

	// init db
	kvDB, err := db.New(dsn, zlog, tracer)
	if err != nil {
		return fmt.Errorf("new psql loader: %w", err)
	}

	if sinkConfigType == "pcs.services.v1.GenericService" {
		go func() {
			zlog.Info("starting to listen on: ", zap.String("addr", listenAddr))
			ListenConnectWeb(listenAddr, kvDB, zlog, mustGetBool(cmd, "run-listen-ssl-self-signed"))
		}()
	} else {
		go wasmQueryServiceServe(kvDB, pkg.SinkConfig)
	}

	signalHandler := derr.SetupSignalHandler(0 * time.Second)
	zlog.Info("ready, waiting for signal to quit")
	select {
	case <-signalHandler:
		zlog.Info("received termination signal, quitting application")
	}

	//srv.RegisterService(&grpc.ServiceDesc{
	//	ServiceName:
	//})

	//s.Launch(":7878")

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

func wasmQueryServiceServe(db *db.DB, configType *anypb.Any) {

}

func mustGetString(cmd *cobra.Command, flagName string) string {
	val, err := cmd.Flags().GetString(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}

func mustGetBool(cmd *cobra.Command, flagName string) bool {
	val, err := cmd.Flags().GetBool(flagName)
	if err != nil {
		panic(fmt.Sprintf("flags: couldn't find flag %q", flagName))
	}
	return val
}
