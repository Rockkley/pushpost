package kafka

import (
	"context"
	"fmt"
	"github.com/rockkley/pushpost/services/common_service/outbox"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"time"
)

type Publisher struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewPublisher(brokers []string, log *slog.Logger) *Publisher {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireOne,
		Async:        false}
	return &Publisher{
		writer: w,
		log:    log,
	}
}

func (p *Publisher) Publish(ctx context.Context, event *outbox.OutboxEvent) error {
	msg := kafka.Message{
		Topic: event.EventType,
		Key:   []byte(event.AggregateID),
		Value: event.Payload,
		Headers: []kafka.Header{
			{Key: "aggregate_type", Value: []byte(event.AggregateType)},
			{Key: "event_id", Value: []byte(event.ID.String())},
			{Key: "published_at", Value: []byte(time.Now().UTC().Format(time.RFC3339))},
		},
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("kafka publish %q aggregate=%s: %w", event.EventType, event.AggregateType, err)
	}
	p.log.Debug("kafka: event published",
		slog.String("event_type", event.EventType),
		slog.String("aggregate_type", event.AggregateType),
		slog.String("aggregate_id", event.AggregateID),
		slog.String("event_id", event.ID.String()),
	)

	return nil
}

func (p *Publisher) Close() error {
	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close kafka writer: %w", err)
	}
	return nil
}
