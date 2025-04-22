package app

import (
	"context"

	grpcapp "github.com/DenisBochko/yandex_Canvas/internal/app/grpc_app"
	"github.com/DenisBochko/yandex_Canvas/internal/config"
	"github.com/DenisBochko/yandex_Canvas/internal/services/canvas"
	miniostorage "github.com/DenisBochko/yandex_Canvas/internal/storage/minio_storage"
	postgresstorage "github.com/DenisBochko/yandex_Canvas/internal/storage/postgres_storage"
	"github.com/DenisBochko/yandex_Canvas/internal/transport"
	"github.com/DenisBochko/yandex_Canvas/pkg/kafka"
	"github.com/DenisBochko/yandex_Canvas/pkg/minio"
	"github.com/DenisBochko/yandex_Canvas/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	GRPCServer *grpcapp.App
	log        *zap.Logger
	dbConn     *pgxpool.Pool
}

func New(ctx context.Context, log *zap.Logger, cfg *config.Config) *App {
	// ======= Экземпляры клиентов =======

	// Создаём экземпляр подключения к бд
	conn, err := postgres.New(ctx, cfg.Postgres)

	if err != nil {
		log.Info("failed to connect to database", zap.Error(err))
		return nil
	}

	if conn.Ping(ctx) != nil {
		log.Info("failed to ping database", zap.Error(err))
		return nil
	}

	// Создаём экземпляр minIO клиента
	minioClient, err := minio.New(ctx, log, cfg.Minio)
	if err != nil {
		log.Info("failed to connect to minIO", zap.Error(err))
		return nil
	}

	// Создаём экземпляр Kafka клиента
	kafkaProducer, err := kafka.NewSyncProducer(ctx, log, cfg.Kafka)
	if err != nil {
		log.Info("failed to connect to kafka", zap.Error(err))
		return nil
	}

	// ======= Экземпляры сервисов =======

	// Создаём экземпляр postgres storage
	postgresStorage := postgresstorage.New(conn)

	// Создаём экземпляр minio storage
	minioStorage := miniostorage.New(minioClient, cfg.Minio.Bucket)

	// Создаём экэемпляр kafka transport
	kafkaTransport := transport.New(log, kafkaProducer, cfg.Kafka.Topic)

	// Создаём экземпляр canvas service
	canvasService := canvas.New(postgresStorage, minioStorage, kafkaTransport)

	grpcapp := grpcapp.New(log, canvasService, cfg.GRPC.Port, cfg.GRPC.Timeout)

	return &App{
		GRPCServer: grpcapp,
		log:        log,
		dbConn:     conn,
	}
}

func (a *App) Stop() {
	a.GRPCServer.Stop()

	a.log.Info("Stopping database connection")
	a.dbConn.Close()
}
