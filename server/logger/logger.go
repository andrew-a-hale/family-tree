package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

func InitLogger(filename string) (*slog.Logger, *os.File) {
	logFile, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file %s: %v", filename, err)
	}

	handler := slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, logFile),
		&slog.HandlerOptions{Level: slog.LevelInfo},
	)
	return slog.New(handler), logFile
}
