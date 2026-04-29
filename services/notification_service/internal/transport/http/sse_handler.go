package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	"github.com/rockkley/pushpost/services/notification_service/internal/delivery/inapp"
)

type SSEHandler struct{ rdb *goredis.Client }

func NewSSEHandler(rdb *goredis.Client) *SSEHandler { return &SSEHandler{rdb: rdb} }

func (h *SSEHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.From(r.Context()).With(slog.String("op", "NotificationSSE.Subscribe"))
	userID, ok := commonmiddleware.UserIDFromContext(r.Context())

	if !ok || userID == uuid.Nil {
		http.Error(w, `{"code":"unauthorized"}`, http.StatusUnauthorized)

		return
	}

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	startID := r.Header.Get("Last-Event-ID")

	if startID == "" {
		startID = "$"
	}

	streamKey := inapp.StreamKey(userID)

	for {
		records, err := h.rdb.XRead(r.Context(), &goredis.XReadArgs{
			Streams: []string{streamKey, startID},
			Count:   20,
			Block:   25 * time.Second,
		}).Result()

		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			if errors.Is(err, goredis.Nil) {
				if _, writeErr := fmt.Fprintf(w, "event: ping\ndata: {}\n\n"); writeErr != nil {
					log.Warn("failed to write ping", slog.Any("error", writeErr))
					return
				}

				flusher.Flush()
				continue
			}

			log.Error("xread error", slog.Any("error", err))
			continue
		}

		lastID := ""

		for _, stream := range records {
			for _, msg := range stream.Messages {
				payloadRaw, ok := msg.Values["payload"].(string)
				if !ok {
					log.Warn("message payload has invalid type", slog.String("message_id", msg.ID))
					continue
				}

				var payload map[string]any

				if err = json.Unmarshal([]byte(payloadRaw), &payload); err != nil {
					log.Warn("failed to decode payload", slog.String("message_id", msg.ID), slog.Any("error", err))
					continue
				}

				data, err := json.Marshal(payload)

				if err != nil {
					log.Error("failed to encode payload", slog.String("message_id", msg.ID), slog.Any("error", err))

					continue
				}

				eventType, ok := msg.Values["type"].(string)

				if !ok || eventType == "" {
					eventType = "notification"
				}

				if _, err = fmt.Fprintf(w, "id: %s\nevent: %s\ndata: %s\n\n", msg.ID, eventType, data); err != nil {
					log.Warn("failed to write sse event", slog.String("message_id", msg.ID), slog.Any("error", err))
					return
				}

				flusher.Flush()
				lastID = msg.ID
			}
		}

		if lastID != "" {
			startID = lastID
		}
	}
}
