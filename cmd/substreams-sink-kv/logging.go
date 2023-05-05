package main

import (
	"github.com/streamingfast/cli"
	"github.com/streamingfast/logging"
)

var zlog, tracer = logging.RootLogger("sink-kv", "github.com/streamingfast/substreams-sink-kv/cmd/substreams-sink-kv")

func init() {
	cli.SetLogger(zlog, tracer)
}
