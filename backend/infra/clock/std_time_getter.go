package clock

import (
	"context"
	"time"
)

// StdTimeGetter implements application/usecase/thread.ITimeGetter using the standard library clock.
type StdTimeGetter struct{}

func NewStdTimeGetter() *StdTimeGetter {
	return &StdTimeGetter{}
}

func (g *StdTimeGetter) Now(ctx context.Context) time.Time {
	return time.Now()
}
