package wasmquery

import (
	"context"
	"fmt"
	"github.com/streamingfast/substreams/reqctx"
	"go.uber.org/zap"
)

type EnginePool struct {
	engines chan *Engine
}

type EngineFactory = func() (*Engine, error)

func NewEnginePool(ctx context.Context, engineCount uint64, engineFactory EngineFactory) (*EnginePool, error) {
	logger := reqctx.Logger(ctx)

	logger.Info("initializing engine pool", zap.Uint64("engine_count", engineCount))
	engines := make(chan *Engine, engineCount)
	for i := uint64(0); i < engineCount; i++ {
		engine, err := engineFactory()
		if err != nil {
			return nil, fmt.Errorf("unable to create engine: %w", err)
		}
		engines <- engine


	}

	return &EnginePool{
		engines: engines,
	}, nil
}

func (p *EnginePool) Borrow(ctx context.Context) *Engine {
	select {
	case <-ctx.Done():
		return nil
	case w := <-p.engines:
		return w
	}
}

func (p *EnginePool) Return(engine *Engine) {
	p.engines <- engine
}
