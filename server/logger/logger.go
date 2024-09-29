package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

func InitLogger(filename string) (*slog.Logger, *os.File) {
	logFile, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create log file %s: %v", filename, err)
	}

	handler := slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, logFile),
		&slog.HandlerOptions{Level: slog.LevelInfo},
	)
	return slog.New(handler), logFile
}
