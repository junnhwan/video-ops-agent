package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"video-ops-agent/internal/agent/contextbuilder"
	"video-ops-agent/internal/agent/guard"
	"video-ops-agent/internal/agent/llm"
	"video-ops-agent/internal/agent/tools"
	"video-ops-agent/internal/store"
)

type LLMClient interface {
	Chat(ctx context.Context, request llm.ChatRequest) (llm.ChatResponse, error)
}

type RuntimeConfig struct {
	MaxToolRounds   int
	MaxGuardRetries int
	TotalTimeout    time.Duration
}

type Dependencies struct {
	LLM            LLMClient
	ToolRegistry   *tools.Registry
	ToolExecutor   *tools.Executor
	ContextBuilder *contextbuilder.Builder
	Repositories   contextbuilder.Repositories
}

type Runtime struct {
	llm            LLMClient
	toolRegistry   *tools.Registry
	toolExecutor   *tools.Executor
	contextBuilder *contextbuilder.Builder
	repos          contextbuilder.Repositories
	config         RuntimeConfig
}

type RunRequest struct {
	SessionID        uint
	UserMessage      string
	RequiredEvidence []string
}

type RunResult struct {
	SessionID     uint
	FinalAnswer   string
	RoundCount    int
	ToolCallCount int
}

func NewRuntime(deps Dependencies, config RuntimeConfig) *Runtime {
	if config.MaxToolRounds <= 0 {
		config.MaxToolRounds = 6
	}
	if config.MaxGuardRetries <= 0 {
		config.MaxGuardRetries = 2
	}
	if config.TotalTimeout <= 0 {
		config.TotalTimeout = 30 * time.Second
	}
	return &Runtime{
		llm:            deps.LLM,
		toolRegistry:   deps.ToolRegistry,
		toolExecutor:   deps.ToolExecutor,
		contextBuilder: deps.ContextBuilder,
		repos:          deps.Repositories,
		config:         config,
	}
}

func (r *Runtime) Run(ctx context.Context, request RunRequest) (RunResult, error) {
	if err := r.validate(); err != nil {
		return RunResult{}, err
	}
	if request.SessionID == 0 {
		return RunResult{}, fmt.Errorf("session_id is required")
	}
	if strings.TrimSpace(request.UserMessage) == "" {
		return RunResult{}, fmt.Errorf("user_message is required")
	}

	runCtx, cancel := context.WithTimeout(ctx, r.config.TotalTimeout)
	defer cancel()

	userMessage, err := r.repos.Messages.Create(runCtx, store.CreateMessageInput{
		SessionID: request.SessionID,
		Role:      store.MessageRoleUser,
		Content:   request.UserMessage,
	})
	if err != nil {
		return RunResult{}, err
	}

	result := RunResult{SessionID: request.SessionID}
	requiredEvidence := request.RequiredEvidence
	if len(requiredEvidence) == 0 {
		requiredEvidence = guard.RequiredTools(guard.DetectScenario(request.UserMessage))
	}
	guardRetries := 0
	completeEvidenceRetries := 0
	guardInstruction := ""

	for {
		built, err := r.contextBuilder.Build(runCtx, contextbuilder.BuildRequest{
			SessionID:        request.SessionID,
			RequiredEvidence: requiredEvidence,
		})
		if err != nil {
			return RunResult{}, err
		}

		messages := appendRuntimeReminder(built.Messages)
		if guardInstruction != "" {
			messages = append(messages, llm.Message{Role: llm.RoleSystem, Content: guardInstruction})
		}

		result.RoundCount++
		chatResp, err := r.llm.Chat(runCtx, llm.ChatRequest{
			Messages: messages,
			Tools:    r.toolRegistry.Schemas(),
		})
		if err != nil {
			return RunResult{}, fmt.Errorf("llm chat: %w", err)
		}

		if len(chatResp.Message.ToolCalls) > 0 {
			if result.ToolCallCount >= r.config.MaxToolRounds {
				return RunResult{}, fmt.Errorf("max tool rounds reached: %d", r.config.MaxToolRounds)
			}
			check, err := r.checkEvidence(runCtx, request.SessionID, requiredEvidence)
			if err != nil {
				return RunResult{}, err
			}
			alreadyExecuted, err := r.allToolCallsAlreadySucceeded(runCtx, request.SessionID, chatResp.Message.ToolCalls)
			if err != nil {
				return RunResult{}, err
			}
			if alreadyExecuted {
				if check.Complete {
					if completeEvidenceRetries >= r.config.MaxGuardRetries {
						return RunResult{}, fmt.Errorf("evidence complete but llm kept requesting duplicate tools")
					}
					completeEvidenceRetries++
					guardInstruction = "Evidence is complete. Do not call tools again. Produce the final answer using the existing tool evidence summaries."
					continue
				}
				if guardRetries >= r.config.MaxGuardRetries {
					return RunResult{}, fmt.Errorf("evidence incomplete after %d guard retries, missing tools: %s", r.config.MaxGuardRetries, strings.Join(check.MissingTools, ", "))
				}
				guardRetries++
				guardInstruction = guard.RetryInstruction(check.MissingTools)
				continue
			}
			for _, call := range chatResp.Message.ToolCalls {
				toolResult := r.executeToolCall(runCtx, request.SessionID, userMessage.ID, call)
				if toolResult.err != nil {
					return RunResult{}, toolResult.err
				}
				result.ToolCallCount++
			}
			guardInstruction = ""
			continue
		}

		finalAnswer := strings.TrimSpace(chatResp.Message.Content)
		if finalAnswer == "" {
			return RunResult{}, fmt.Errorf("llm returned empty final answer")
		}
		check, err := r.checkEvidence(runCtx, request.SessionID, requiredEvidence)
		if err != nil {
			return RunResult{}, err
		}
		if !check.Complete {
			if guardRetries >= r.config.MaxGuardRetries {
				return RunResult{}, fmt.Errorf("evidence incomplete after %d guard retries, missing tools: %s", r.config.MaxGuardRetries, strings.Join(check.MissingTools, ", "))
			}
			guardRetries++
			guardInstruction = guard.RetryInstruction(check.MissingTools)
			continue
		}
		if _, err := r.repos.Messages.Create(runCtx, store.CreateMessageInput{
			SessionID: request.SessionID,
			Role:      store.MessageRoleAssistant,
			Content:   finalAnswer,
		}); err != nil {
			return RunResult{}, err
		}
		result.FinalAnswer = finalAnswer
		return result, nil
	}
}

func (r *Runtime) validate() error {
	if r.llm == nil {
		return fmt.Errorf("llm client is required")
	}
	if r.toolRegistry == nil {
		return fmt.Errorf("tool registry is required")
	}
	if r.toolExecutor == nil {
		return fmt.Errorf("tool executor is required")
	}
	if r.contextBuilder == nil {
		return fmt.Errorf("context builder is required")
	}
	if r.repos.Messages == nil || r.repos.ToolCalls == nil {
		return fmt.Errorf("runtime repositories are required")
	}
	return nil
}

type toolCallExecution struct {
	err error
}

func (r *Runtime) executeToolCall(ctx context.Context, sessionID uint, messageID uint, call llm.ToolCall) toolCallExecution {
	started := time.Now()
	result, err := r.toolExecutor.Execute(ctx, call.Function.Name, call.Function.Arguments)
	latencyMS := time.Since(started).Milliseconds()

	messageIDPtr := messageID
	input := store.CreateToolCallInput{
		SessionID:     sessionID,
		MessageID:     &messageIDPtr,
		ToolName:      call.Function.Name,
		ArgumentsJSON: string(call.Function.Arguments),
		LatencyMS:     latencyMS,
	}

	if err != nil {
		input.Status = statusForToolError(err)
		input.ErrorMessage = err.Error()
	} else {
		input.Status = store.ToolCallStatusSuccess
		input.ResultSummary = result.Summary
		if result.Data != nil {
			encoded, marshalErr := json.Marshal(result.Data)
			if marshalErr != nil {
				input.Status = store.ToolCallStatusError
				input.ErrorMessage = fmt.Sprintf("marshal tool result: %v", marshalErr)
			} else {
				input.ResultJSON = string(encoded)
			}
		}
	}

	if _, createErr := r.repos.ToolCalls.Create(ctx, input); createErr != nil {
		return toolCallExecution{err: createErr}
	}
	return toolCallExecution{}
}

func statusForToolError(err error) string {
	if strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "deadline") {
		return store.ToolCallStatusTimeout
	}
	return store.ToolCallStatusError
}

func (r *Runtime) checkEvidence(ctx context.Context, sessionID uint, requiredTools []string) (guard.EvidenceCheck, error) {
	if len(requiredTools) == 0 {
		return guard.CheckRequired(nil, nil), nil
	}
	toolCalls, err := r.repos.ToolCalls.ListBySession(ctx, sessionID)
	if err != nil {
		return guard.EvidenceCheck{}, err
	}
	calledTools := make([]string, 0, len(toolCalls))
	for _, call := range toolCalls {
		if call.Status == store.ToolCallStatusSuccess {
			calledTools = append(calledTools, call.ToolName)
		}
	}
	return guard.CheckRequired(requiredTools, calledTools), nil
}

func (r *Runtime) allToolCallsAlreadySucceeded(ctx context.Context, sessionID uint, calls []llm.ToolCall) (bool, error) {
	if len(calls) == 0 {
		return false, nil
	}
	toolCalls, err := r.repos.ToolCalls.ListBySession(ctx, sessionID)
	if err != nil {
		return false, err
	}
	successful := make(map[string]struct{}, len(toolCalls))
	for _, call := range toolCalls {
		if call.Status != store.ToolCallStatusSuccess {
			continue
		}
		successful[toolCallKey(call.ToolName, json.RawMessage(call.ArgumentsJSON))] = struct{}{}
	}
	for _, call := range calls {
		if _, ok := successful[toolCallKey(call.Function.Name, call.Function.Arguments)]; !ok {
			return false, nil
		}
	}
	return true, nil
}

func toolCallKey(name string, arguments json.RawMessage) string {
	return strings.TrimSpace(name) + "\x00" + canonicalToolArguments(arguments)
}

func canonicalToolArguments(arguments json.RawMessage) string {
	trimmed := strings.TrimSpace(string(arguments))
	if trimmed == "" {
		return "{}"
	}
	var value any
	if err := json.Unmarshal(arguments, &value); err != nil {
		return trimmed
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return trimmed
	}
	return string(encoded)
}
