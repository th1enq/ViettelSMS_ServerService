package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/IBM/sarama"
	"github.com/gammazero/workerpool"
	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/mq"
	"github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/kafka/producer"
	"go.uber.org/zap"
)

const (
	WORKER_NUMBER   = 5
	MAX_RETRY       = 3
	DLQ             = "dead-letter"
	RETRY_TOPIC_10S = "retry-10s"
	RETRY_TOPIC_1M  = "retry-1m"
	RETRY_TOPIC_5M  = "retry-5m"

	// Header keys
	HEADER_RETRY_COUNT  = "retry-count"
	HEADER_ORIGIN_TOPIC = "origin-topic"
	HEADER_ERROR        = "error"
	HEADER_RETRY_TIME   = "retry-time"
)

type (
	HandlerFunc func(ctx context.Context, queueName string, payload []byte) error

	consumerHandler struct {
		handlerFunc       HandlerFunc
		exitSignalChannel chan os.Signal
		prod              producer.MessageBroker
		logger            *zap.Logger
		handlerMap        map[string]HandlerFunc // Add this to access original handlers
	}

	Consumer interface {
		RegisterHandler(queueName string, handlerFunc HandlerFunc)
		Start(ctx context.Context) error
	}

	consumer struct {
		saramaConsumer            sarama.ConsumerGroup
		logger                    *zap.Logger
		prod                      producer.MessageBroker
		queueNameToHandlerFuncMap map[string]HandlerFunc
	}
)

func newConsumerHandler(
	handlerFunc HandlerFunc,
	exitSignalChannel chan os.Signal,
	prod producer.MessageBroker,
	logger *zap.Logger,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc:       handlerFunc,
		exitSignalChannel: exitSignalChannel,
		logger:            logger,
		prod:              prod,
		handlerMap:        make(map[string]HandlerFunc),
	}
}

func newConsumerHandlerWithMap(
	handlerFunc HandlerFunc,
	exitSignalChannel chan os.Signal,
	prod producer.MessageBroker,
	logger *zap.Logger,
	handlerMap map[string]HandlerFunc,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc:       handlerFunc,
		exitSignalChannel: exitSignalChannel,
		logger:            logger,
		prod:              prod,
		handlerMap:        handlerMap,
	}
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func getHeader(headers []*sarama.RecordHeader, key string) (string, bool) {
	for _, h := range headers {
		if string(h.Key) == key {
			return string(h.Value), true
		}
	}
	return "", false
}

func getHeaderOrDefault(headers []*sarama.RecordHeader, key, defaultValue string) string {
	if value, ok := getHeader(headers, key); ok {
		return value
	}
	return defaultValue
}

func (h *consumerHandler) sendToRetryQueue(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
	err error,
) {
	retryCount := 0
	if header, ok := getHeader(msg.Headers, HEADER_RETRY_COUNT); ok {
		if count, parseErr := strconv.Atoi(header); parseErr == nil {
			retryCount = count
		}
	}
	retryCount++

	if retryCount > MAX_RETRY {
		h.logger.Warn("Message moved to DLQ",
			zap.String("topic", msg.Topic),
			zap.ByteString("body", msg.Value),
			zap.Int("retry_count", retryCount),
			zap.Error(err))

		h.prod.Send(mq.Message{
			Headers: map[string]string{
				HEADER_ORIGIN_TOPIC: msg.Topic,
				HEADER_RETRY_COUNT:  fmt.Sprintf("%d", retryCount),
				HEADER_ERROR:        err.Error(),
			},
			Body:  msg.Value,
			Topic: DLQ,
		})
		return
	}

	var retryTopic string
	switch retryCount {
	case 1:
		retryTopic = RETRY_TOPIC_10S
	case 2:
		retryTopic = RETRY_TOPIC_1M
	case 3:
		retryTopic = RETRY_TOPIC_5M
	default:
		// Fallback to DLQ if retry count exceeds expected values
		retryTopic = DLQ
	}

	h.logger.Info("Sending message to retry topic",
		zap.String("retry_topic", retryTopic),
		zap.ByteString("body", msg.Value),
		zap.Int("retry_count", retryCount),
		zap.Error(err))

	h.prod.Send(mq.Message{
		Headers: map[string]string{
			HEADER_ORIGIN_TOPIC: msg.Topic,
			HEADER_RETRY_COUNT:  fmt.Sprintf("%d", retryCount),
			HEADER_ERROR:        err.Error(),
		},
		Body:  msg.Value,
		Topic: retryTopic,
	})
}

func (h *consumerHandler) handleRetryMessage(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
) error {
	originTopic, ok := getHeader(msg.Headers, HEADER_ORIGIN_TOPIC)
	if !ok {
		h.logger.Error("Missing origin topic header in retry message", zap.String("topic", msg.Topic), zap.ByteString("body", msg.Value))
		return fmt.Errorf("missing origin topic header")
	}

	var delay time.Duration
	switch msg.Topic {
	case RETRY_TOPIC_10S:
		delay = 10 * time.Second
	case RETRY_TOPIC_1M:
		delay = 1 * time.Minute
	case RETRY_TOPIC_5M:
		delay = 5 * time.Minute
	default:
		delay = 0
	}

	h.logger.Info("Retrying message",
		zap.String("origin_topic", originTopic),
		zap.String("retry_topic", msg.Topic),
		zap.Duration("delay", delay),
		zap.ByteString("body", msg.Value))

	// Simulate delay
	time.Sleep(delay)

	headers := map[string]string{}

	for _, header := range msg.Headers {
		headers[string(header.Key)] = string(header.Value)
	}

	h.prod.Send(mq.Message{
		Headers: headers,
		Body:    msg.Value,
		Topic:   originTopic,
	})

	return nil
}

func (h *consumerHandler) handleDLQMessage(
	session sarama.ConsumerGroupSession,
	msg *sarama.ConsumerMessage,
) error {
	// Log DLQ message for monitoring/alerting
	originTopic := getHeaderOrDefault(msg.Headers, HEADER_ORIGIN_TOPIC, "unknown")
	retryCount := getHeaderOrDefault(msg.Headers, HEADER_RETRY_COUNT, "0")
	errorMsg := getHeaderOrDefault(msg.Headers, HEADER_ERROR, "unknown error")

	h.logger.Error("Message in Dead Letter Queue",
		zap.String("origin_topic", originTopic),
		zap.String("retry_count", retryCount),
		zap.String("error", errorMsg),
		zap.ByteString("body", msg.Value))

	// Here you could implement additional DLQ handling like:
	// - Sending alerts
	// - Storing in database for manual review
	// - Metrics reporting

	return nil
}

func (h *consumerHandler) stopWorkerAndCommit(
	pool *workerpool.WorkerPool,
	session sarama.ConsumerGroupSession,
) {
	const shutdownTimeout = 5 * time.Second

	done := make(chan struct{})
	go func() {
		pool.StopWait()
		close(done)
	}()

	select {
	case <-done:
		h.logger.Info("Worker pool stopped successfully")
	case <-time.After(shutdownTimeout):
		h.logger.Warn("Worker pool stop timed out, forcing commit")
	}
	session.Commit()
}

func (h *consumerHandler) handleMessageAsync(
	pool *workerpool.WorkerPool,
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
	msg *sarama.ConsumerMessage,
) {
	m := msg
	pool.Submit(func() {
		var err error

		workerID := fmt.Sprintf("worker-%d", time.Now().UnixNano())

		// Handle different types of messages
		switch m.Topic {
		case RETRY_TOPIC_10S, RETRY_TOPIC_1M, RETRY_TOPIC_5M:
			h.logger.Debug("Handling retry message", zap.String("worker_id", workerID), zap.String("topic", m.Topic), zap.String("header", fmt.Sprintf("%v", m.Headers)), zap.ByteString("body", m.Value))
			err = h.handleRetryMessage(session, m)
		case DLQ:
			h.logger.Debug("Handling DLQ message", zap.String("worker_id", workerID), zap.String("topic", m.Topic), zap.String("header", fmt.Sprintf("%v", m.Headers)), zap.ByteString("body", m.Value))
			err = h.handleDLQMessage(session, m)
		default:
			h.logger.Debug("Handling regular message", zap.String("worker_id", workerID), zap.String("topic", m.Topic), zap.String("header", fmt.Sprintf("%v", m.Headers)), zap.ByteString("body", m.Value))
			// Regular message processing
			err = h.handlerFunc(context.Background(), m.Topic, m.Value)
		}

		if err != nil {
			h.logger.Error("Error processing message", zap.String("worker_id", workerID), zap.String("topic", m.Topic), zap.String("header", fmt.Sprintf("%v", m.Headers)), zap.ByteString("body", m.Value), zap.Error(err))
			// Only send to retry queue for regular messages and retry failures
			if m.Topic != DLQ {
				h.sendToRetryQueue(session, m, err)
			}
			session.MarkMessage(m, "error")
			return
		}

		session.MarkMessage(m, "ok")
	})
}

func (h *consumerHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	h.logger.Info("Starting message consumption", zap.String("topic", claim.Topic()))
	workerPool := workerpool.New(WORKER_NUMBER)
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				h.stopWorkerAndCommit(workerPool, session)
				return nil
			}
			h.handleMessageAsync(workerPool, session, claim, msg)
		case <-h.exitSignalChannel:
			h.logger.Info("Exit signal received, stopping consumer")
			h.stopWorkerAndCommit(workerPool, session)
			return nil
		}
	}
}

func NewConsumer(
	cfg *config.Config,
	prod producer.MessageBroker,
	logger *zap.Logger,
) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.MaxProcessingTime = 30 * time.Second
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false // Manual commit for better control

	saramaConsumer, err := sarama.NewConsumerGroup(cfg.Broker.Address, cfg.Broker.ClientID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &consumer{
		saramaConsumer:            saramaConsumer,
		prod:                      prod,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

var ConsumerSet = wire.NewSet(NewConsumer)

func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.logger.Info("Registering handler for queue", zap.String("queue", queueName), zap.String("handler", fmt.Sprintf("%T", handlerFunc)))
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}

func (c *consumer) Start(ctx context.Context) error {
	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt)

	// Start consumers for regular topics
	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			c.logger.Info("Starting consumer for topic", zap.String("topic", queueName))
			if err := c.saramaConsumer.Consume(
				ctx,
				[]string{queueName, RETRY_TOPIC_10S, RETRY_TOPIC_1M, RETRY_TOPIC_5M, DLQ},
				newConsumerHandler(handlerFunc, exitSignalChannel, c.prod, c.logger),
			); err != nil {
				c.logger.Error("Failed to start consumer", zap.String("queue", queueName), zap.Error(err))
			}
		}(queueName, handlerFunc)
	}

	c.logger.Info("All consumers started, waiting for exit signal")
	<-exitSignalChannel
	c.logger.Info("Exit signal received, shutting down consumer")
	return c.saramaConsumer.Close()
}
