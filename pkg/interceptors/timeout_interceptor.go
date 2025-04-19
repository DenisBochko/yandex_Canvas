package interceptors

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TimeoutInterceptor struct {
	timeout time.Duration
	log     *zap.Logger
}

func NewTimeoutInterceptor(log *zap.Logger, timeout time.Duration) *TimeoutInterceptor {
	return &TimeoutInterceptor{
		timeout: timeout,
		log:     log,
	}
}

func (t TimeoutInterceptor) UnaryTimeoutInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	done := make(chan struct{})
	var (
		resp any
		err  error
	)

	go func() {
		resp, err = handler(ctx, req)
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.log.Error("gRPC timeout",
			zap.String("method", info.FullMethod),
			zap.Error(ctx.Err()),
		)
		return nil, status.Error(codes.DeadlineExceeded, "request timed out")
	case <-done:
		return resp, err
	}
}
