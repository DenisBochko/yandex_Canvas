package app

import (
	"context"

	grpcapp "github.com/DenisBochko/yandex_Canvas/internal/app/grpc_app"
	"github.com/DenisBochko/yandex_Canvas/internal/config"
	"github.com/DenisBochko/yandex_Canvas/internal/services/canvas"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
	log        *zap.Logger
	// dbConn     *pgxpool.Pool
}

func New(ctx context.Context, log *zap.Logger, cfg *config.Config) *App {
	// Создаём экземпляр canvas service
	canvasService := canvas.New()

	grpcapp := grpcapp.New(log, canvasService, cfg.GRPC.Port, cfg.GRPC.Timeout)

	return &App{
		GRPCServer: grpcapp,
		log:        log,
		// dbConn:     dbConn,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()

	a.log.Info("stopping database connection")
	// a.dbConn.Close()
}
