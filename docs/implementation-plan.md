# VideoOps Agent Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development if subagents are available, otherwise use superpowers:executing-plans. Steps use checkbox syntax for tracking.

**Goal:** Build a standalone Go Agent service that uses LLM Function Calling to call `video-feed` platform tools and generate evidence-grounded operation reports.

**Architecture:** The service runs independently from `video-feed`. It exposes Agent Chat APIs, calls `video-feed` through a typed HTTP client, executes registered tools, builds compact LLM context from session memory, records trace in a local database, and uses Evidence Guard to prevent unsupported final answers.

**Tech Stack:** Go, Gin, GORM, SQLite first, OpenAI-compatible Chat Completions, `net/http` test servers, PowerShell-friendly commands.

---

## Phase 0: Project Bootstrap

**Files:**

- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `internal/config/config.go`
- Create: `internal/http/router.go`
- Create: `configs/config.example.yaml`
- Create: `.gitignore`

Steps:

- [ ] Initialize Go module, recommended module name: `video-ops-agent`.
- [ ] Add Gin, GORM, SQLite driver, YAML config dependency.
- [ ] Implement `/health`.
- [ ] Add config fields:
  - server address
  - database DSN
  - LLM base URL
  - LLM model
  - LLM API key environment variable name
  - `video-feed` base URL
- [ ] Add `.gitignore` for local config, DB files, logs, and build artifacts.
- [ ] Verify:

```powershell
go test ./...
go run ./cmd/server
Invoke-RestMethod http://127.0.0.1:8090/health
```

Expected: health returns `{"status":"ok"}`.

## Phase 1: VideoFeedClient

**Files:**

- Create: `internal/platform/videofeed/client.go`
- Create: `internal/platform/videofeed/types.go`
- Create: `internal/platform/videofeed/client_test.go`

Steps:

- [ ] Define typed methods:
  - `GetVideoDetail(ctx, videoID)`
  - `GetHotVideos(ctx, limit)`
  - `GetVideoComments(ctx, videoID, limit)`
  - `GetAuthorProfile(ctx, authorID)`
  - `ListAuthorVideos(ctx, authorID, limit)`
  - `ListTagVideos(ctx, tagName, limit)`
- [ ] Use existing `video-feed` HTTP routes:
  - `/video/getDetail`
  - `/feed/listByPopularity`
  - `/comment/listAll`
  - `/account/getProfile`
  - `/video/listByAuthorID`
  - `/feed/listByTag`
- [ ] Write tests with `httptest.Server`.
- [ ] Test non-200 response and malformed JSON.
- [ ] Verify:

```powershell
go test ./internal/platform/videofeed -v
```

Expected: client tests pass without starting real `video-feed`.

## Phase 2: Tool Registry and Executor

**Files:**

- Create: `internal/agent/tools/tool.go`
- Create: `internal/agent/tools/registry.go`
- Create: `internal/agent/tools/executor.go`
- Create: `internal/agent/tools/platform_tools.go`
- Create: `internal/agent/tools/comment_risk.go`
- Create: `internal/agent/tools/*_test.go`

Steps:

- [ ] Define `Tool`, `ToolSchema`, `ToolResult`.
- [ ] Implement registry lookup by tool name.
- [ ] Implement argument decoding and validation.
- [ ] Implement read-only platform tools.
- [ ] Implement `analyze_comment_risk` with deterministic rules first:
  - repeated content
  - sensitive words from a local list
  - excessive mentions
  - negative keywords
- [ ] Implement `analyze_video_comment_risk(video_id, limit)` as the preferred comment-risk tool so the backend fetches comments internally instead of requiring the LLM to repackage a comments array.
- [ ] Add per-tool timeout support.
- [ ] Verify:

```powershell
go test ./internal/agent/tools -v
```

Expected: all tools can be executed through the registry with fake platform client dependencies.

## Phase 3: Trace Persistence

**Files:**

- Create: `internal/store/db.go`
- Create: `internal/store/models.go`
- Create: `internal/store/session_repo.go`
- Create: `internal/store/message_repo.go`
- Create: `internal/store/tool_call_repo.go`
- Create: `internal/store/*_test.go`

Steps:

- [ ] Add GORM models:
  - `AgentSession`
  - `AgentMessage`
  - `AgentToolCall`
- [ ] AutoMigrate on server startup.
- [ ] Implement repositories for session, message, and tool call records.
- [ ] Use SQLite in tests.
- [ ] Verify:

```powershell
go test ./internal/store -v
```

Expected: repositories can create and read sessions, messages, and tool calls.

## Phase 4: LLM Client

**Files:**

- Create: `internal/agent/llm/client.go`
- Create: `internal/agent/llm/types.go`
- Create: `internal/agent/llm/client_test.go`

Steps:

- [ ] Implement OpenAI-compatible `/chat/completions` request.
- [ ] Support tool schemas in requests.
- [ ] Parse assistant final answer.
- [ ] Parse tool calls.
- [ ] Do not log API keys.
- [ ] Use fake HTTP server tests.
- [ ] Verify:

```powershell
go test ./internal/agent/llm -v
```

Expected: final-answer and tool-call responses are parsed correctly.

## Phase 5: Context Builder and Session Memory

**Files:**

- Create: `internal/agent/contextbuilder/context.go`
- Create: `internal/agent/contextbuilder/builder.go`
- Create: `internal/agent/contextbuilder/summarizer.go`
- Create: `internal/agent/contextbuilder/builder_test.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/message_repo.go`
- Modify: `internal/store/tool_call_repo.go`

Steps:

- [ ] Add optional `content_summary` to messages.
- [ ] Add optional `context_policy_json` to sessions.
- [ ] Define `ContextPolicy`:
  - recent message limit
  - max tool result characters
  - max comments for LLM
  - max characters per comment
- [ ] Implement context selection:
  - system prompt
  - active scenario
  - latest user message
  - recent session messages
  - previous tool-call summaries
  - required evidence list
- [ ] Implement deterministic truncation for large tool results.
- [ ] Implement comment compaction so long comments do not dominate the prompt.
- [ ] Do not include API keys, internal stack traces, or raw oversized payloads in LLM context.
- [ ] Verify:

```powershell
go test ./internal/agent/contextbuilder ./internal/store -v
```

Expected: builder tests prove recent-message selection, tool-result truncation, and comment compaction.

## Phase 6: Agent Runtime Loop

**Files:**

- Create: `internal/agent/runtime/runtime.go`
- Create: `internal/agent/runtime/prompt.go`
- Create: `internal/agent/runtime/runtime_test.go`
- Use: `internal/agent/contextbuilder`

Steps:

- [ ] Accept user message and session ID.
- [ ] Save user message.
- [ ] Build compact context from session memory and tool-call summaries.
- [ ] Send compact context and tool schemas to LLM.
- [ ] Execute returned tool calls.
- [ ] Save each tool call trace:
  - tool name
  - arguments
  - result summary
  - latency
  - status
  - error
- [ ] Summarize tool result and append it to the next context.
- [ ] Stop on final answer or max rounds.
- [ ] Save assistant answer.
- [ ] Verify with fake LLM:

```powershell
go test ./internal/agent/runtime -v
```

Expected: runtime executes tool calls and produces final answer with persisted trace.

## Phase 7: Evidence Guard

**Files:**

- Create: `internal/agent/guard/scenario.go`
- Create: `internal/agent/guard/evidence.go`
- Create: `internal/agent/guard/evidence_test.go`
- Modify: `internal/agent/runtime/runtime.go`

Steps:

- [ ] Detect scenario from user question:
  - hot rank analysis
  - comment risk analysis
  - author profile analysis
  - tag trend analysis
  - general
- [ ] Define required tool sets per scenario.
- [ ] Before accepting final answer, check whether required tools were called.
- [ ] If evidence is missing, append instruction asking LLM to call missing tools.
- [ ] Prevent infinite retries with max guard retry count.
- [ ] Verify:

```powershell
go test ./internal/agent/guard ./internal/agent/runtime -v
```

Expected: a fake LLM that tries to answer too early is forced to call missing tools.

## Phase 8: Agent HTTP APIs

**Files:**

- Create: `internal/http/agent_handler.go`
- Create: `internal/http/agent_handler_test.go`
- Modify: `internal/http/router.go`
- Modify: `cmd/server/main.go`

Endpoints:

```text
POST /agent/sessions
GET /agent/sessions
GET /agent/sessions/:id
POST /agent/sessions/:id/messages
GET /agent/sessions/:id/tool-calls
```

Steps:

- [ ] Add session creation endpoint.
- [ ] Add session detail endpoint with messages.
- [ ] Add message endpoint that runs Agent Runtime.
- [ ] Add tool-call trace endpoint.
- [ ] Verify:

```powershell
go test ./internal/http -v
go test ./...
```

Expected: HTTP tests pass.

## Phase 9: Local Integration With video-feed

Prerequisite:

- `video-feed` API service is running.
- The database has seeded or manually created videos, comments, authors, and tags.

Steps:

- [ ] Start `video-feed`.
- [ ] Start `video-ops-agent`.
- [ ] Create Agent session.
- [ ] Ask a hot rank question.
- [ ] Ask a comment risk question.
- [ ] Confirm tool calls are persisted.
- [ ] Confirm final answer references tool evidence.

Manual smoke commands should be written after actual API request bodies are implemented.

## Phase 10: SSE and Simple Console

This phase is optional for first backend MVP.

**Backend:**

- Add `GET /agent/sessions/:id/stream`.
- Emit:
  - `agent_start`
  - `tool_call`
  - `tool_result`
  - `guard_retry`
  - `final_answer`
  - `error`

**Frontend:**

- Keep it compact:
  - left: sessions
  - center: chat
  - right: tool trace and report

## Phase 11: Evaluation and Resume Evidence

Metrics must come from local runs.

Recommended evaluation cases:

- 10 hot-rank prompts
- 10 comment-risk prompts
- 10 author-profile prompts
- 10 tag-trend prompts

Record:

- tool-call success rate
- invalid tool-call count
- average latency
- evidence completeness rate
- guard intervention count
- average prompt context size
- tool result truncation count
- final answer with evidence ratio

Compare:

```text
Function Calling only
Function Calling + Evidence Guard
```

Do not publish unmeasured metrics.

## Suggested Commit Boundaries

- `chore: bootstrap video ops agent service`
- `feat: add video-feed platform client`
- `feat: add agent tool registry`
- `feat: persist agent trace`
- `feat: add llm function calling client`
- `feat: add context builder and session memory`
- `feat: implement agent runtime loop`
- `feat: add evidence guard`
- `feat: expose agent chat api`
- `feat: add agent stream events`
- `docs: add evaluation report`
