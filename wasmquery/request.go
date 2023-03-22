package wasmquery

import (
	"context"

	"go.uber.org/zap"
)

func (v *vm) CurrentRequest() *Request {
	if v.currentRequest == nil {
		panic("current request should not be nil")
	}

	return v.currentRequest
}

type Request struct {
	ctx    context.Context
	logger *zap.Logger
}

func (r *Request) Context() context.Context {
	return r.ctx
}
func (r *Request) Logger() *zap.Logger {
	return r.logger

}
