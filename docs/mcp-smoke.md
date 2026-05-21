# MCP Smoke

Authentication: none in local stdio v1. Keep API keys in `CONFIG_PATH`/environment only.

Local stdio entrypoint:

```powershell
$env:CONFIG_PATH = "configs/config.yaml"
go run ./cmd/mcp-server
```

The server speaks JSON-RPC 2.0 over stdio with `Content-Length` frames. It exposes:

- Tools: read-only VideoOps tools through the existing Tool Gateway.
- Prompts: Diagnosis Skills rendered as prompt templates.
- Resources: `videoops://tools`, `videoops://skills`, `videoops://evidence-rules`, and `videoops://sessions/{id}/trace`.

Tool calls are recorded in `gateway_tool_invocations` with `source=mcp_client`.

## JSON-RPC Methods

Transport: stdio with `Content-Length` frames.

Error shape:

```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"message"}}
```

### initialize

Request:

```json
{"jsonrpc":"2.0","id":1,"method":"initialize"}
```

Response result includes `serverInfo` and capabilities for `tools`, `resources`, and `prompts`.

### tools/list

Request:

```json
{"jsonrpc":"2.0","id":2,"method":"tools/list"}
```

Response:

```json
{"jsonrpc":"2.0","id":2,"result":{"tools":[]}}
```

### tools/call

Request:

```json
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_video_detail","arguments":{"video_id":101}}}
```

Response result:

```json
{"content":[{"type":"text","text":"..."}],"structuredContent":{"tool_name":"get_video_detail","summary":"...","data":{}}}
```

### resources/list and resources/read

Resources:

- `videoops://tools`
- `videoops://skills`
- `videoops://evidence-rules`
- `videoops://sessions/{id}/trace`

### prompts/list and prompts/get

Diagnosis Skills are exposed as prompt names, for example `comment_risk_analysis`.
