# Console Backend API Contract

Base URL for local smoke: `http://127.0.0.1:8090`.

Authentication: none in local v1. Do not send API keys to backend endpoints.

Error shape:

```json
{"error":"message"}
```

## Health

`GET /health`

Response:

```json
{"status":"ok"}
```

Smoke:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/health
```

## Agent Sessions

`POST /agent/sessions`

Request:

```json
{
  "user_id": "operator-1",
  "title": "Skill smoke",
  "scenario": "comment_risk_analysis",
  "skill_id": "comment_risk_analysis",
  "skill_version": "1.0.0",
  "context_policy": {"max_recent_messages": 6}
}
```

Response:

```json
{"session":{"id":1,"user_id":"operator-1","skill_id":"comment_risk_analysis","status":"active"}}
```

`GET /agent/sessions?user_id=operator-1&limit=20`

Response:

```json
{"sessions":[]}
```

`GET /agent/sessions/{id}`

Response:

```json
{"session":{},"messages":[]}
```

## Blocking Agent Message

`POST /agent/sessions/{id}/messages`

Request:

```json
{
  "content": "请分析 video_id=101 的评论风险",
  "skill_id": "comment_risk_analysis",
  "required_evidence": []
}
```

Response:

```json
{
  "session_id": 1,
  "final_answer": "...",
  "round_count": 3,
  "tool_call_count": 2
}
```

## Agent Trace

`GET /agent/sessions/{id}/tool-calls`

Response:

```json
{"tool_calls":[]}
```

Smoke:

```powershell
$session = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8090/agent/sessions" -ContentType "application/json" -Body (@{
  user_id = "local-smoke"
  title = "Skill smoke"
  scenario = "comment_risk_analysis"
  skill_id = "comment_risk_analysis"
} | ConvertTo-Json)

Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:8090/agent/sessions/$($session.session.id)/messages" -ContentType "application/json" -Body (@{
  content = "请分析 video_id=101 的评论风险"
} | ConvertTo-Json)
```
