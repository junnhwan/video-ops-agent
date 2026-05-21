package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClientConfig struct {
	BaseURL string
	Model   string
	APIKey  string
	Timeout time.Duration
}

type Client struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

func NewClient(config ClientConfig) (*Client, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("llm base url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse llm base url: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("llm base url must include scheme and host")
	}
	if strings.HasSuffix(baseURL, "/v1") {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		return nil, fmt.Errorf("llm model is required")
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		model:   model,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	if len(request.Messages) == 0 {
		return ChatResponse{}, fmt.Errorf("messages are required")
	}

	payload := openAIChatRequest{
		Model:    c.model,
		Messages: request.Messages,
		Tools:    request.Tools,
	}
	if len(request.Tools) > 0 {
		payload.ToolChoice = "auto"
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("encode chat completions request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("build chat completions request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("call chat completions: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ChatResponse{}, chatAPIError(resp)
	}

	var providerResp openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return ChatResponse{}, fmt.Errorf("decode chat completions response: %w", err)
	}
	return providerResp.toChatResponse()
}

type openAIChatRequest struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	Tools      any       `json:"tools,omitempty"`
	ToolChoice string    `json:"tool_choice,omitempty"`
}

type openAIChatResponse struct {
	ID      string         `json:"id"`
	Choices []openAIChoice `json:"choices"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason"`
	Message      openAIMessage `json:"message"`
}

type openAIMessage struct {
	Role      string           `json:"role"`
	Content   *string          `json:"content"`
	ToolCalls []openAIToolCall `json:"tool_calls"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (r openAIChatResponse) toChatResponse() (ChatResponse, error) {
	if len(r.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("chat completions response has no choices")
	}

	choice := r.Choices[0]
	message := Message{Role: choice.Message.Role}
	if choice.Message.Content != nil {
		message.Content = *choice.Message.Content
	}

	if len(choice.Message.ToolCalls) > 0 {
		message.ToolCalls = make([]ToolCall, 0, len(choice.Message.ToolCalls))
		for _, rawCall := range choice.Message.ToolCalls {
			arguments := json.RawMessage(rawCall.Function.Arguments)
			if !json.Valid(arguments) {
				return ChatResponse{}, fmt.Errorf("tool arguments for %q are not valid JSON", rawCall.Function.Name)
			}
			message.ToolCalls = append(message.ToolCalls, ToolCall{
				ID:   rawCall.ID,
				Type: rawCall.Type,
				Function: FunctionCall{
					Name:      rawCall.Function.Name,
					Arguments: arguments,
				},
			})
		}
	}

	return ChatResponse{
		ID:           r.ID,
		Message:      message,
		FinishReason: choice.FinishReason,
	}, nil
}

func chatAPIError(resp *http.Response) error {
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if readErr != nil {
		return fmt.Errorf("chat completions returned status %d and unreadable error body: %w", resp.StatusCode, readErr)
	}

	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil && parsed.Error.Message != "" {
		return fmt.Errorf("chat completions returned status %d: %s", resp.StatusCode, parsed.Error.Message)
	}
	if len(body) == 0 {
		return fmt.Errorf("chat completions returned status %d", resp.StatusCode)
	}
	return fmt.Errorf("chat completions returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}
