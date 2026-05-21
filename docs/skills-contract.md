# Diagnosis Skills Contract

Base URL: `http://127.0.0.1:8090`.

Authentication: none in local v1.

Error shape:

```json
{"error":"message"}
```

Skill status values: `enabled`, `disabled`.

Built-in skill IDs:

- `hot_rank_attribution`
- `comment_risk_analysis`
- `author_support_evaluation`
- `tag_trend_analysis`
- `content_review_summary`

## List Skills

`GET /skills`

Response:

```json
{"skills":[{"id":"comment_risk_analysis","status":"enabled","allowed_tools":[],"required_evidence":[]}]}
```

## Get Skill

`GET /skills/{id}`

Response:

```json
{"skill":{"id":"comment_risk_analysis","version":"1.0.0"}}
```

## Create Skill

`POST /skills`

Request:

```json
{
  "id": "custom_comment_review",
  "name": "自定义评论复盘",
  "description": "custom skill",
  "version": "1.0.0",
  "status": "enabled",
  "scenario": "comment_risk_analysis",
  "allowed_tools": ["get_video_detail", "analyze_video_comment_risk"],
  "required_evidence": ["get_video_detail"],
  "prompt_template": "Use evidence.",
  "output_sections": ["结论", "证据"]
}
```

Response:

```json
{"skill":{"id":"custom_comment_review","status":"enabled"}}
```

## Update Skill

`PUT /skills/{id}`

Request body matches create body. The path `id` is authoritative.

Response:

```json
{"skill":{"id":"custom_comment_review","version":"1.0.1"}}
```

## Enable Or Disable

`POST /skills/{id}/enable`

`POST /skills/{id}/disable`

Response:

```json
{"skill":{"id":"custom_comment_review","status":"disabled"}}
```

Smoke:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/skills | ConvertTo-Json -Depth 8
Invoke-RestMethod http://127.0.0.1:8090/skills/comment_risk_analysis | ConvertTo-Json -Depth 8
```
