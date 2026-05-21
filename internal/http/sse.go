package httpapi

import (
	"context"
	"net/http"

	"video-ops-agent/internal/agent/events"

	"github.com/gin-gonic/gin"
)

type sseEventSink struct {
	ctx *gin.Context
}

func newSSEEventSink(ctx *gin.Context) *sseEventSink {
	return &sseEventSink{ctx: ctx}
}

func prepareSSE(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Status(http.StatusOK)
}

func (s *sseEventSink) Emit(ctx context.Context, event events.RuntimeEvent) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.ctx.SSEvent(event.Type, event)
	s.ctx.Writer.Flush()
	return nil
}

func runtimeErrorEvent(sessionID uint, skillID string, err error) events.RuntimeEvent {
	return events.RuntimeEvent{
		Type:      events.TypeError,
		SessionID: sessionID,
		SkillID:   skillID,
		Status:    "error",
		Error:     err.Error(),
	}
}
