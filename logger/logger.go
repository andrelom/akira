package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	logger *log.Logger
}

func New(output io.Writer) *Logger {
	return &Logger{
		logger: log.New(output, "", log.LstdFlags),
	}
}

func NewFile(filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return New(io.MultiWriter(os.Stdout, file)), nil
}

func (lgr *Logger) Info(format string, args ...any) {
	lgr.logger.Printf("[INFO] "+format, args...)
}

func (lgr *Logger) Warn(format string, args ...any) {
	lgr.logger.Printf("[WARN] "+format, args...)
}

func (lgr *Logger) Error(format string, args ...any) {
	lgr.logger.Printf("[ERROR] "+format, args...)
}
