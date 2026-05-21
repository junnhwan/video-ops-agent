package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Executor struct {
	registry       *Registry
	defaultTimeout time.Duration
}

func NewExecutor(registry *Registry, defaultTimeout time.Duration) *Executor {
	if defaultTimeout <= 0 {
		defaultTimeout = 2 * time.Second
	}
	return &Executor{registry: registry, defaultTimeout: defaultTimeout}
}

func (e *Executor) Execute(ctx context.Context, toolName string, arguments json.RawMessage) (ToolResult, error) {
	tool, ok := e.registry.Get(toolName)
	if !ok {
		return ToolResult{}, fmt.Errorf("unknown tool %q", toolName)
	}

	timeout := tool.Timeout()
	if timeout <= 0 {
		timeout = e.defaultTimeout
	}

	toolCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type outcome struct {
		result ToolResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() {
		result, err := tool.Execute(toolCtx, arguments)
		done <- outcome{result: result, err: err}
	}()

	select {
	case <-toolCtx.Done():
		return ToolResult{}, fmt.Errorf("tool %q timeout after %s: %w", toolName, timeout, toolCtx.Err())
	case outcome := <-done:
		if outcome.err != nil {
			return ToolResult{}, fmt.Errorf("execute tool %q: %w", toolName, outcome.err)
		}
		if outcome.result.ToolName == "" {
			outcome.result.ToolName = toolName
		}
		return outcome.result, nil
	}
}
