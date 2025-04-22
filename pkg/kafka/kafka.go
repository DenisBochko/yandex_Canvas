package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type KafkaConfig struct {
	Brokers []string `yaml:"KAFKA_BROKERS" env-required:"true"`
	Topic   string   `yaml:"KAFKA_TOPIC" env-required:"true"`
}

func NewSyncProducer(ctx context.Context, log *zap.Logger, cfg KafkaConfig) (sarama.SyncProducer, error) {
	brokers := cfg.Brokers

	if len(brokers) == 0 {
		return nil, sarama.ErrBrokerNotFound
	}

	// Конфигурация Kafka
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0 
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	// Создаём топик, если его нет
	if err := createTopic(cfg, config, log); err != nil {
		log.Error("failed to create topic", zap.Error(err))
		return nil, err
	}

	// Создание продюсера
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Error("failed to create kafka producer", zap.Error(err))
		return nil, err
	}

	log.Info("Kafka producer created", zap.Strings("brokers", brokers), zap.String("topic", cfg.Topic))
	return producer, nil
}

// createTopic создаёт топик, если он не существует
func createTopic(cfg KafkaConfig, config *sarama.Config, log *zap.Logger) error {
	admin, err := sarama.NewClusterAdmin(cfg.Brokers, config)
	if err != nil {
		return err
	}
	defer admin.Close()

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     5,
		ReplicationFactor: 3,
	}

	err = admin.CreateTopic(cfg.Topic, topicDetail, false)
	if err != nil {
		if errors.Is(err, sarama.ErrTopicAlreadyExists) {
			log.Info("topic already exists", zap.String("topic", cfg.Topic))
			return nil
		}
		return err
	}

	log.Info("topic created", zap.String("topic", cfg.Topic))
	return nil
}

func NewAsyncProducer(ctx context.Context, cfg KafkaConfig) (sarama.AsyncProducer, error) {
	brokers := cfg.Brokers

	if len(brokers) == 0 {
		return nil, sarama.ErrBrokerNotFound
	}

	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(brokers, config)

	return producer, err
}

func PrepareMessage(topic string, message []byte) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}

	return msg
}