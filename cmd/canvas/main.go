package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/DenisBochko/yandex_Canvas/internal/app"
	"github.com/DenisBochko/yandex_Canvas/internal/config"
	"github.com/DenisBochko/yandex_Canvas/pkg/logger"
)

func main() {
	ctx := context.Background()

	// Обработка сигналов завершения
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// кофигурация
	cfg := config.MustLoad()

	// инициализация логгера
	logger := logger.SetupLogger(cfg.Env)
	defer logger.Sync()

	// инициализация приложения и его запуск
	application := app.New(ctx, logger, cfg)
	go application.GRPCServer.Run()

	// graceful shutdown
	// Ожидаем сигнал завершения
	<-ctx.Done()
	logger.Info("Stopping Canvas service...")

	application.Stop()

	logger.Info("Canvas service stopped")
}
