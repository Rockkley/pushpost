package logger

import (
	"github.com/rockkley/pushpost/services/common/logger/handlers/slogpretty"
	"log/slog"
	"os"
)

func SetupLogger(env string) *slog.Logger {
	switch env {
	case "prod":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	case "development":
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	default:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}
		return slog.New(opts.NewPrettyHandler(os.Stdout))
	}
}
