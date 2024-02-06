package sinker

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	LIVE_BLOCK_FLUSH_EACH       = 1
)

type KVSinker struct {
	*shutter.Shutter
	*sink.Sinker

	dbLoader      db.Loader
	flushInterval time.Duration
	logger        *zap.Logger
	tracer        logging.Tracer

	lastCursor *sink.Cursor
	stats      *Stats
}

func New(sinker *sink.Sinker, dbLoader db.Loader, flushInterval time.Duration, logger *zap.Logger, tracer logging.Tracer) (*KVSinker, error) {
	s := &KVSinker{
		Shutter:       shutter.New(),
		Sinker:        sinker,
		dbLoader:      dbLoader,
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
	cursor, err := s.dbLoader.GetCursor(ctx)
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

	_ = s.dbLoader.WriteCursor(ctx, s.lastCursor)
}

func (s *KVSinker) handleBlockScopedData(ctx context.Context, data *pbsubstreamsrpc.BlockScopedData, isLive *bool, cursor *sink.Cursor) error {
	kvOps := &pbkv.KVOperations{}
	err := proto.Unmarshal(data.GetOutput().MapOutput.Value, kvOps)
	if err != nil {
		return fmt.Errorf("unmarshal database changes: %w", err)
	}

	s.dbLoader.StoreReverseOperations(ctx, data.Clock.GetNumber(), kvOps.Operations)
	s.dbLoader.AddOperations(kvOps)

	s.lastCursor = cursor

	blockRef := cursor.Block()
	if blockRef.Num()%s.batchBlockModulo(data, isLive) == 0 {
		flushStart := time.Now()
		count, err := s.dbLoader.Flush(ctx, cursor)
		if err != nil {
			return fmt.Errorf("failed to flush: %w", err)
		}

		s.stats.RecordBlock(blockRef)
		FlushCount.Inc()
		FlushedEntriesCount.AddInt(count)
		FlushDuration.AddInt64(time.Since(flushStart).Nanoseconds())
	}

	//todo Handle the deletion when a block is considered as LIB

	return nil
}

func (s *KVSinker) handleBlockUndoSignal(ctx context.Context, data *pbsubstreamsrpc.BlockUndoSignal, cursor *sink.Cursor) error {
	err := s.dbLoader.HandleBlockUndo(ctx, data.LastValidBlock.GetNumber(), cursor)
	if err != nil {
		return fmt.Errorf("unable to handle undo signal: %w", err)
	}

	//todo: Do not forget to update the cursor...

	return fmt.Errorf("received undo signal but there is no handling of undo, this is because you used `--undo-buffer-size=0` which is invalid right now")
}

func (s *KVSinker) batchBlockModulo(blockData *pbsubstreamsrpc.BlockScopedData, isLive *bool) uint64 {
	if isLive == nil {
		panic(fmt.Errorf("liveness checker has been disabled on the Sinker instance, this is invalid in the context of 'substreams-sink-postgres'"))
	}

	if *isLive {
		return LIVE_BLOCK_FLUSH_EACH
	}

	if s.flushInterval > 0 {
		return uint64(s.flushInterval)
	}

	return HISTORICAL_BLOCK_FLUSH_EACH
}
