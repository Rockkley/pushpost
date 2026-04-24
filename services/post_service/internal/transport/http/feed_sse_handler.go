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
	"github.com/redis/go-redis/v9"
	"github.com/rockkley/pushpost/services/common_service/ctxlog"
	commonmiddleware "github.com/rockkley/pushpost/services/common_service/middleware"
	realtimepkg "github.com/rockkley/pushpost/services/post_service/internal/realtime"
)

type FeedSSEHandler struct {
	rdb *redis.Client
}

func NewFeedSSEHandler(rdb *redis.Client) *FeedSSEHandler {
	return &FeedSSEHandler{rdb: rdb}
}

// Subscribe — SSE endpoint для получения событий ленты.
func (h *FeedSSEHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	log := ctxlog.From(r.Context()).With(slog.String("op", "FeedSSEHandler.Subscribe"))

	userID, ok := commonmiddleware.UserIDFromContext(r.Context())
	if !ok || userID == uuid.Nil {
		http.Error(w, `{"code":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Проверка поддержки SSE у клиента
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// SSE заголовки
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// Определяем точку восстановления
	startID := r.Header.Get("Last-Event-ID")
	if startID == "" {
		startID = "$" // только новые события
	}

	streamKey := realtimepkg.StreamKey(userID)

	log.Info("SSE client connected",
		slog.String("user_id", userID.String()),
		slog.String("start_id", startID),
	)
	defer log.Info("SSE client disconnected", slog.String("user_id", userID.String()))

	for {
		// Чтение событий из Redis Stream с блокировкой на 25 секунд
		cmd := h.rdb.XRead(r.Context(), &redis.XReadArgs{
			Streams: []string{streamKey, startID},
			Count:   50,
			Block:   25 * time.Second,
		})

		records, err := cmd.Result()
		if err != nil {
			// Нормальное завершение запроса
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}

			// Таймаут ожидания — просто отправляем heartbeat
			if errors.Is(err, redis.Nil) {
				fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
				flusher.Flush()
				continue
			}

			// Логируем, но не роняем SSE
			log.Error("xread error", slog.Any("error", err))
			continue
		}

		// Если новых событий нет — heartbeat
		if len(records) == 0 {
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
			continue
		}

		// Последний ID обновляется *после* обработки всех сообщений (важно для burst и консистентности)
		var lastID string

		for _, stream := range records {
			for _, msg := range stream.Messages {

				payloadRaw, ok := msg.Values["payload"].(string)
				if !ok {
					continue
				}

				var event realtimepkg.FeedEvent
				if err = json.Unmarshal([]byte(payloadRaw), &event); err != nil {
					continue
				}

				// Упаковка события
				data, _ := json.Marshal(event)

				// Отправка SSE
				fmt.Fprintf(w,
					"id: %s\nevent: %s\ndata: %s\n\n",
					msg.ID,
					event.Type,
					data,
				)
				flusher.Flush()
				lastID = msg.ID
			}
		}

		// Обновляем cursor только один раз в самом конце
		if lastID != "" {
			startID = lastID
		}
	}
}
