package httplog

import (
	"github.com/rockkley/pushpost/services/common/ctxlog"
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func Logger(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqLog := base.With(
				slog.String("request_id", chimiddleware.GetReqID(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
			)

			ctx := ctxlog.With(r.Context(), reqLog)
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r.WithContext(ctx))

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			level := slog.LevelInfo
			switch {
			case status >= 500:
				level = slog.LevelError
			case status >= 400:
				level = slog.LevelWarn
			}

			reqLog.Log(ctx, level, "request completed",
				slog.Int("status", status),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
