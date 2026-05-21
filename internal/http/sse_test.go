package httpapi

import (
	"context"
	"errors"
	"testing"

	"video-ops-agent/internal/agent/events"
)

func TestSSEEventSinkReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := (&sseEventSink{}).Emit(ctx, events.RuntimeEvent{Type: events.TypeAgentStart, SessionID: 1})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Emit error = %v, want context.Canceled", err)
	}
}
