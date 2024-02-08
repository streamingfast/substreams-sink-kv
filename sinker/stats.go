package sinker

import (
	"time"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/dmetrics"
	"github.com/streamingfast/shutter"
	sink "github.com/streamingfast/substreams-sink"
	"go.uber.org/zap"
)

type Stats struct {
	*shutter.Shutter

	dbFlushRate                    *dmetrics.AvgRatePromCounter
	flushedEntries                 *dmetrics.ValueFromMetric
	lastBlock                      bstream.BlockRef
	logger                         *zap.Logger
	blockRate                      *dmetrics.AvgRatePromCounter
	flushDuration                  *dmetrics.AvgDurationCounter
	blockScopedDataProcessDuration *dmetrics.AvgDurationCounter
	durationBetweenBlock           *dmetrics.AvgDurationCounter
	finalBlockHeight               uint64
}

func NewStats(logger *zap.Logger) *Stats {
	return &Stats{
		Shutter: shutter.New(),

		dbFlushRate:                    dmetrics.MustNewAvgRateFromPromCounter(FlushCount, 1*time.Second, 30*time.Second, "flush"),
		blockRate:                      dmetrics.MustNewAvgRateFromPromCounter(BlockCount, 1*time.Second, 30*time.Second, "block"),
		flushedEntries:                 dmetrics.NewValueFromMetric(FlushedEntriesCount, "entries"),
		flushDuration:                  dmetrics.NewAvgDurationCounter(30*time.Second, time.Millisecond, "flush duration"),
		blockScopedDataProcessDuration: dmetrics.NewAvgDurationCounter(30*time.Second, time.Millisecond, "process duration"),
		durationBetweenBlock:           dmetrics.NewAvgDurationCounter(30*time.Second, time.Millisecond, "duration between block"),
		lastBlock:                      unsetBlockRef{},
		logger:                         logger,
	}
}

func (s *Stats) RecordBlock(block bstream.BlockRef) {
	s.lastBlock = block
}
func (s *Stats) RecordFinalBlockHeight(num uint64) {
	s.finalBlockHeight = num
}

func (s *Stats) RecordFlushDuration(duration time.Duration) {
	s.flushDuration.AddDuration(duration)
}
func (s *Stats) RecordProcessDuration(duration time.Duration) {
	s.blockScopedDataProcessDuration.AddDuration(duration)
}
func (s *Stats) RecordDuractionBetweenBlock(duration time.Duration) {
	s.durationBetweenBlock.AddDuration(duration)
}

func (s *Stats) Start(each time.Duration, cursor *sink.Cursor) {
	if !cursor.IsBlank() {
		s.lastBlock = cursor.Block()
	}

	if s.IsTerminating() || s.IsTerminated() {
		panic("already shutdown, refusing to start again")
	}

	go func() {
		ticker := time.NewTicker(each)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.LogNow()
			case <-s.Terminating():
				return
			}
		}
	}()
}

func (s *Stats) LogNow() {
	// Logging fields order is important as it affects the final rendering, we carefully ordered
	// them so the development logs looks nicer.
	s.logger.Info("substreams kv stats",
		zap.Stringer("db_flush_rate", s.dbFlushRate),
		zap.String("flush_duration", s.flushDuration.String()),
		zap.String("process_duration", s.blockScopedDataProcessDuration.String()),
		zap.String("duration_between_block", s.durationBetweenBlock.String()),
		zap.Stringer("block_rate", s.blockRate),
		zap.Uint64("flushed_entries", s.flushedEntries.ValueUint()),
		zap.Stringer("last_block", s.lastBlock),
		zap.Uint64("final_block_height", s.finalBlockHeight),
	)
}

func (s *Stats) Close() {
	s.dbFlushRate.SyncNow()
	s.LogNow()

	s.Shutdown(nil)
}

type unsetBlockRef struct{}

func (unsetBlockRef) ID() string     { return "" }
func (unsetBlockRef) Num() uint64    { return 0 }
func (unsetBlockRef) String() string { return "None" }
