package sinker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/shutter"
	sink "github.com/streamingfast/substreams-sink"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	pbsubstreamsrpc "github.com/streamingfast/substreams/pb/sf/substreams/rpc/v2"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const (
	HISTORICAL_BLOCK_FLUSH_EACH = 1000
)

type KVSinker struct {
	*shutter.Shutter
	*sink.Sinker

	operationDB   *db.OperationDB
	flushInterval uint64
	logger        *zap.Logger
	tracer        logging.Tracer

	lastCursor *sink.Cursor
	stats      *Stats
}

func New(sinker *sink.Sinker, dbLoader *db.OperationDB, flushInterval uint64, logger *zap.Logger, tracer logging.Tracer) (*KVSinker, error) {
	s := &KVSinker{
		Shutter:       shutter.New(),
		Sinker:        sinker,
		operationDB:   dbLoader,
		flushInterval: flushInterval,
		logger:        logger,
		tracer:        tracer,

		stats: NewStats(logger),
	}

	s.OnTerminating(func(err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.onTerminating(ctx, err)
	})

	return s, nil
}

func (s *KVSinker) Run(ctx context.Context) {
	cursor, err := s.operationDB.GetCursor(ctx)
	if err != nil && !errors.Is(err, db.ErrCursorNotFound) {
		s.Shutdown(fmt.Errorf("unable to retrieve cursor: %w", err))
		return
	}

	s.Sinker.OnTerminating(s.Shutdown)
	s.OnTerminating(func(err error) {
		s.logger.Info("kv sinker terminating", zap.Stringer("last_block_written", s.stats.lastBlock))
		s.Sinker.Shutdown(err)
	})

	s.OnTerminating(func(_ error) { s.stats.Close() })
	s.stats.OnTerminated(func(err error) { s.Shutdown(err) })

	logEach := 15 * time.Second
	if s.logger.Core().Enabled(zap.DebugLevel) {
		logEach = 5 * time.Second
	}

	s.stats.Start(logEach, cursor)

	s.logger.Info("starting kv sink", zap.Duration("stats_refresh_each", logEach), zap.Stringer("restarting_at", cursor.Block()))
	s.Sinker.Run(ctx, cursor, sink.NewSinkerHandlers(s.handleBlockScopedData, s.handleBlockUndoSignal))
}

func (s *KVSinker) onTerminating(ctx context.Context, err error) {
	if s.lastCursor == nil || err != nil {
		return
	}

	_ = s.operationDB.WriteCursor(ctx, s.lastCursor)
}

var lastBlockCompletedAt = time.Now()

func (s *KVSinker) handleBlockScopedData(ctx context.Context, data *pbsubstreamsrpc.BlockScopedData, isLive *bool, cursor *sink.Cursor) error {
	s.stats.RecordDuractionBetweenBlock(time.Since(lastBlockCompletedAt))

	start := time.Now()
	kvOps := &pbkv.KVOperations{}
	err := proto.Unmarshal(data.GetOutput().MapOutput.Value, kvOps)
	if err != nil {
		return fmt.Errorf("unmarshal database changes: %w", err)
	}

	err = s.operationDB.HandleOperations(ctx, data.Clock.Number, data.FinalBlockHeight, cursor.Step, kvOps)
	if err != nil {
		return fmt.Errorf("handling operation: %w", err)
	}

	BlockCount.Inc()
	if s.shouldFlushKeys(cursor) {
		count, err := s.operationDB.Flush(ctx, cursor)
		if err != nil {
			return fmt.Errorf("flushing operations: %w", err)
		}
		FlushedEntriesCount.AddInt(count)
		flushStart := time.Now()

		FlushCount.Inc()
		s.stats.RecordFlushDuration(time.Since(flushStart))
		s.stats.RecordBlock(cursor.Block())
		s.stats.RecordFinalBlockHeight(data.FinalBlockHeight)
	}

	s.stats.RecordProcessDuration(time.Since(start))
	s.lastCursor = cursor
	lastBlockCompletedAt = time.Now()
	return nil
}

func (s *KVSinker) handleBlockUndoSignal(ctx context.Context, data *pbsubstreamsrpc.BlockUndoSignal, cursor *sink.Cursor) error {

	err := s.operationDB.HandleBlockUndo(ctx, data.LastValidBlock.GetNumber())
	if err != nil {
		return fmt.Errorf("handling undo signal: %w", err)
	}

	_, err = s.operationDB.Flush(ctx, cursor)
	if err != nil {
		return fmt.Errorf("flushing undo operations for: %w", err)
	}

	return nil
}

func (s *KVSinker) shouldFlushKeys(cursor *sink.Cursor) bool {
	currentBlockNum := cursor.Block().Num()
	if cursor.Step == bstream.StepNew {
		return true
	}
	return currentBlockNum%s.flushInterval == 0
}
