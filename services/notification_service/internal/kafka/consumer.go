package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Envelope struct {
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type Router interface {
	Route(ctx context.Context, topic string, env Envelope) error
}

type Consumer struct {
	reader *kafka.Reader
	router Router
	log    *slog.Logger
}

func NewConsumer(brokers []string, groupID string, topics []string, router Router, log *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		GroupTopics:    topics,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	return &Consumer{reader: reader, router: router, log: log.With("component", "notification_consumer")}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}
			return fmt.Errorf("fetch kafka message: %w", err)
		}

		env := c.decodeEnvelope(msg.Topic, msg.Value)
		if env.EventType == "" || len(env.Payload) == 0 {
			c.log.Error("cannot decode kafka message, skipping", slog.String("topic", msg.Topic), slog.Int64("offset", msg.Offset))
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err = c.router.Route(ctx, msg.Topic, env); err != nil {
			c.log.Error("handler error, message not committed", slog.String("topic", msg.Topic), slog.String("event_type", env.EventType), slog.Any("error", err))
			continue
		}

		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}
	}
}

// decodeEnvelope supports both formats currently present in the codebase:
// 1) Standard envelope JSON: {"event_type":"...","payload":{...}}
// 2) Raw payload JSON with event type carried by Kafka topic.
func (c *Consumer) decodeEnvelope(topic string, raw []byte) Envelope {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err == nil {
		if env.EventType == "" {
			env.EventType = topic
		}
		if len(env.Payload) > 0 {
			return env
		}
	}

	// fallback: treat full message as payload and topic as event type
	return Envelope{EventType: topic, Payload: raw}
}

func (c *Consumer) Close() error { return c.reader.Close() }
