# SSE Agent Events Contract

Base URL: `http://127.0.0.1:8090`.

Authentication: none in local v1.

Endpoint:

`POST /agent/sessions/{id}/messages/stream`

Headers:

```text
Accept: text/event-stream
Content-Type: application/json
```

Request:

```json
{
  "content": "请分析 video_id=101 的评论风险，并展示证据链。",
  "skill_id": "comment_risk_analysis",
  "required_evidence": []
}
```

Response content type: `text/event-stream`.

Event names:

- `agent_start`
- `skill_loaded`
- `llm_round_start`
- `tool_call`
- `tool_result`
- `guard_retry`
- `final_answer`
- `error`

Event data shape:

```json
{
  "type": "tool_result",
  "session_id": 1,
  "skill_id": "comment_risk_analysis",
  "tool_name": "analyze_video_comment_risk",
  "summary": "low comment risk for video 101",
  "status": "success",
  "round_count": 2,
  "tool_call_count": 1
}
```

Error event shape:

```json
{"type":"error","session_id":1,"skill_id":"comment_risk_analysis","status":"error","error":"message"}
```

Smoke:

```powershell
Invoke-WebRequest -Method Post -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages/stream" -ContentType "application/json" -Headers @{ Accept = "text/event-stream" } -Body (@{
  content = "请分析 video_id=101 的评论风险，并展示证据链。"
  skill_id = "comment_risk_analysis"
} | ConvertTo-Json)
```
