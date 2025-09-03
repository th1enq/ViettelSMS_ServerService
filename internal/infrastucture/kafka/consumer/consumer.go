package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	log "github.com/th1enq/ViettelSMS_ServerService/internal/infrastucture/logger"
	"go.uber.org/zap"
)

type HandlerFunc func(ctx context.Context, queueName string, payload []byte) error

type consumerHandler struct {
	handlerFunc HandlerFunc
	logger      *zap.Logger
}

func newConsumerHandler(
	handlerFunc HandlerFunc,
	logger *zap.Logger,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc: handlerFunc,
		logger:      logger,
	}
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group session started")
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group session cleanup")
	return nil
}

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Process messages
	for message := range claim.Messages() {
		h.logger.Debug("Processing message",
			zap.String("topic", message.Topic),
			zap.Int32("partition", message.Partition),
			zap.Int64("offset", message.Offset))

		// Process message with retry logic
		if err := h.processMessageWithRetry(session.Context(), message); err != nil {
			h.logger.Error("Failed to process message after retries",
				zap.String("topic", message.Topic),
				zap.Int64("offset", message.Offset),
				zap.Error(err))

			// Decide whether to continue or stop based on error criticality
			// For now, we'll mark and continue to avoid blocking the consumer
			session.MarkMessage(message, "")
			continue
		}

		// Mark message as processed
		session.MarkMessage(message, "")

		// Commit periodically (optional, can be configured)
		if message.Offset%100 == 0 {
			session.Commit()
		}
	}

	return nil
}

func (h *consumerHandler) processMessageWithRetry(ctx context.Context, message *sarama.ConsumerMessage) error {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff
			backoff := time.Duration(i*i) * time.Second
			h.logger.Warn("Retrying message processing",
				zap.Int("attempt", i+1),
				zap.Duration("backoff", backoff))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err := h.handlerFunc(ctx, message.Topic, message.Value); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

type Consumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
	Stop() error
}

type consumer struct {
	saramaConsumer            sarama.ConsumerGroup
	logger                    *zap.Logger
	queueNameToHandlerFuncMap map[string]HandlerFunc
	cancelFunc                context.CancelFunc
	wg                        sync.WaitGroup
	mu                        sync.RWMutex
	running                   bool
}

func NewConsumer(
	cfg *config.Config,
	logger *zap.Logger,
	consumerID string,
) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	config.Consumer.MaxProcessingTime = 30 * time.Second
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false

	// Add version for compatibility
	config.Version = sarama.V2_6_0_0

	logger.Info("Creating Kafka consumer group",
		zap.Strings("brokers", cfg.Kafka.Address),
		zap.String("consumerID", consumerID))

	saramaConsumer, err := sarama.NewConsumerGroup(cfg.Kafka.Address, consumerID, config)
	if err != nil {
		logger.Error("Failed to create consumer group",
			zap.Strings("brokers", cfg.Kafka.Address),
			zap.String("consumerID", consumerID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &consumer{
		saramaConsumer:            saramaConsumer,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		c.logger.Warn("Cannot register handler while consumer is running",
			zap.String("queue_name", queueName))
		return
	}

	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
	c.logger.Info("Handler registered",
		zap.String("queue_name", queueName))
}

func (c *consumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return fmt.Errorf("consumer already running")
	}
	c.running = true
	c.mu.Unlock()

	logger := log.LoggerWithContext(ctx, c.logger)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel

	// Handle OS signals for graceful shutdown
	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt, syscall.SIGTERM)

	// Start error handler
	go func() {
		for err := range c.saramaConsumer.Errors() {
			logger.Error("Consumer error", zap.Error(err))
		}
	}()

	// Start consumers for each topic
	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		c.wg.Add(1)
		go func(queueName string, handlerFunc HandlerFunc) {
			defer c.wg.Done()

			logger.Info("Starting consumer for queue",
				zap.String("queue_name", queueName))

			handler := newConsumerHandler(handlerFunc, logger)

			for {
				// Check if context is cancelled
				if ctx.Err() != nil {
					logger.Info("Context cancelled, stopping consumer",
						zap.String("queue_name", queueName))
					return
				}

				// Consume messages
				err := c.saramaConsumer.Consume(
					ctx,
					[]string{queueName},
					handler,
				)

				if err != nil {
					if ctx.Err() != nil {
						// Context cancelled, exit gracefully
						return
					}

					logger.Error("Failed to consume messages",
						zap.String("queue_name", queueName),
						zap.Error(err))

					// Wait before retrying
					select {
					case <-ctx.Done():
						return
					case <-time.After(5 * time.Second):
						continue
					}
				}
			}
		}(queueName, handlerFunc)
	}

	// Wait for signal or context cancellation
	select {
	case <-exitSignalChannel:
		logger.Info("Received exit signal, shutting down...")
		c.Stop()
	case <-ctx.Done():
		logger.Info("Context cancelled, shutting down...")
		c.Stop()
	}

	// Wait for all goroutines to finish
	c.wg.Wait()

	c.mu.Lock()
	c.running = false
	c.mu.Unlock()

	logger.Info("Consumer stopped successfully")
	return nil
}

func (c *consumer) Stop() error {
	c.mu.RLock()
	if !c.running {
		c.mu.RUnlock()
		return nil
	}
	c.mu.RUnlock()

	c.logger.Info("Stopping consumer...")

	// Cancel context to stop all consumers
	if c.cancelFunc != nil {
		c.cancelFunc()
	}

	// Close the consumer group
	if err := c.saramaConsumer.Close(); err != nil {
		c.logger.Error("Error closing consumer", zap.Error(err))
		return err
	}

	return nil
}
