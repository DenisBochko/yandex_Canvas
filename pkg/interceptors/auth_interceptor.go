package interceptors

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	UIDKey      contextKey = "uid"
	VerifiedKey contextKey = "verified"
)

func UnaryAuthInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("metadata not found")
		return handler(ctx, req)
	}

	if uids := md.Get("uid"); len(uids) > 0 {
		ctx = context.WithValue(ctx, UIDKey, uids[0])
	}
	if verified := md.Get("verified"); len(verified) > 0 {
		ctx = context.WithValue(ctx, VerifiedKey, verified[0])
	}

	// передаём управление дальше
	return handler(ctx, req)
}
