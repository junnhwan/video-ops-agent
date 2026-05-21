# MCP Smoke

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
