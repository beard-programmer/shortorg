package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

type AppLogger = slog.Logger

func NewLogger() (*AppLogger, error) {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{TimeFormat: time.DateTime}))

	return logger, nil
}
