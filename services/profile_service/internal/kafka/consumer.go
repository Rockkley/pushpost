package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	domain "github.com/rockkley/pushpost/services/profile_service/internal/domain/events"
	"github.com/segmentio/kafka-go"
	"log/slog"
	"time"
)

type EventRouter interface {
	Route(ctx context.Context, event domain.Envelope) error
}

type Consumer struct {
	reader *kafka.Reader
	router EventRouter
	log    *slog.Logger
}

func NewConsumer(brokers []string, topic, groupID string, router EventRouter, log *slog.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})

	return &Consumer{
		reader: reader,
		router: router,
		log:    log.With("component", "kafka_consumer", "topic", topic),
	}
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

		var envelope domain.Envelope

		if err = json.Unmarshal(msg.Value, &envelope); err != nil {
			c.log.Error("invalid envelope",
				slog.Int64("offset", msg.Offset),
				slog.Any("error", err),
			)
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		if err = c.router.Route(ctx, envelope); err != nil {
			c.log.Error("router failed",
				slog.String("event_type", envelope.EventType),
				slog.Any("error", err),
			)
			continue
		}
		fmt.Println("after routing")
		if err = c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit kafka message: %w", err)
		}

		c.log.Info("event processed",
			slog.String("event_type", envelope.EventType),
			slog.Int64("offset", msg.Offset),
		)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
