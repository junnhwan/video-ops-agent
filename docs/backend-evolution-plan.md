# VideoOps Agent Console Backend Evolution Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade `video-ops-agent` from a chat-only Agent backend into a backend with Tool Gateway governance, Diagnosis Skills, streamable Agent execution events, MCP exposure, and evaluation metrics.

**Architecture:** Keep the current Go Agent Runtime, Tool Registry, Context Builder, Trace persistence, and Evidence Guard as the core. Add new backend layers around them: a Tool Gateway API for tool catalog/playground/trace, a Diagnosis Skills domain for configurable analysis methods, an SSE event path for frontend observability, and a later MCP adapter that exposes existing read-only capabilities as MCP Tools, Resources, and Prompts. Do not copy the Java `ai-mcp-gateway` architecture into this repo; borrow its concepts only where they fit this domain.

**Tech Stack:** Go 1.26, Gin, GORM, SQLite, OpenAI-compatible Chat Completions, existing `internal/agent/tools`, local `video-feed` HTTP client, optional future MCP Go SDK or minimal protocol adapter.

---

## 0. Context And Boundaries

### Current Backend Baseline

The current backend already has:

- `internal/agent/tools`: unified `Tool`, `ToolSchema`, `ToolResult`, `Registry`, and `Executor`.
- `internal/platform/videofeed`: typed read-only HTTP client for `video-feed`.
- `internal/store`: SQLite persistence for sessions, messages, and tool calls.
- `internal/agent/contextbuilder`: compacts session context and tool evidence before LLM calls.
- `internal/agent/runtime`: multi-round tool calling runtime with max tool round limits.
- `internal/agent/guard`: scenario detection and required tool evidence checks.
- `internal/http`: synchronous JSON Agent Chat API.

Recent review fixes already added `analyze_video_comment_risk(video_id, limit)` so the backend fetches comments internally instead of requiring the LLM to repackage comments.

### Product Direction

Build the backend for:

```text
VideoOps Agent Console
  -> Diagnosis Skills: configurable operation-analysis methods
  -> Agent Runtime: multi-round evidence collection and report generation
  -> Tool Gateway: catalog, playground, source-aware trace
  -> SSE Events: observable Agent execution timeline
  -> MCP Adapter: external AI clients can reuse tools/prompts/resources
  -> video-feed: real platform data source
```

### What We Are Not Building

- No generic MCP gateway platform that maps arbitrary HTTP APIs.
- No direct integration of the Java `ai-mcp-gateway` service as a component.
- No write tools such as posting comments, sending notifications, changing recommendations, or mutating `video-feed` data.
- No executable script-based Skills in v1.
- No arbitrary file/network access from Skills.
- No claims about accuracy, latency, or guard effectiveness until Phase 15 runs measured evaluation.

### Design Principles

- Keep business value first: short-video operation analysis, evidence collection, and traceability.
- Reuse existing `tools.Registry` and `tools.Executor`; do not create a parallel tool execution system.
- Make every new layer testable without a real LLM.
- Keep frontend contract stable and documented.
- Commit by phase or focused backend slice; avoid a single giant commit.

---

## 1. Target Module Layout

Create or evolve these backend packages:

```text
internal/gateway/
  catalog.go             # tool catalog view derived from tools.Registry
  service.go             # gateway service: list tools, call tools, record invocations
  handler.go             # Gin routes for gateway APIs
  dto.go                 # HTTP request/response DTOs
  source.go              # invocation source constants

internal/agent/skills/
  model.go               # DiagnosisSkill domain model
  builtin.go             # built-in skill definitions
  service.go             # skill lookup, validation, enable/disable rules
  prompt.go              # skill prompt rendering helpers
  filter.go              # allowed tool filtering

internal/agent/events/
  event.go               # RuntimeEvent and event type constants
  sink.go                # EventSink interface and noop sink
  sse.go                 # SSE sink implementation, if kept outside http package

internal/mcp/
  server.go              # MCP server assembly
  tool_adapter.go        # Tool Gateway -> MCP Tool adapter
  resources.go           # MCP resources from tools/skills/evidence/trace
  prompts.go             # MCP prompts from Diagnosis Skills

cmd/mcp-server/
  main.go                # local stdio MCP server entrypoint
```

Modify these existing files as needed:

```text
cmd/server/main.go
internal/http/router.go
internal/http/agent_handler.go
internal/store/models.go
internal/store/db.go
internal/store/session_repo.go
internal/store/tool_call_repo.go
internal/agent/runtime/runtime.go
internal/agent/contextbuilder/builder.go
internal/agent/guard/scenario.go
docs/local-smoke.md
docs/design.md
docs/implementation-plan.md
```

Do not edit or depend on frontend code in this backend plan. The frontend AI should consume the API contracts documented below.

---

## 2. Phase 10: Tool Gateway API

### Goal

Expose the existing backend tools as a governed catalog and playground so the frontend can show tool capabilities, manually call read-only tools, and inspect invocation traces.

### Backend Behavior

Add API group:

```text
GET  /gateway/tools
GET  /gateway/tools/:name
POST /gateway/tools/:name/call
GET  /gateway/invocations
GET  /gateway/invocations/:id
```

`GET /gateway/tools` response shape:

```json
{
  "tools": [
    {
      "name": "get_video_detail",
      "display_name": "视频详情",
      "category": "video",
      "description": "Fetch one video from video-feed.",
      "read_only": true,
      "schema": {
        "type": "function",
        "function": {
          "name": "get_video_detail",
          "description": "Fetch one video from video-feed.",
          "parameters": {}
        }
      }
    }
  ]
}
```

`POST /gateway/tools/:name/call` request:

```json
{
  "arguments": {
    "video_id": 101
  },
  "source": "manual_console",
  "session_id": 1,
  "skill_id": "hot_rank_attribution"
}
```

`POST /gateway/tools/:name/call` response:

```json
{
  "invocation": {
    "id": 1,
    "source": "manual_console",
    "tool_name": "get_video_detail",
    "status": "success",
    "latency_ms": 12,
    "result_summary": "video 101: ...",
    "created_at": "2026-05-21T20:00:00+08:00"
  },
  "result": {
    "tool_name": "get_video_detail",
    "summary": "video 101: ...",
    "data": {}
  }
}
```

`GET /gateway/invocations` query params:

```text
source=manual_console|agent_runtime|mcp_client
tool_name=get_video_detail
session_id=1
skill_id=hot_rank_attribution
status=success|error|timeout
limit=50
```

### Store Changes

Add a new table rather than overloading `agent_tool_calls` immediately:

```go
type GatewayToolInvocation struct {
    ID            uint      `gorm:"primaryKey" json:"id"`
    Source        string    `gorm:"size:32;index;not null" json:"source"`
    SessionID     *uint     `gorm:"index" json:"session_id,omitempty"`
    MessageID     *uint     `gorm:"index" json:"message_id,omitempty"`
    SkillID       string    `gorm:"size:64;index" json:"skill_id,omitempty"`
    SkillVersion  string    `gorm:"size:32" json:"skill_version,omitempty"`
    ToolName      string    `gorm:"size:128;index;not null" json:"tool_name"`
    ArgumentsJSON string    `gorm:"type:text;not null" json:"arguments_json"`
    ResultJSON    string    `gorm:"type:text" json:"result_json,omitempty"`
    ResultSummary string    `gorm:"type:text" json:"result_summary,omitempty"`
    LatencyMS     int64     `gorm:"not null;default:0" json:"latency_ms"`
    Status        string    `gorm:"size:32;index;not null" json:"status"`
    ErrorMessage  string    `gorm:"type:text" json:"error_message,omitempty"`
    CreatedAt     time.Time `gorm:"index" json:"created_at"`
}
```

Source constants:

```go
const (
    InvocationSourceAgentRuntime   = "agent_runtime"
    InvocationSourceManualConsole  = "manual_console"
    InvocationSourceMCPClient      = "mcp_client"
)
```

### Files

- Create: `internal/gateway/catalog.go`
- Create: `internal/gateway/dto.go`
- Create: `internal/gateway/service.go`
- Create: `internal/gateway/handler.go`
- Create: `internal/gateway/source.go`
- Create: `internal/gateway/service_test.go`
- Create: `internal/gateway/handler_test.go`
- Create: `internal/store/gateway_invocation_repo.go`
- Create: `internal/store/gateway_invocation_repo_test.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/db.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/http/router.go`

### TDD Tasks

- [ ] Write failing store test: creating and listing gateway invocations by source/tool/session.
- [ ] Run: `go test ./internal/store -run GatewayInvocation -count=1`
- [ ] Implement `GatewayToolInvocation` model, AutoMigrate entry, and repository.
- [ ] Run: `go test ./internal/store -count=1`
- [ ] Write failing gateway service test: catalog returns stable sorted tools with categories.
- [ ] Implement catalog builder over `tools.Registry.Schemas()`.
- [ ] Write failing gateway service test: manual tool call records success invocation.
- [ ] Implement `Service.CallTool`.
- [ ] Write failing handler tests for list tools, call tool, and list invocations.
- [ ] Register routes through `WithGatewayHandler`.
- [ ] Run: `go test ./internal/gateway ./internal/http ./internal/store -count=1`
- [ ] Run full verification: `go test ./... && go vet ./... && git diff --check`
- [ ] Commit: `feat: add tool gateway api`

### Acceptance Criteria

- Frontend can list tools without calling LLM.
- Frontend can manually call a tool and see structured result plus summary.
- Every manual tool call is persisted with `source=manual_console`.
- Existing Agent Chat APIs still pass tests.

---

## 3. Phase 11: Diagnosis Skills Domain

### Goal

Introduce configurable diagnosis methods so operators can choose how the Agent should analyze a business question.

### Skill Model

Do not implement arbitrary executable Skills in v1. A Skill is structured metadata plus prompt/report/evidence rules:

```go
type DiagnosisSkill struct {
    ID                   string   `json:"id"`
    Name                 string   `json:"name"`
    Description          string   `json:"description"`
    Version              string   `json:"version"`
    Status               string   `json:"status"` // enabled|disabled
    Scenario             string   `json:"scenario"`
    AllowedTools         []string `json:"allowed_tools"`
    RequiredEvidence     []string `json:"required_evidence"`
    PromptTemplate       string   `json:"prompt_template"`
    OutputSections       []string `json:"output_sections"`
    RiskNotes            []string `json:"risk_notes,omitempty"`
}
```

Built-in Skills:

```text
hot_rank_attribution
comment_risk_analysis
author_support_evaluation
tag_trend_analysis
content_review_summary
```

Example built-in skill:

```yaml
id: comment_risk_analysis
name: 评论风险分析
description: 识别评论区是否存在敏感词、重复内容、负面反馈和异常互动。
version: 1.0.0
status: enabled
scenario: comment_risk_analysis
allowed_tools:
  - get_video_detail
  - analyze_video_comment_risk
required_evidence:
  - get_video_detail
  - analyze_video_comment_risk
output_sections:
  - 结论
  - 命中规则
  - 代表证据
  - 运营建议
```

### API Contract

```text
GET  /skills
GET  /skills/:id
POST /skills
PUT  /skills/:id
POST /skills/:id/enable
POST /skills/:id/disable
```

For v1, custom skill persistence can be SQLite-backed. If time is tight, implement built-in read-only Skills first, then add create/update in a separate commit.

### Store Changes

Add:

```go
type DiagnosisSkillRecord struct {
    ID                   string    `gorm:"primaryKey;size:64" json:"id"`
    Name                 string    `gorm:"size:128;not null" json:"name"`
    Description          string    `gorm:"type:text" json:"description"`
    Version              string    `gorm:"size:32;not null" json:"version"`
    Status               string    `gorm:"size:32;index;not null" json:"status"`
    Scenario             string    `gorm:"size:64;index" json:"scenario"`
    AllowedToolsJSON     string    `gorm:"type:text;not null" json:"allowed_tools_json"`
    RequiredEvidenceJSON string    `gorm:"type:text;not null" json:"required_evidence_json"`
    PromptTemplate       string    `gorm:"type:text;not null" json:"prompt_template"`
    OutputSectionsJSON   string    `gorm:"type:text;not null" json:"output_sections_json"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`
}
```

### Files

- Create: `internal/agent/skills/model.go`
- Create: `internal/agent/skills/builtin.go`
- Create: `internal/agent/skills/service.go`
- Create: `internal/agent/skills/prompt.go`
- Create: `internal/agent/skills/filter.go`
- Create: `internal/agent/skills/service_test.go`
- Create: `internal/store/skill_repo.go`
- Create: `internal/store/skill_repo_test.go`
- Create: `internal/http/skill_handler.go`
- Create: `internal/http/skill_handler_test.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/db.go`
- Modify: `cmd/server/main.go`
- Modify: `internal/http/router.go`

### TDD Tasks

- [ ] Write failing test: built-in skill list contains the five expected Skills.
- [ ] Implement built-in Skill definitions.
- [ ] Write failing test: Skill validation rejects unknown tools, empty evidence, disabled skill for runtime use.
- [ ] Implement `skills.Service`.
- [ ] Write failing store test for custom Skill CRUD.
- [ ] Implement `SkillRepository`.
- [ ] Write failing HTTP tests for `GET /skills` and `GET /skills/:id`.
- [ ] Implement `SkillHandler`.
- [ ] Add create/update/enable/disable only after read endpoints are green.
- [ ] Run: `go test ./internal/agent/skills ./internal/store ./internal/http -count=1`
- [ ] Run full verification.
- [ ] Commit: `feat: add diagnosis skills`

### Acceptance Criteria

- Backend exposes built-in Skills through API.
- Disabled Skills cannot be selected for runtime use.
- Skill definitions are validated against actual registered tools.
- No arbitrary script execution exists.

---

## 4. Phase 12: Skill-Driven Runtime

### Goal

Make Skills affect the Agent's actual runtime behavior: tool visibility, evidence requirements, prompt style, report structure, and trace metadata.

### Runtime Request Changes

Add `SkillID` and optional explicit `RequiredEvidence` override:

```go
type RunRequest struct {
    SessionID        uint
    UserMessage      string
    SkillID          string
    RequiredEvidence []string
}
```

Session should persist:

```go
SkillID      string `gorm:"size:64;index" json:"skill_id,omitempty"`
SkillVersion string `gorm:"size:32" json:"skill_version,omitempty"`
```

When a session has `skill_id`, runtime should:

- Load the Skill.
- Fail if Skill is disabled or missing.
- Expose only `AllowedTools` schemas to the LLM.
- Use `RequiredEvidence` from Skill unless request explicitly overrides it.
- Add a Skill prompt block to context.
- Record `skill_id` and `skill_version` in tool invocations.

### Tool Filtering

Add method to registry:

```go
func (r *Registry) SchemasFor(names []string) ([]ToolSchema, error)
```

The order should be stable and should reject unknown tool names during Skill validation.

### Context Builder Change

Add optional build request fields:

```go
type BuildRequest struct {
    SessionID        uint
    RequiredEvidence []string
    SkillPrompt      string
}
```

System prompt should include:

```text
Active diagnosis skill: 评论风险分析
Skill instructions:
...
Required output sections:
- 结论
- 命中规则
- 代表证据
- 运营建议
```

### Evidence Guard Change

Keep `guard.DetectScenario` for fallback, but Skill should be primary:

```text
if session.skill_id != "":
  requiredEvidence = skill.RequiredEvidence
else:
  requiredEvidence = guard.RequiredTools(guard.DetectScenario(userMessage))
```

### Files

- Modify: `internal/store/models.go`
- Modify: `internal/store/session_repo.go`
- Modify: `internal/http/agent_handler.go`
- Modify: `internal/agent/tools/registry.go`
- Modify: `internal/agent/runtime/runtime.go`
- Modify: `internal/agent/contextbuilder/builder.go`
- Modify: `internal/agent/guard/scenario.go` only for fallback compatibility.
- Modify tests in corresponding packages.

### TDD Tasks

- [ ] Write failing session repo test: create session with `skill_id` and `skill_version`.
- [ ] Implement session model/repo changes.
- [ ] Write failing registry test: `SchemasFor` returns only requested tools and rejects unknown names.
- [ ] Implement `SchemasFor`.
- [ ] Write failing runtime test: selected Skill limits schemas sent to fake LLM.
- [ ] Write failing runtime test: Skill required evidence drives guard retry.
- [ ] Write failing runtime test: disabled Skill returns clear error.
- [ ] Implement runtime Skill loading and filtering.
- [ ] Write failing contextbuilder test: Skill prompt appears in system context.
- [ ] Implement contextbuilder Skill prompt injection.
- [ ] Run: `go test ./internal/agent/runtime ./internal/agent/contextbuilder ./internal/agent/tools ./internal/store ./internal/http -count=1`
- [ ] Run full verification.
- [ ] Commit: `feat: enforce skill-driven runtime`

### Acceptance Criteria

- Starting a session with `skill_id=comment_risk_analysis` makes the LLM see only comment-risk related tools.
- Evidence Guard uses the Skill evidence list, not hard-coded scenario rules.
- Existing sessions without `skill_id` still work through scenario fallback.
- Frontend can display the Skill used by each session.

---

## 5. Phase 13: Source-Aware Unified Trace

### Goal

Unify Agent runtime tool traces and Tool Gateway invocations enough for frontend inspection and future metrics.

### Decision

Do not delete `agent_tool_calls` yet. Existing runtime tests and context builder depend on it. Instead:

- Keep writing `agent_tool_calls` for Agent evidence/context.
- Also write `gateway_tool_invocations` from Agent runtime with `source=agent_runtime`.
- Link records by `session_id`, `message_id`, `tool_name`, and timestamps.

Later, after tests are stable, consider making `gateway_tool_invocations` the canonical table and generating context from it.

### Runtime Change

Add optional `InvocationRecorder` dependency:

```go
type InvocationRecorder interface {
    Record(ctx context.Context, input gateway.RecordInvocationInput) error
}
```

`executeToolCall` should write both:

- `agent_tool_calls` for existing context/evidence behavior.
- `gateway_tool_invocations` for unified frontend trace.

### API Additions

Enhance:

```text
GET /gateway/invocations?source=agent_runtime&session_id=1
```

### TDD Tasks

- [ ] Write failing runtime test: runtime tool call writes gateway invocation with `source=agent_runtime`.
- [ ] Implement recorder interface and runtime dependency.
- [ ] Write failing handler test: list invocations by session/source.
- [ ] Run targeted and full verification.
- [ ] Commit: `feat: record source-aware tool invocations`

### Acceptance Criteria

- Frontend can show manual, MCP, and Agent tool calls in one trace list.
- Context Builder behavior remains unchanged.
- No duplicate final assistant messages or changed Agent behavior.

---

## 6. Phase 14: SSE Agent Events

### Goal

Expose Agent execution progress to the frontend as event stream. This gives a much better product demo than a blocking JSON response.

### API Contract

Keep existing blocking endpoint:

```text
POST /agent/sessions/:id/messages
```

Add:

```text
POST /agent/sessions/:id/messages/stream
Accept: text/event-stream
Content-Type: application/json
```

Request body matches existing message endpoint:

```json
{
  "content": "请分析 video_id=101 的评论风险",
  "required_evidence": [],
  "skill_id": "comment_risk_analysis"
}
```

SSE events:

```text
agent_start
skill_loaded
llm_round_start
tool_call
tool_result
guard_retry
final_answer
error
```

Example:

```text
event: tool_call
data: {"tool_name":"analyze_video_comment_risk","arguments":{"video_id":101,"limit":50}}

event: tool_result
data: {"tool_name":"analyze_video_comment_risk","status":"success","summary":"low comment risk for video 101...","latency_ms":4}

event: final_answer
data: {"content":"...","round_count":3,"tool_call_count":2}
```

### Runtime Change

Add:

```go
type RuntimeEvent struct {
    Type          string         `json:"type"`
    SessionID     uint           `json:"session_id"`
    SkillID       string         `json:"skill_id,omitempty"`
    ToolName      string         `json:"tool_name,omitempty"`
    Arguments     map[string]any `json:"arguments,omitempty"`
    Summary       string         `json:"summary,omitempty"`
    Status        string         `json:"status,omitempty"`
    Error         string         `json:"error,omitempty"`
    FinalAnswer   string         `json:"final_answer,omitempty"`
    RoundCount    int            `json:"round_count,omitempty"`
    ToolCallCount int            `json:"tool_call_count,omitempty"`
}

type EventSink interface {
    Emit(ctx context.Context, event RuntimeEvent) error
}
```

Blocking endpoint uses a noop sink. Streaming endpoint uses an SSE sink.

### Files

- Create: `internal/agent/events/event.go`
- Create: `internal/agent/events/sink.go`
- Create: `internal/http/sse.go`
- Modify: `internal/agent/runtime/runtime.go`
- Modify: `internal/http/agent_handler.go`
- Add tests.

### TDD Tasks

- [ ] Write failing runtime test: emits `agent_start`, `tool_call`, `tool_result`, `final_answer`.
- [ ] Implement event sink support.
- [ ] Write failing HTTP test: stream endpoint emits `text/event-stream`.
- [ ] Implement Gin SSE streaming.
- [ ] Verify client disconnect cancels runtime context.
- [ ] Run targeted and full verification.
- [ ] Commit: `feat: stream agent runtime events`

### Acceptance Criteria

- Frontend can display execution timeline while Agent is running.
- Existing blocking endpoint still works.
- Tool errors stream `error` event and return cleanly.

---

## 7. Phase 15: MCP Adapter

### Goal

Expose VideoOps capabilities to external AI clients through MCP without turning the project into a generic gateway platform.

### Scope

MCP Tools:

- Expose only read-only Tool Gateway tools.
- Use existing `tools.Tool` implementations.
- Record `source=mcp_client` gateway invocation.

MCP Prompts:

- Expose Diagnosis Skills as prompt templates.
- Prompt content should include skill instructions, required evidence, and output sections.

MCP Resources:

```text
videoops://tools
videoops://skills
videoops://evidence-rules
videoops://sessions/{id}/trace
```

Transport:

- Start with local stdio server in `cmd/mcp-server`.
- Do not implement generic HTTP+SSE MCP gateway unless a later requirement needs remote clients.

### Files

- Create: `cmd/mcp-server/main.go`
- Create: `internal/mcp/server.go`
- Create: `internal/mcp/tool_adapter.go`
- Create: `internal/mcp/resources.go`
- Create: `internal/mcp/prompts.go`
- Create tests for adapters.
- Modify: `go.mod` if using official Go MCP SDK.

### Implementation Notes

Prefer official MCP Go SDK if stable and easy to wire. If SDK integration is heavy, implement a minimal adapter plan first and keep actual MCP execution for a later branch. Do not block Tool Gateway or Skills on MCP SDK churn.

MCP tool output should include both text and structured data when supported:

```json
{
  "content": [
    {
      "type": "text",
      "text": "low comment risk for video 101 with 0 findings across 1 comments"
    }
  ],
  "structuredContent": {
    "tool_name": "analyze_video_comment_risk",
    "summary": "...",
    "data": {}
  }
}
```

### TDD Tasks

- [ ] Write adapter test: internal `ToolSchema` maps to MCP tool schema.
- [ ] Write adapter test: MCP tool call executes through Tool Gateway service.
- [ ] Write adapter test: MCP call records `source=mcp_client`.
- [ ] Write resources test: tools/skills/evidence resources return stable JSON.
- [ ] Write prompts test: built-in Skills appear as MCP prompts.
- [ ] Implement `cmd/mcp-server`.
- [ ] Add `docs/mcp-smoke.md` with local startup instructions.
- [ ] Run targeted and full verification.
- [ ] Commit: `feat: expose videoops capabilities through mcp`

### Acceptance Criteria

- An MCP client can list VideoOps tools.
- An MCP client can call a read-only VideoOps tool.
- MCP calls appear in gateway invocation trace.
- Skills are visible as MCP prompts.
- Trace/evidence metadata is visible as MCP resources.

---

## 8. Phase 16: Backend Console Contracts

### Goal

Give the frontend implementer stable API documents.

### Docs To Create

```text
docs/api-console-contract.md
docs/gateway-contract.md
docs/skills-contract.md
docs/sse-events.md
docs/mcp-smoke.md
```

Each doc should include:

- Endpoint path and method.
- Request JSON.
- Response JSON.
- Error format.
- Authentication assumptions.
- Local smoke command.

### TDD Tasks

- [ ] Update docs after each backend phase.
- [ ] Add curl/PowerShell smoke examples.
- [ ] Run `git diff --check`.
- [ ] Commit: `docs: add console backend contracts`

---

## 9. Phase 17: Evaluation Metrics

### Goal

Measure whether Skills + Evidence Guard improve evidence completeness and reduce unsupported final answers.

### Metrics

Add derived metrics from existing traces:

```text
tool_call_success_rate
tool_call_error_count
unauthorized_tool_call_count
guard_retry_count
evidence_complete_final_answer_count
evidence_incomplete_final_answer_rejected_count
average_tool_latency_ms
average_round_count
average_tool_call_count
skill_success_count
skill_failure_count
```

### Suggested Eval Modes

```text
baseline: Function Calling + Tool Registry
skill_guard: Diagnosis Skills + Evidence Guard
```

Do not publish metrics in resume until this phase produces local measured output.

### API

```text
GET /eval/summary
GET /eval/skills/:id/summary
POST /eval/runs
GET /eval/runs/:id
```

### Files

- Create: `internal/eval/model.go`
- Create: `internal/eval/service.go`
- Create: `internal/eval/handler.go`
- Create tests.
- Create: `docs/evaluation-plan.md`

### TDD Tasks

- [ ] Write service tests over seeded invocation/session records.
- [ ] Implement metric aggregation.
- [ ] Add HTTP summary endpoints.
- [ ] Add documented manual evaluation prompts.
- [ ] Run full verification.
- [ ] Commit: `feat: add agent evaluation metrics`

---

## 10. Recommended Commit Boundaries

Use these commit titles unless the implementation splits further:

```text
feat: add tool gateway api
feat: add diagnosis skills
feat: enforce skill-driven runtime
feat: record source-aware tool invocations
feat: stream agent runtime events
feat: expose videoops capabilities through mcp
docs: add console backend contracts
feat: add agent evaluation metrics
```

Each commit must pass:

```powershell
go test ./...
go vet ./...
git diff --check
git status --ignored -sb
```

If frontend files are still untracked, do not stage them unless the user explicitly asks.

---

## 11. Local Smoke Path

Assume existing local setup:

```text
video-feed: http://127.0.0.1:8080
video-ops-agent: http://127.0.0.1:8090
local LLM: http://127.0.0.1:8317/v1
model: gpt-5.4-mini
```

Use ignored `configs/config.yaml` for secrets. Do not commit API keys.

After Phase 10:

```powershell
Invoke-RestMethod http://127.0.0.1:8090/gateway/tools | ConvertTo-Json -Depth 8

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/gateway/tools/analyze_video_comment_risk/call" `
  -ContentType "application/json" `
  -Body (@{
    source = "manual_console"
    arguments = @{
      video_id = 101
      limit = 50
    }
  } | ConvertTo-Json -Depth 8) | ConvertTo-Json -Depth 10
```

After Phase 12:

```powershell
$session = Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions" `
  -ContentType "application/json" `
  -Body (@{
    user_id = "local-smoke"
    title = "Skill smoke"
    scenario = "comment_risk_analysis"
    skill_id = "comment_risk_analysis"
  } | ConvertTo-Json)

$sid = $session.session.id

Invoke-RestMethod `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages" `
  -ContentType "application/json" `
  -Body (@{
    content = "请分析 video_id=101 的评论风险，并给出运营建议。"
  } | ConvertTo-Json) `
  -TimeoutSec 180
```

After Phase 14:

```powershell
# Use Invoke-WebRequest for raw SSE inspection.
Invoke-WebRequest `
  -Method Post `
  -Uri "http://127.0.0.1:8090/agent/sessions/$sid/messages/stream" `
  -ContentType "application/json" `
  -Headers @{ Accept = "text/event-stream" } `
  -Body (@{
    content = "请分析 video_id=101 的评论风险，并展示证据链。"
    skill_id = "comment_risk_analysis"
  } | ConvertTo-Json)
```

---

## 12. Risks And Decisions

### Risk: Scope Creep Into Generic Gateway

Keep v1 tools domain-specific. Do not add arbitrary HTTP registration or protocol mapping. The Java `ai-mcp-gateway` is a reference for concepts, not a dependency or target architecture.

### Risk: Skill Prompt Injection

Skills are stored prompts and rules. They must not execute scripts or read files. Validate custom Skill content and keep built-ins as the default path.

### Risk: Duplicate Trace Tables

Phase 13 intentionally duplicates writes to preserve existing runtime behavior. After the UI and metrics are stable, a later refactor can choose one canonical trace table.

### Risk: MCP SDK Churn

Keep MCP as Phase 15. If the Go SDK creates friction, finish Tool Gateway and Skills first. MCP should adapt stable internal APIs, not drive the domain design.

### Risk: SSE Cancellation And Partial Writes

SSE endpoint must cancel runtime on client disconnect and emit clear `error` events. Blocking endpoint must stay available for tests and simple clients.

---

## 13. Next Session Startup Prompt

Use this prompt in the implementation session:

```text
你现在在 D:\dev\my_proj\go\video-ops-agent。

目标：按 docs/backend-evolution-plan.md 执行后端演进计划。先从 Phase 10 Tool Gateway API 开始，不要跳到 Skills/SSE/MCP。每个 phase 必须 TDD：先写失败测试，再实现，再跑验证。每个 phase 或小闭环都要单独 commit，不要堆一个大 commit。

重要边界：
- 只做后端，不改前端，除非我明确要求。
- 不要把 D:\dev\learn_proj\java\xfg\ai-mcp-gateway 作为组件接进来；只借鉴它的网关/工具/协议/鉴权建模思路。
- 不做通用 MCP 网关，不做任意 HTTP 协议映射。
- 不做写操作工具，不做可执行脚本 Skills。
- 不要提交 configs/config.yaml、data/、docs/study-notes/、web/node_modules/、web/dist/ 等本地产物。
- 当前本地 LLM 在 http://127.0.0.1:8317/v1，模型 gpt-5.4-mini；API key 只放环境变量，不要写入代码或文档。

执行要求：
1. 先读 docs/backend-evolution-plan.md。
2. 再检查 git status -sb 和现有后端结构。
3. 执行 Phase 10：Tool Gateway API。
4. 每个测试用 go test 定向验证，最后跑 go test ./...、go vet ./...、git diff --check。
5. commit 标题优先使用文档里的推荐标题，例如 feat: add tool gateway api。
6. 完成 Phase 10 后停下来汇报，不要自动继续 Phase 11，除非我要求继续。
```

---

## 14. Definition Of Done For This Roadmap

The roadmap is complete only when:

- Tool Gateway can list and manually call tools.
- Diagnosis Skills can define tool whitelist, evidence requirements, and report structure.
- Runtime enforces Skill-driven tool visibility and evidence checks.
- Frontend can consume documented APIs for tools, skills, trace, and SSE.
- MCP clients can list and call read-only VideoOps tools through the adapter.
- Metrics are produced from local measured runs, not invented.
