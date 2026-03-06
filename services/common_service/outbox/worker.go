package outbox

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

type Worker struct {
	repo           OutboxRepository
	publisher      Publisher
	interval       time.Duration
	batchSize      int
	maxAttempts    int
	publishTimeout time.Duration
	stuckAfter     time.Duration
	log            *slog.Logger
}

func NewWorker(repo OutboxRepository, publisher Publisher, cfg WorkerConfig, log *slog.Logger) *Worker {
	if log == nil {
		log = slog.Default()
	}
	return &Worker{
		repo:           repo,
		publisher:      publisher,
		interval:       cfg.Interval,
		batchSize:      cfg.BatchSize,
		maxAttempts:    cfg.MaxAttempts,
		publishTimeout: cfg.PublishTimeout,
		stuckAfter:     cfg.StuckAfter,
		log:            log.With("component", "outbox_worker"),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if err := w.repo.ResetStuck(ctx, w.stuckAfter); err != nil {
		w.log.Error("failed to reset stuck events", "error", err)
	}

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.log.Info("outbox worker started",
		slog.Duration("interval", w.interval),
		slog.Int("batch_size", w.batchSize),
		slog.Int("max_attempts", w.maxAttempts))

	for {
		select {
		case <-ctx.Done():
			w.log.Info("outbox worker stopped")

			return

		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	if ctx.Err() != nil {

		return
	}

	events, err := w.repo.ClaimPending(ctx, w.batchSize, w.maxAttempts)

	if err != nil {
		w.log.Error("failed to claim pending events", slog.Any("error", err))
		return
	}

	if len(events) == 0 {
		return
	}

	w.log.Debug("outbox: processBatch", slog.Int("count", len(events)))

	for _, event := range events {
		w.publishOne(ctx, event)
	}
}

func (w *Worker) publishOne(ctx context.Context, event *OutboxEvent) {
	publishCtx, cancel := context.WithTimeout(ctx, w.publishTimeout)
	defer cancel()

	if err := w.publisher.Publish(publishCtx, event); err != nil {
		if errors.Is(err, context.Canceled) {
			w.log.Info("outbox: publish canceled, shutting down", slog.String("event_id", event.ID.String()))
			return
		}
		w.log.Error("failed to publish event",
			slog.String("event_id", event.ID.String()),
			slog.String("event_type", event.EventType),
			slog.Int("attempts", event.Attempts+1),
			slog.Any("error", err))

		if err = w.repo.IncrementAttempts(ctx, event.ID); err != nil {
			w.log.Error("outbox: failed to increment attempts",
				slog.String("event_id", event.ID.String()),
				slog.Any("error", err),
			)
		}
		return
	}

	if err := w.repo.MarkAsProcessed(ctx, event.ID); err != nil {
		w.log.Error("outbox: failed to mark event as processed",
			slog.String("event_id", event.ID.String()),
			slog.Any("error", err),
		)
	}
}
