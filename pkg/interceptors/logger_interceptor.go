package interceptors

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type LoggerInterceptor struct {
	log *zap.Logger
}

func NewLoggerInterceptor(log *zap.Logger) *LoggerInterceptor {
	return &LoggerInterceptor{
		log: log,
	}
}

func (l *LoggerInterceptor) UnaryLoggerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	l.log.Info("gRPC request",
		zap.String("method", info.FullMethod),
		zap.Any("request", req),
	)

	resp, err := handler(ctx, req)

	l.log.Info("gRPC response",
		zap.String("method", info.FullMethod),
		zap.Any("response", resp),
		zap.Error(err),
	)
	return resp, err
}
