package sinker

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/streamingfast/bstream"
	"github.com/streamingfast/logging"
	"github.com/streamingfast/shutter"
	sink "github.com/streamingfast/substreams-sink"
	"github.com/streamingfast/substreams-sink-kv/db"
	pbkv "github.com/streamingfast/substreams-sink-kv/pb/substreams/sink/kv/v1"
	"github.com/streamingfast/substreams/client"
	"github.com/streamingfast/substreams/manifest"
	pbsubstreams "github.com/streamingfast/substreams/pb/sf/substreams/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const BLOCK_PROGRESS = 1000

type Config struct {
	BlockRange       string
	Pkg              *pbsubstreams.Package
	OutputModule     *pbsubstreams.Module
	OutputModuleName string
	OutputModuleHash manifest.ModuleHash
	ClientConfig     *client.SubstreamsClientConfig
	DBLoader         *db.Loader
}

type KVSinker struct {
	*shutter.Shutter

	DBLoader         *db.Loader
	Pkg              *pbsubstreams.Package
	OutputModule     *pbsubstreams.Module
	OutputModuleName string
	OutputModuleHash manifest.ModuleHash
	ClientConfig     *client.SubstreamsClientConfig

	sink       *sink.Sinker
	lastCursor *sink.Cursor

	stats *Stats

	blockRange *bstream.Range

	logger *zap.Logger
	tracer logging.Tracer
}

func New(config *Config, logger *zap.Logger, tracer logging.Tracer) (*KVSinker, error) {
	s := &KVSinker{
		Shutter: shutter.New(),
		stats:   NewStats(logger),
		logger:  logger,
		tracer:  tracer,

		DBLoader:         config.DBLoader,
		Pkg:              config.Pkg,
		OutputModule:     config.OutputModule,
		OutputModuleName: config.OutputModuleName,
		OutputModuleHash: config.OutputModuleHash,
		ClientConfig:     config.ClientConfig,
	}

	s.OnTerminating(func(err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		s.Stop(ctx, err)
	})

	var err error
	s.blockRange, err = resolveBlockRange(config.BlockRange, config.OutputModule)
	if err != nil {
		return nil, fmt.Errorf("resolve block range: %w", err)
	}

	return s, nil
}

func (s *KVSinker) Start(ctx context.Context) error {
	cursor, err := s.DBLoader.GetCursor(ctx)
	if err != nil && !errors.Is(err, db.ErrCursorNotFound) {
		return fmt.Errorf("unable to retrieve cursor: %w", err)
	}

	if errors.Is(err, db.ErrCursorNotFound) {
		cursorStartBlock := s.OutputModule.InitialBlock
		if s.blockRange.StartBlock() > 0 {
			cursorStartBlock = s.blockRange.StartBlock() - 1
		}

		cursor = sink.NewCursor("", bstream.NewBlockRef("", cursorStartBlock))

		if err = s.DBLoader.WriteCursor(ctx, cursor); err != nil {
			return fmt.Errorf("failed to create initial cursor: %w", err)
		}
	}

	s.OnTerminating(func(_ error) { s.stats.Close() })
	s.stats.OnTerminated(func(err error) { s.Shutdown(err) })
	s.stats.Start(2 * time.Second)

	return s.Run(ctx)
}

func (s *KVSinker) Stop(ctx context.Context, err error) {
	if s.lastCursor == nil || err != nil {
		return
	}

	_ = s.DBLoader.WriteCursor(ctx, s.lastCursor)
}

func (s *KVSinker) Run(ctx context.Context) error {
	cursor, err := s.DBLoader.GetCursor(ctx)
	if err != nil {
		return fmt.Errorf("unable to retrieve cursor: %w", err)
	}

	s.sink, err = sink.New(
		sink.SubstreamsModeProduction,
		s.Pkg.Modules,
		s.OutputModule,
		s.OutputModuleHash,
		s.handleBlockScopeData,
		s.ClientConfig,
		[]pbsubstreams.ForkStep{pbsubstreams.ForkStep_STEP_NEW, pbsubstreams.ForkStep_STEP_UNDO},
		s.logger,
		s.tracer,
	)
	if err != nil {
		return fmt.Errorf("unable to create sink: %w", err)
	}

	s.sink.OnTerminating(s.Shutdown)
	s.OnTerminating(func(err error) {
		s.logger.Info("terminating sink")
		s.sink.Shutdown(err)
	})

	if err := s.sink.Start(ctx, s.blockRange, cursor); err != nil {
		return fmt.Errorf("sink failed: %w", err)
	}

	return nil
}

func (s *KVSinker) handleBlockScopeData(ctx context.Context, cursor *sink.Cursor, data *pbsubstreams.BlockScopedData) error {
	for _, output := range data.Outputs {
		if output.Name != s.OutputModuleName {
			continue
		}

		kvOps := &pbkv.KVOperations{}
		err := proto.Unmarshal(output.GetMapOutput().GetValue(), kvOps)
		if err != nil {
			return fmt.Errorf("unmarshal database changes: %w", err)
		}

		s.DBLoader.AddOperations(kvOps)
	}

	s.lastCursor = cursor

	if cursor.Block.Num()%BLOCK_PROGRESS == 0 {
		flushStart := time.Now()
		count, err := s.DBLoader.Flush(ctx, hex.EncodeToString(s.OutputModuleHash), cursor)
		if err != nil {
			return fmt.Errorf("failed to flush: %w", err)
		}

		flushDuration := time.Since(flushStart)
		FlushCount.Inc()
		FlushedEntriesCount.AddUint64(uint64(count))
		FlushDuration.AddInt(int(flushDuration.Nanoseconds()))
	}

	return nil
}
