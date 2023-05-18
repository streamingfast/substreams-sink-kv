package wasmquery

import (
	"github.com/streamingfast/logging"
)

var zlog, _ = logging.PackageLogger("substreams", "github.com/streamingfast/substreams/sink-kv/wasmquery")

func init() {
	logging.InstantiateLoggers()
}
