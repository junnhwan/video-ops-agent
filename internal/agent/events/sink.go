package events

import "context"

type EventSink interface {
	Emit(ctx context.Context, event RuntimeEvent) error
}

type noopSink struct{}

func NoopSink() EventSink {
	return noopSink{}
}

func (noopSink) Emit(context.Context, RuntimeEvent) error {
	return nil
}
