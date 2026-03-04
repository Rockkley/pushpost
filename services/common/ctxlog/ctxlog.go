package ctxlog

import (
	"context"
	"log/slog"
)

type key struct{}

func With(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, key{}, l)
}

func From(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(key{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
