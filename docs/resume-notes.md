# VideoOps Agent Resume Notes

## Recommended Project Title

`VideoOps Agent：短视频内容运营诊断 Agent`

## Relationship With video-feed

This project should be described as a second project that builds on the first one.

```text
video-feed:
短视频 Feed 流平台，负责账号、视频、Feed、热榜、评论、关注、通知、缓存、MQ 和压测。

video-ops-agent:
基于 video-feed 平台数据的内容运营诊断 Agent，负责工具调用、证据追踪和运营报告生成。
```

## Resume-Safe Summary

> 基于 Go 构建短视频内容运营诊断 Agent，将热榜、评论、作者画像和标签 Feed 等平台能力封装为 Function Calling 工具，并实现 Agent Runtime、Tool Registry、Context Builder、Session Memory、Trace Recorder 和 Evidence Guard，支持热度归因、评论风险分析、作者画像和标签趋势报告。

## Possible Bullet Points

- 设计并实现工具调用型 Agent Runtime，支持 LLM 基于 Function Calling 动态调用 `get_video_detail`、`get_video_comments`、`get_hot_videos`、`get_author_profile` 等平台工具，完成短视频运营分析任务。
- 将 `video-feed` 的热榜、评论、作者、标签等内部能力封装为结构化只读 tools，统一处理参数校验、超时控制、错误返回和结果摘要，降低 LLM 直接拼接接口的不可控性。
- 实现会话级 Context Builder 与 Session Memory，对多轮对话和工具调用结果进行摘要、裁剪和证据组织，控制上下文长度并提升多轮分析稳定性。
- 实现 Agent Trace 记录机制，持久化每次工具调用的参数、结果摘要、耗时、状态和错误信息，使最终运营报告能够追溯到真实平台数据。
- 设计 Evidence Guard 证据完整性校验，对热榜归因、评论风险、作者画像和标签趋势等场景约束必要工具调用，防止模型缺少证据时直接生成空泛结论。
- 基于评论规则检测与 LLM 总结生成评论风险报告，输出风险等级、代表评论、风险类型和运营处理建议。

## Strong Interview Angles

### 1. Why not only call an LLM API?

Because the LLM only decides tool-call intent. The backend must own:

- tool registry
- argument validation
- permission boundary
- tool execution
- timeout control
- session memory
- context building and truncation
- trace persistence
- evidence completeness guard

This is the engineering value of the project.

### 2. Why Function Calling?

Natural-language operation questions often require multiple platform data sources. Function Calling lets the model choose tools based on the question and previous tool results instead of forcing every scenario into a fixed API.

### 3. Why Evidence Guard?

LLMs can answer too early. For example, hot-rank attribution is weak if the model only calls `get_video_detail` and never reads comments or hot-rank context. Evidence Guard ensures the final report has minimum required evidence.

### 4. Why Context Builder and Session Memory?

Function Calling creates multi-step conversations. If every raw tool result is sent back to the LLM, the context becomes noisy and can exceed the model window. Session Memory stores messages and tool-call history, while Context Builder selects recent messages, evidence requirements, and compact tool summaries for the next LLM call.

The MVP only implements session-level memory. It does not implement cross-session long-term memory or vector memory.

### 5. Why not Multi-Agent in MVP?

The first version focuses on making a single Agent Runtime reliable and measurable. Multi-Agent can be added later as a Supervisor workflow with evidence, risk, and report roles, but calling several functions "agents" without independent coordination would be weak in an interview.

### 6. Why independent service?

It makes the resume project boundary clearer:

- `video-feed` is the platform backend.
- `video-ops-agent` is the AI Agent application built on top of platform APIs.

It also avoids coupling Agent logic directly into the platform's core business service.

## Claims To Avoid Before Implementation

Do not claim:

- production-grade moderation accuracy
- autonomous multi-agent system
- long-term memory or vector memory
- personalized recommendation algorithm
- RAG or vector retrieval
- ASR/video transcript support
- latency or accuracy improvement numbers

Use measured local evidence before writing numbers.

## Metrics To Collect Later

After implementation, collect local measured data:

- tool-call success rate
- invalid tool-call count
- evidence completeness rate
- guard intervention count
- average prompt context size
- tool result truncation count
- average Agent response latency
- final reports with evidence ratio

Useful comparison:

```text
Function Calling only
Function Calling + Evidence Guard
```

Only use results from actual local runs.

## One-Minute Interview Pitch

这个项目是我在短视频 Feed 平台之上做的内容运营诊断 Agent。平台侧已经有视频、评论、热榜、作者和标签数据，Agent 服务不直接访问数据库，而是把这些平台能力封装成 Function Calling 工具。LLM 负责判断要调用哪些工具，Go 后端负责工具注册、参数校验、执行、上下文组装、超时控制和 Trace 落库。为了避免模型没有证据就生成结论，我还做了 Evidence Guard，比如热榜归因必须至少查询视频详情、热榜上下文和评论数据，最终报告可以追溯到每次工具调用结果。
