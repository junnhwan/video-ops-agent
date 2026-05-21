# Tool Gateway Contract

Base URL: `http://127.0.0.1:8090`.

Authentication: none in local v1.

Error shape:

```json
{"error":"message"}
```

## List Tools

`GET /gateway/tools`

Response:

```json
{
  "tools": [
    {
      "name": "get_video_detail",
      "display_name": "视频详情",
      "category": "video",
      "description": "Get one video detail from video-feed by video_id.",
      "read_only": true,
      "schema": {"type":"function","function":{"name":"get_video_detail","parameters":{}}}
    }
  ]
}
```

## Get Tool

`GET /gateway/tools/{name}`

Response:

```json
{"tool":{"name":"get_video_detail","read_only":true}}
```

## Call Tool

`POST /gateway/tools/{name}/call`

Request:

```json
{
  "source": "manual_console",
  "session_id": 1,
  "skill_id": "comment_risk_analysis",
  "arguments": {"video_id": 101, "limit": 50}
}
```

Response:

```json
{
  "invocation": {
    "id": 1,
    "source": "manual_console",
    "tool_name": "get_video_detail",
    "status": "success",
    "latency_ms": 12,
    "result_summary": "video 101: test"
  },
  "result": {"tool_name":"get_video_detail","summary":"video 101: test","data":{}}
}
```

## Invocation Trace

`GET /gateway/invocations?source=manual_console&tool_name=get_video_detail&session_id=1&skill_id=comment_risk_analysis&status=success&limit=50`

Response:

```json
{"invocations":[]}
```

`GET /gateway/invocations/{id}`

Response:

```json
{"invocation":{}}
```

Smoke:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/gateway/tools | ConvertTo-Json -Depth 8

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8090/gateway/tools/analyze_video_comment_risk/call" -ContentType "application/json" -Body (@{
  source = "manual_console"
  arguments = @{ video_id = 101; limit = 50 }
} | ConvertTo-Json -Depth 8) | ConvertTo-Json -Depth 10
```
