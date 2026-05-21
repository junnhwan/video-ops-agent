# VideoOps Agent Design

## 1. Project Name

English name: `VideoOps Agent`

Chinese name: `短视频内容运营诊断 Agent`

Repository directory: `D:\dev\my_proj\go\video-ops-agent`

One-line positioning:

> 面向短视频平台内容运营场景，基于 Feed、热榜、评论、作者画像和标签数据，构建一个支持 Function Calling、上下文管理、工具调用追踪和证据校验的 AI Agent，帮助运营人员完成热度归因、评论风险识别、作者画像和标签趋势分析。

This project is designed as a second resume project paired with `video-feed`.

- `video-feed`: short-video platform backend, providing account, video, Feed, hot rank, comments, social data, cache, MQ, benchmark tooling.
- `video-ops-agent`: operations diagnosis Agent, calling `video-feed` APIs as tools and producing evidence-grounded reports.

## 2. Target User

Primary user: platform content operator.

The operator needs to answer questions such as:

- Why did this video enter the hot rank?
- Is this comment section risky?
- Is this author worth promoting?
- What type of content is becoming popular under a tag?
- Is the platform data enough to support an operation decision?

Secondary user: creator.

Creator-facing analysis can be added later, but it is not the first MVP target.

Non-target for the first version: normal viewer.

Viewer-facing features often become recommendation, semantic search, or video Q&A. The current `video-feed` project does not yet have watch history modeling, embeddings, vector retrieval, or ranking models, so viewer-facing AI recommendation is not the right first scope.

## 3. Core Problem

The project should not be framed as "calling an LLM API". The engineering problem is:

- Platform data is scattered across videos, comments, hot rank, tags, authors, and social modules.
- Operators need an explanation, not just raw metrics.
- LLM answers are not trustworthy if the model can answer without evidence.
- Function Calling needs a backend runtime to register tools, execute tools, validate arguments, control loops, and record trace.
- Multi-turn Agent work needs context management so message history and tool results do not overflow or pollute the LLM input.

Therefore, the backend value is the Agent runtime:

```text
LLM decides tool_call intent.
Go backend validates, executes, records, and guards the tool calls.
```

## 4. MVP Scenarios

### 4.1 Hot Rank Attribution

Example prompt:

> 分析一下视频 123 为什么上热榜，有没有运营风险。

Required evidence:

- video detail
- current hot videos / hot context
- comments
- author profile

Expected report:

- conclusion
- key evidence
- possible growth reason
- comment risk
- operation suggestion

### 4.2 Comment Risk Analysis

Example prompt:

> 分析视频 123 的评论区有没有争议、攻击或刷屏风险。

Required evidence:

- video detail
- comments
- rule-based risk scan
- optional LLM risk summary

Expected report:

- risk level
- risk categories
- representative comments
- suggested action

### 4.3 Author Profile Diagnosis

Example prompt:

> 作者 8 最近表现怎么样，值得扶持吗？

Required evidence:

- author profile
- author videos
- interaction data from videos

Expected report:

- content direction
- interaction performance
- strengths and weaknesses
- promotion suggestion

### 4.4 Tag Trend Report

Example prompt:

> #Go 后端 这个标签最近内容表现怎么样？

Required evidence:

- tag videos
- representative video details
- comments from selected videos

Expected report:

- trending content types
- representative videos
- audience concerns
- operation suggestion

## 5. Architecture

```text
User
  -> Agent Chat API
  -> Agent Runtime
  -> Context Builder builds compact LLM context
  -> LLM Client with tool schemas
  -> LLM returns tool_call
  -> Tool Registry finds tool
  -> Tool Executor validates and runs tool
  -> VideoFeedClient calls video-feed API
  -> Tool result is summarized, recorded, and returned to LLM
  -> Loop continues until final answer
  -> Evidence Guard checks evidence completeness
  -> Final report is saved with trace
```

## 6. Backend Modules

Recommended Go package layout:

```text
cmd/server
internal/config
internal/http
internal/agent/runtime
internal/agent/llm
internal/agent/contextbuilder
internal/agent/tools
internal/agent/guard
internal/agent/trace
internal/platform/videofeed
internal/store
```

Responsibilities:

- `cmd/server`: service entrypoint.
- `internal/config`: config loading for HTTP server, DB, LLM, and `video-feed` base URL.
- `internal/http`: Gin router and handlers.
- `internal/agent/runtime`: Agent loop, max rounds, timeout control, final answer flow.
- `internal/agent/llm`: OpenAI-compatible Chat Completions client and Function Calling protocol.
- `internal/agent/contextbuilder`: session context, recent message window, tool-result summaries, and context size control.
- `internal/agent/tools`: tool definitions, schemas, and executors.
- `internal/agent/guard`: argument validation, evidence completeness, permission boundary.
- `internal/agent/trace`: session, message, and tool-call recording.
- `internal/platform/videofeed`: typed HTTP client for `video-feed`.
- `internal/store`: database models and repositories.

## 7. Tool List

First version should use read-only tools.

```text
get_video_detail(video_id)
get_hot_videos(limit)
get_video_comments(video_id, limit)
get_author_profile(author_id)
list_author_videos(author_id, limit)
list_tag_videos(tag_name, limit)
analyze_video_comment_risk(video_id, limit)
analyze_comment_risk(video_id, comments)
```

`analyze_video_comment_risk` is the preferred MVP tool for comment-risk scenarios because the backend fetches comments internally before running deterministic rules. `analyze_comment_risk` remains a lower-level scanner for already available comment arrays.

Possible second-stage tools:

```text
compare_video_metrics(video_id, peer_video_ids)
generate_operation_report(evidence)
read_benchmark_summary(result_path)
```

Write tools such as publishing comments, sending messages, following users, or changing recommendations are excluded from MVP. If added later, they must require explicit confirmation.

## 8. Context Builder and Session Memory

The MVP should include context management, but only at session level.

Context Builder decides what is sent to the LLM for the current Agent step:

```text
- system prompt
- current user message
- recent session messages
- active scenario and evidence requirements
- previous tool-call summaries
- compact tool results needed for reasoning
```

Session Memory means persisted conversation and tool history inside one Agent session:

```text
agent_messages
agent_tool_calls
```

It is not long-term user memory. The MVP does not remember cross-session user preferences.

Context Builder responsibilities:

- Keep the last `N` user/assistant messages, for example `6`.
- Include tool results as summaries rather than full raw payloads when possible.
- Truncate comment lists and large JSON payloads before sending to the LLM.
- Preserve evidence needed by Evidence Guard.
- Avoid sending API keys, internal stack traces, or unnecessary raw database-like payloads to the LLM.
- Provide deterministic tests for message selection and truncation.

Suggested context budget policy for MVP:

```text
max_recent_messages: 6
max_tool_result_chars: 4000
max_comments_for_llm: 50
max_comment_chars_each: 300
```

## 9. Function Calling Runtime

The first real Agent loop:

```text
1. Save user message.
2. Build compact LLM context from session messages and tool-call summaries.
3. Send context and tool schemas to LLM.
4. If LLM returns tool_call:
   - validate tool name
   - validate arguments
   - execute tool with timeout
   - save tool call trace
   - summarize tool result
   - append tool result summary to context
   - continue loop
5. If LLM returns final answer:
   - run Evidence Guard
   - if evidence is missing, ask LLM to call missing tools
   - otherwise save final answer
6. Return final answer and trace summary.
```

Runtime limits:

- Max tool rounds: `6`
- Single tool timeout: `2s`
- Total request timeout: `30s`
- Comment fetch limit: `50`
- Tool result must be summarized before being stored as `result_summary`
- Full result can be stored as JSON only when size is controlled

## 10. Evidence Guard

Evidence Guard is not a replacement for LLM reasoning. It prevents unsupported final answers.

Example rules:

```text
hot_rank_analysis requires:
- get_video_detail
- get_hot_videos
- get_video_comments

comment_risk_analysis requires:
- get_video_detail
- analyze_video_comment_risk

author_profile_analysis requires:
- get_author_profile
- list_author_videos

tag_trend_analysis requires:
- list_tag_videos
```

If the model tries to produce a final answer before required evidence exists, the runtime appends a system instruction:

```text
Evidence is incomplete. Call the missing tools before producing the final report.
```

## 11. Data Model

Minimum tables:

```text
agent_sessions
- id
- user_id
- title
- scenario
- status
- context_policy_json
- created_at
- updated_at

agent_messages
- id
- session_id
- role
- content
- content_summary
- created_at

agent_tool_calls
- id
- session_id
- message_id
- tool_name
- arguments_json
- result_json
- result_summary
- latency_ms
- status
- error_message
- created_at
```

Optional table:

```text
agent_reports
- id
- session_id
- report_type
- target_type
- target_id
- content
- evidence_json
- created_at
```

MVP can store reports as assistant messages first. `agent_reports` can be added after report templates become stable.

## 12. API Design

MVP JSON APIs:

```text
GET /health

POST /agent/sessions
GET /agent/sessions
GET /agent/sessions/:id

POST /agent/sessions/:id/messages
GET /agent/sessions/:id/tool-calls
```

Second-stage SSE:

```text
GET /agent/sessions/:id/stream
```

SSE event types:

```text
agent_start
tool_call
tool_result
guard_retry
final_answer
error
```

## 13. Tech Decisions

Default decisions:

- Backend: Go + Gin.
- Persistence: GORM with SQLite for first local MVP; MySQL can be added later if needed.
- LLM: OpenAI-compatible Chat Completions API.
- Frontend: not required in first backend MVP.
- Integration: call `video-feed` through HTTP APIs, not direct DB sharing.
- Config: no API keys committed. Use environment variables or local config ignored by git.
- Memory: session memory only for MVP; no cross-session long-term memory.
- Multi-agent: not included in MVP. The single Agent Runtime should be made reliable first.

## 14. Non-Goals

First version does not include:

- ASR or video transcript extraction
- RAG or vector database
- viewer recommendation
- autonomous write operations
- multi-agent collaboration
- long-term memory or operator preference memory
- production moderation enforcement
- generated metrics without local measurement

## 15. Future Extensions

Potential second-stage capabilities:

- SSE Agent Console.
- Long-term operator preference memory.
- Evaluation dashboard.
- Supervisor-style multi-agent workflow with separate evidence, risk, and report roles.

These should be added only after the single Agent Runtime, Context Builder, Evidence Guard, and Trace Recorder are working and measurable.

## 16. Resume-Safe Framing

Safe project framing:

> 基于 Go 构建短视频内容运营诊断 Agent，将热榜、评论、作者画像和标签 Feed 等平台能力封装为 Function Calling 工具，并实现 Agent Runtime、Tool Registry、Context Builder、Session Memory、Trace Recorder 和 Evidence Guard，支持热度归因、评论风险分析、作者画像和标签趋势报告。

Claims to avoid until implemented and measured:

- AI recommendation system
- vector search / RAG
- autonomous multi-agent platform
- production-grade content moderation
- specific latency or accuracy improvements without local experiment evidence
