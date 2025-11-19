package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func GetLogger(env string) *slog.Logger {
	switch env {
	case "local":
		return slog.New(
			tint.NewHandler(
				os.Stdout,
				&tint.Options{
					Level: slog.LevelDebug,
				},
			),
		)
	case "prod":
		return slog.New(
			slog.NewJSONHandler(
				os.Stdout,
				&slog.HandlerOptions{
					Level: slog.LevelInfo,
				},
			),
		)
	}
	return nil
}
