package grpcapp

import (
	"fmt"
	"net"
	"time"

	grpcHandlersCanvas "github.com/DenisBochko/yandex_Canvas/internal/grpc_handlers/canvas"
	"github.com/DenisBochko/yandex_Canvas/pkg/interceptors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *zap.Logger, canvasService grpcHandlersCanvas.CanvasService, port int, timeout time.Duration) *App {
	loggerInterceptor := interceptors.NewLoggerInterceptor(log)
	timeoutInterceptor := interceptors.NewTimeoutInterceptor(log, timeout)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggerInterceptor.UnaryLoggerInterceptor,
			timeoutInterceptor.UnaryTimeoutInterceptor,
		),
	)

	grpcHandlersCanvas.Register(grpcServer, canvasService)

	return &App{
		log:        log,
		gRPCServer: grpcServer,
		port:       port,
	}
}

func (a *App) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("failed to listen tcp: %v", err)
	}

	a.log.Info("gRPC server is running", zap.Int("port", a.port), zap.String("addres", lis.Addr().String()))

	// Запускаем сервер на порту
	if err := a.gRPCServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (a *App) Stop() {
	a.log.Info("Stopping gRPC server", zap.Int("port", a.port))
	a.gRPCServer.GracefulStop()
}
