# VideoOps Agent

`video-ops-agent` 是一个面向短视频内容运营诊断的 Go 后端服务和 React 控制台。它把 OpenAI 兼容的 Chat Completions 模型接入到只读的 `video-feed` 平台工具中，通过工具证据、会话上下文、Evidence Guard 和 Diagnosis Skills 生成可追溯的运营分析结果。

当前项目已经不只是一个 Agent Chat API，而是一个完整的 VideoOps Agent Console 原型，包含：

- Agent 会话和基于工具证据的多轮运行时。
- Diagnosis Skills：定义工具白名单、必需证据、提示词风格和报告结构。
- Tool Gateway：工具目录、手动调用、统一调用轨迹。
- SSE 运行事件：前端可以展示 Agent 执行时间线。
- MCP stdio 适配器：外部 AI 客户端可以复用工具、资源和提示词。
- Evaluation Metrics：从持久化会话和工具调用轨迹推导评估指标。
- `web/` 下的 React/Vite 前端控制台。

## 目录结构

```text
cmd/server/             # HTTP API 服务入口
cmd/mcp-server/         # 本地 stdio MCP 服务入口
configs/                # 示例配置；本地配置不提交
docs/                   # 设计、接口合同、smoke 命令、评估说明
internal/agent/         # LLM runtime、工具、上下文、证据守卫、Skills、事件
internal/eval/          # 评估指标聚合和 HTTP handler
internal/gateway/       # Tool Gateway 目录、调用、轨迹服务
internal/http/          # Gin router 和 HTTP handler
internal/mcp/           # 最小 MCP JSON-RPC 适配器
internal/platform/      # video-feed HTTP client
internal/store/         # SQLite 模型和 repository
web/                    # React/Vite 前端控制台
```

## 当前后端能力

### Agent Runtime

核心接口：

- `POST /agent/sessions`：创建 Agent 会话。
- `GET /agent/sessions`：查询会话列表。
- `GET /agent/sessions/{id}`：查询会话详情和消息。
- `POST /agent/sessions/{id}/messages`：阻塞式运行 Agent。
- `POST /agent/sessions/{id}/messages/stream`：通过 SSE 流式返回执行过程。
- `GET /agent/sessions/{id}/tool-calls`：查询该会话的工具调用轨迹。

当 session 或 message 带有 `skill_id` 时，Runtime 会：

- 加载对应 Diagnosis Skill。
- 只把 Skill 允许的工具 schema 暴露给 LLM。
- 优先使用 Skill 的 `required_evidence` 做证据校验。
- 把 Skill prompt 和输出结构注入 system context。
- 在工具调用轨迹里记录 `skill_id` 和 `skill_version`。

### Tool Gateway

核心接口：

- `GET /gateway/tools`
- `GET /gateway/tools/{name}`
- `POST /gateway/tools/{name}/call`
- `GET /gateway/invocations`
- `GET /gateway/invocations/{id}`

调用来源 `source`：

- `manual_console`：前端或人工手动调用。
- `agent_runtime`：Agent Runtime 自动调用。
- `mcp_client`：MCP 客户端调用。

### Diagnosis Skills

内置 Skills：

- `hot_rank_attribution`：热榜归因分析。
- `comment_risk_analysis`：评论风险分析。
- `author_support_evaluation`：作者扶持评估。
- `tag_trend_analysis`：标签趋势分析。
- `content_review_summary`：内容复盘摘要。

Skill 接口：

- `GET /skills`
- `GET /skills/{id}`
- `POST /skills`
- `PUT /skills/{id}`
- `POST /skills/{id}/enable`
- `POST /skills/{id}/disable`

当前 Skill 是结构化元数据和 prompt/report/evidence 规则，不是可执行脚本。

### SSE 执行事件

流式接口：

```text
POST /agent/sessions/{id}/messages/stream
```

事件类型：

- `agent_start`
- `skill_loaded`
- `llm_round_start`
- `tool_call`
- `tool_result`
- `guard_retry`
- `final_answer`
- `error`

前端可以基于这些事件展示 Agent 执行时间线、工具调用过程、证据守卫重试和最终回答。

### MCP Adapter

`cmd/mcp-server` 提供本地 stdio JSON-RPC MCP 适配器，暴露：

- MCP Tools：复用只读 Tool Gateway 工具。
- MCP Prompts：由 Diagnosis Skills 渲染。
- MCP Resources：
  - `videoops://tools`
  - `videoops://skills`
  - `videoops://evidence-rules`
  - `videoops://sessions/{id}/trace`

MCP 调用会写入 `gateway_tool_invocations`，并标记 `source=mcp_client`。

注意：这里的 MCP 不是通用 HTTP/API 网关，只是 VideoOps 领域能力适配器。

### Evaluation Metrics

评估接口：

- `GET /eval/summary`
- `GET /eval/skills/{id}/summary`
- `POST /eval/runs`
- `GET /eval/runs/{id}`

当前指标来自已持久化的 session 和 `gateway_tool_invocations`。没有持久化依据的指标不会编造，会以 `unsupported_metrics` 明确说明。

## 本地依赖

- Go `1.26.1`
- Node.js / npm
- 本地 `video-feed` API：默认 `http://127.0.0.1:8080`
- OpenAI 兼容 Chat Completions 服务
- 以下命令默认使用 Windows PowerShell

## 后端启动

先创建不会提交的本地配置 `configs/config.yaml`：

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

只在当前 shell 里设置 API key：

```powershell
$env:VIDEO_OPS_LOCAL_LLM_API_KEY = "<your-local-llm-key>"
$env:CONFIG_PATH = "configs/config.yaml"
go run ./cmd/server
```

健康检查：

```powershell
Invoke-RestMethod http://127.0.0.1:8090/health
```

期望结果：

```json
{"status":"ok"}
```

## 前端启动

前端在 `web/` 目录，Vite dev server 会把 `/api` 代理到后端 `http://127.0.0.1:8090`。

```powershell
cd web
npm install
npm run dev
```

默认地址：

```text
http://127.0.0.1:3000
```

常用前端环境变量：

```text
VITE_API_BASE_URL=/api
VITE_USE_MOCK=false
```

## MCP 启动

从仓库根目录运行：

```powershell
$env:CONFIG_PATH = "configs/config.yaml"
go run ./cmd/mcp-server
```

MCP server 使用 stdio + `Content-Length` JSON-RPC frame。

## 边界和已知限制

- 工具都是只读工具，不会发评论、改推荐、改内容状态或写入 `video-feed` 业务数据。
- Skills 是结构化规则和 prompt，不是脚本执行系统。
- MCP 是 VideoOps 领域适配器，不是通用 API 网关。
- Eval 当前只从已持久化数据推导指标，不声称模型准确率或证据守卫效果。
- `configs/config.yaml`、`data/`、`docs/study-notes/`、`web/node_modules/`、`web/dist/` 等本地产物不要提交。
