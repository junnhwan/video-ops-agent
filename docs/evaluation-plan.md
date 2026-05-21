# Evaluation Plan

Goal: measure whether Diagnosis Skills plus Evidence Guard improve evidence completeness and reduce unsupported final answers.

## Modes

- `baseline`: Function Calling plus Tool Registry without selected Diagnosis Skill.
- `skill_guard`: Diagnosis Skills plus tool whitelist and required evidence guard.

## Endpoints

`GET /eval/summary`

Returns aggregate metrics from persisted sessions and `gateway_tool_invocations`.

`GET /eval/skills/{id}/summary`

Returns the same aggregate metrics filtered by `skill_id`.

`POST /eval/runs`

Request:

```json
{"mode":"skill_guard","skill_id":"comment_risk_analysis"}
```

Response:

```json
{"run":{"id":1,"mode":"skill_guard","summary":{}}}
```

`GET /eval/runs/{id}`

Returns the in-process run snapshot created by `POST /eval/runs`.

## Metrics From Persisted Trace

- `tool_call_success_rate`
- `tool_call_error_count`
- `unauthorized_tool_call_count`
- `evidence_complete_final_answer_count`
- `average_tool_latency_ms`
- `average_tool_call_count`
- `skill_success_count`
- `skill_failure_count`

## Explicitly Unsupported Until Persisted

These fields are returned as `null` and listed in `unsupported_metrics` until runtime events or run results are persisted:

- `guard_retry_count`
- `evidence_incomplete_final_answer_rejected_count`
- `average_round_count`

Do not use unsupported metrics in resume or benchmark claims.

## Manual Prompt Set

Run each prompt once in `baseline` mode and once with `skill_id=comment_risk_analysis`:

```text
请分析 video_id=101 的评论风险，并给出证据链。
请复盘 video_id=101 的内容表现、评论反馈和运营建议。
请判断 video_id=101 是否存在评论区刷屏或敏感词风险。
```

After each run:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/eval/summary | ConvertTo-Json -Depth 8
Invoke-RestMethod http://127.0.0.1:8090/eval/skills/comment_risk_analysis/summary | ConvertTo-Json -Depth 8
```
