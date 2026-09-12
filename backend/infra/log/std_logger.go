package log

import (
	"context"
	"log"
)

// StdLogger implements application/usecase/thread.ILogger using the standard library logger.
type StdLogger struct{}

func NewStdLogger() *StdLogger {
	return &StdLogger{}
}

func (l *StdLogger) Error(ctx context.Context, err error) {
	log.Printf("ERROR: %v", err)
}
