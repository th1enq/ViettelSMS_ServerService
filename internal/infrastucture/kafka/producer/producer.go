package producer

import (
	"github.com/IBM/sarama"
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/mq"
	"go.uber.org/zap"
)

type MessageBroker interface {
	Send(message mq.Message) error
}

type messageBroker struct {
	producer sarama.SyncProducer
	logger   *zap.Logger
}

func NewBroker(cfg *config.Config, logger *zap.Logger) (MessageBroker, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(cfg.Kafka.Address, saramaConfig)
	if err != nil {
		return nil, err
	}
	return &messageBroker{
		producer: producer,
		logger:   logger,
	}, nil
}

var MessageBrokerSet = wire.NewSet(NewBroker)

func (d messageBroker) Send(event mq.Message) error {
	var headers []sarama.RecordHeader

	for k, v := range event.Headers {
		headers = append(headers, sarama.RecordHeader{
			Key:   sarama.ByteEncoder(k),
			Value: sarama.ByteEncoder(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   event.Topic,
		Value:   sarama.ByteEncoder(event.Body),
		Headers: headers,
	}

	d.logger.Info("Sending message to broker",
		zap.String("topic", event.Topic),
		zap.ByteString("body", event.Body),
	)

	_, _, err := d.producer.SendMessage(msg)

	return err
}
