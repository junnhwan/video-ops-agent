# Local Smoke Commands

PowerShell commands for verifying `video-ops-agent` against a local `video-feed` and an OpenAI-compatible LLM.

## 1. Start video-feed dependencies

Run from `D:\dev\my_proj\go\video-feed`:

```powershell
docker compose up -d mysql redis rabbitmq
docker compose ps
```

Expected: `mysql`, `redis`, and `rabbitmq` are `healthy`.

## 2. Start video-feed API

Run from `D:\dev\my_proj\go\video-feed`:

```powershell
$env:CONFIG_PATH = "configs/config.compose-local.yaml"
go run ./cmd/server
```

Health check:

```powershell
Invoke-RestMethod http://127.0.0.1:8080/health
```

Expected:

```json
{"status":"ok"}
```

## 3. Prepare local video-ops-agent config

Create ignored `configs/config.yaml` in `D:\dev\my_proj\go\video-ops-agent`:

```yaml
server:
  address: "127.0.0.1:8090"

database:
  dsn: "data/video-ops-agent.db"

llm:
  base_url: "http://127.0.0.1:8317/v1"
  model: "gpt-5.4-mini"
  api_key_env: "VIDEO_OPS_LOCAL_LLM_API_KEY"

video_feed:
  base_url: "http://127.0.0.1:8080"
```

Set the API key only in the current shell:

```powershell
$env:VIDEO_OPS_LOCAL_LLM_API_KEY = "<your-local-llm-key>"
```

## 4. Start video-ops-agent

Run from `D:\dev\my_proj\go\video-ops-agent`:

```powershell
$env:CONFIG_PATH = "configs/config.yaml"
go run ./cmd/server
```

Health check:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/health
```

Expected:

```json
{"status":"ok"}
```

## 5. Create an Agent session

```powershell
$session = Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions" `
  -ContentType "application/json" `
  -Body (@{
    user_id = "local-smoke"
    title = "Local smoke"
    scenario = "integration_smoke"
  } | ConvertTo-Json)

$session.session.id
```

## 6. Ask MVP scenario questions

Hot rank attribution:

```powershell
$sid = $session.session.id
$body = @{
  content = "请分析 video_id=101 为什么会上热榜。请必须调用 get_video_detail、get_hot_videos、get_video_comments 后再回答，不要编造指标。"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages" `
  -ContentType "application/json" `
  -Body $body `
  -TimeoutSec 150
```

Comment risk:

```powershell
$body = @{
  content = "请分析 video_id=101 的评论风险。请必须调用 get_video_detail、get_video_comments、analyze_comment_risk 后再回答，不要编造没有提供的风险指标。"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages" `
  -ContentType "application/json" `
  -Body $body `
  -TimeoutSec 150
```

Author profile:

```powershell
$body = @{
  content = "作者 12 最近表现怎么样，值得扶持吗？请必须调用 get_author_profile 和 list_author_videos 后再回答。"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages" `
  -ContentType "application/json" `
  -Body $body `
  -TimeoutSec 150
```

Tag trend:

```powershell
$body = @{
  content = "#feed 这个标签最近内容表现怎么样？请必须调用 list_tag_videos 后再回答。"
} | ConvertTo-Json

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages" `
  -ContentType "application/json" `
  -Body $body `
  -TimeoutSec 150
```

## 7. Check trace persistence

```powershell
Invoke-RestMethod "http://127.0.0.1:8090/agent/sessions/$sid/tool-calls" |
  ConvertTo-Json -Depth 8
```

Expected successful tool names across the MVP smoke:

```text
get_video_detail
get_hot_videos
get_video_comments
analyze_comment_risk
get_author_profile
list_author_videos
list_tag_videos
```

Do not turn these smoke results into resume metrics. Phase 11 metrics must come from a separate measured evaluation run.
