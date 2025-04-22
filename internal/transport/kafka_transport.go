package transport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DenisBochko/yandex_Canvas/internal/domain/models"
	"github.com/DenisBochko/yandex_Canvas/pkg/kafka"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type KafkaTransport struct {
	Producer sarama.SyncProducer
	Topic    string
	log      *zap.Logger
}

func New(log *zap.Logger, producer sarama.SyncProducer, topic string) *KafkaTransport {
	return &KafkaTransport{
		Producer: producer,
		Topic:    topic,
		log:      log,
	}
}

func (k *KafkaTransport) SendAddToWhiteListMessage(ctx context.Context, message models.AddToWhiteListMessage) error {
	messageJson, err := json.Marshal(message)

	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.PrepareMessage(k.Topic, messageJson)
	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		k.log.Info("failed to send message", zap.Error(err))
		return fmt.Errorf("failed to send message: %w", err)
	}
	k.log.Info("AddToWhiteListMessage sent to kafka", zap.String("topic", k.Topic), zap.Int32("partition", partition), zap.Int64("offset", offset))

	return nil
}
