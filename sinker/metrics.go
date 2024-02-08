package sinker

import "github.com/streamingfast/dmetrics"

func RegisterMetrics() {
	metrics.Register()
}

var metrics = dmetrics.NewSet()

var FlushedEntriesCount = metrics.NewCounter("substreams_sink_kv_flushed_entries_count", "The number of flushed entries")
var FlushCount = metrics.NewCounter("substreams_sink_kv_store_flush_count", "The amount of flush that happened so far")
var BlockCount = metrics.NewCounter("substreams_sink_kv_store_block_count", "The block processed so far")
var FlushDuration = metrics.NewCounter("substreams_sink_kv_store_flush_duration", "The amount of time spent flushing cache to db")
