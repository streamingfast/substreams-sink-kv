package wasm

import (
	"github.com/streamingfast/logging"
	"go.uber.org/zap/zapcore"
)

var zlog, _ = logging.PackageLogger("sink-kv", "github.com/streamingfast/substreams-sink-kv/server/wasm.test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zapcore.DebugLevel))
}
