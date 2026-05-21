package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Server struct {
	tools     *ToolAdapter
	resources *ResourceAdapter
	prompts   *PromptAdapter
}

func NewServer(toolAdapter *ToolAdapter, resourceAdapter *ResourceAdapter, promptAdapter *PromptAdapter) *Server {
	return &Server{tools: toolAdapter, resources: resourceAdapter, prompts: promptAdapter}
}

func (s *Server) Handle(ctx context.Context, request JSONRPCRequest) JSONRPCResponse {
	result, err := s.handle(ctx, request)
	if err != nil {
		return JSONRPCResponse{JSONRPC: "2.0", ID: request.ID, Error: &RPCError{Code: -32000, Message: err.Error()}}
	}
	return JSONRPCResponse{JSONRPC: "2.0", ID: request.ID, Result: result}
}

func (s *Server) handle(ctx context.Context, request JSONRPCRequest) (any, error) {
	switch request.Method {
	case "initialize":
		return map[string]any{
			"protocolVersion": "2024-11-05",
			"serverInfo":      map[string]any{"name": "videoops-agent", "version": "0.1.0"},
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
				"prompts":   map[string]any{},
			},
		}, nil
	case "tools/list":
		tools, err := s.tools.ListTools(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]any{"tools": tools}, nil
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return nil, fmt.Errorf("decode tools/call params: %w", err)
		}
		result, err := s.tools.CallTool(ctx, params.Name, params.Arguments)
		if err != nil {
			return nil, err
		}
		return result, nil
	case "resources/list":
		resources, err := s.resources.ListResources(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]any{"resources": resources}, nil
	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return nil, fmt.Errorf("decode resources/read params: %w", err)
		}
		content, err := s.resources.ReadResource(ctx, params.URI)
		if err != nil {
			return nil, err
		}
		return map[string]any{"contents": []map[string]any{{
			"uri":      params.URI,
			"mimeType": "application/json",
			"text":     string(content),
		}}}, nil
	case "prompts/list":
		prompts, err := s.prompts.ListPrompts(ctx)
		if err != nil {
			return nil, err
		}
		return map[string]any{"prompts": prompts}, nil
	case "prompts/get":
		var params struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			return nil, fmt.Errorf("decode prompts/get params: %w", err)
		}
		return s.prompts.GetPrompt(ctx, params.Name)
	default:
		return nil, fmt.Errorf("unsupported method %q", request.Method)
	}
}

func (s *Server) Serve(ctx context.Context, reader io.Reader, writer io.Writer) error {
	buffered := bufio.NewReader(reader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		payload, err := readFrame(buffered)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		var request JSONRPCRequest
		if err := json.Unmarshal(payload, &request); err != nil {
			return fmt.Errorf("decode json-rpc request: %w", err)
		}
		if request.ID == nil {
			continue
		}
		response := s.Handle(ctx, request)
		if err := writeFrame(writer, response); err != nil {
			return err
		}
	}
}

func readFrame(reader *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length %q", value)
			}
			contentLength = parsed
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeFrame(writer io.Writer, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal json-rpc response: %w", err)
	}
	var frame bytes.Buffer
	fmt.Fprintf(&frame, "Content-Length: %d\r\n\r\n", len(payload))
	frame.Write(payload)
	_, err = writer.Write(frame.Bytes())
	return err
}
