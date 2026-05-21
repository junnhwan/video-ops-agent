import type { SSEEvent, SSEEventType } from "../types";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api";
const USE_MOCK = import.meta.env.VITE_USE_MOCK === "true";

export interface SSECallbacks {
  onEvent: (event: SSEEvent) => void;
  onError: (error: Error) => void;
  onClose: () => void;
}

// ===== Mock SSE: simulate a full agent run =====
const mockToolSequences = [
  {
    tool: "get_video_detail",
    args: { video_id: 123 },
    summary: "视频播放量 128.5w，点赞 12.3w，分享 8.9k",
    result: { video_id: 123, title: "Go 微服务实战", play_count: 1285000, like_count: 123000 },
  },
  {
    tool: "get_video_comments",
    args: { video_id: 123, limit: 50, sort: "hot" },
    summary: "获取 50 条评论，正面 68%，中性 28%，负面 4%",
    result: { total: 4560, sampled: 50, sentiment: { positive: 0.68, neutral: 0.28, negative: 0.04 } },
  },
  {
    tool: "analyze_video_comment_risk",
    args: { video_id: 123, depth: "full" },
    summary: "风险等级: Attention，发现 3 条引战评论",
    result: { risk_level: "attention", flagged: 3, rules_triggered: ["confrontation_detection"] },
  },
];

const mockFinalAnswer = `## Mock 诊断结果

这是一个模拟的 Agent 流式回复，用于展示完整的交互流程。

### 核心发现

| 指标 | 数值 | 评估 |
|------|------|------|
| 播放量 | 128.5w | 正常 |
| 点赞率 | 9.6% | 优秀 |
| 评论情感 | 正面 68% | 健康 |

### 风险评估

> 风险等级：**Attention**（需持续关注）

- 评论区出现个别引战言论（约 3.2%），围绕技术选型产生对立
- 建议设置关键词预警，暂不干预

### 建议

1. 开启评论预警，设置关键词自动提醒
2. 准备舆情回应话术模板
3. 持续监控异常分享账号后续行为`;

function createMockSSE(
  callbacks: SSECallbacks,
  controller: AbortController
) {
  const aborted = () => controller.signal.aborted;

  const emit = (type: SSEEventType, data: Partial<SSEEvent> = {}) => {
    if (aborted()) return;
    callbacks.onEvent({ type, ...data } as SSEEvent);
  };

  const wait = (ms: number) =>
    new Promise<void>((resolve) => {
      if (aborted()) return resolve();
      const timer = setTimeout(resolve, ms);
      controller.signal.addEventListener("abort", () => {
        clearTimeout(timer);
        resolve();
      }, { once: true });
    });

  (async () => {
    try {
      await wait(200);
      emit("agent_start");
      await wait(300);
      emit("skill_loaded", { skill_id: "comment_risk_analysis" });
      await wait(400);
      emit("llm_round_start", { round_count: 1 });

      for (let i = 0; i < mockToolSequences.length; i++) {
        if (aborted()) break;
        const seq = mockToolSequences[i];
        await wait(600 + Math.random() * 400);
        emit("tool_call", {
          tool_name: seq.tool,
          arguments: seq.args,
        });
        await wait(400 + Math.random() * 300);
        emit("tool_result", {
          tool_name: seq.tool,
          summary: seq.summary,
          result: seq.result,
        });
      }

      if (aborted()) return;

      await wait(500);
      emit("llm_round_start", { round_count: 2 });
      await wait(800);

      if (aborted()) return;
      emit("final_answer", {
        summary: mockFinalAnswer,
        round_count: 2,
        tool_call_count: mockToolSequences.length,
      });

      callbacks.onClose();
    } catch (err) {
      if (!aborted()) {
        callbacks.onError(err instanceof Error ? err : new Error(String(err)));
      }
    }
  })();
}

// ===== Real SSE =====
function createRealSSE(
  sessionId: number,
  body: {
    content: string;
    skill_id?: string;
    required_evidence?: string[];
  },
  callbacks: SSECallbacks,
  controller: AbortController
) {
  (async () => {
    try {
      const response = await fetch(
        `${API_BASE}/agent/sessions/${sessionId}/messages/stream`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Accept: "text/event-stream",
          },
          body: JSON.stringify(body),
          signal: controller.signal,
        }
      );

      if (!response.ok) {
        const err = await response.json().catch(() => ({}));
        throw new Error(err.error || `HTTP ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error("No response body");

      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");
        buffer = lines.pop() || "";

        let currentEventType = "";
        let currentData = "";

        for (const line of lines) {
          if (line.startsWith("event: ")) {
            currentEventType = line.slice(7).trim();
          } else if (line.startsWith("data: ")) {
            currentData = line.slice(6);
          } else if (line === "" && currentEventType && currentData) {
            try {
              const event = JSON.parse(currentData) as SSEEvent;
              event.type = currentEventType as SSEEventType;
              callbacks.onEvent(event);
            } catch {
              // skip malformed
            }
            currentEventType = "";
            currentData = "";
          }
        }
      }

      callbacks.onClose();
    } catch (err) {
      if (controller.signal.aborted) {
        callbacks.onClose();
      } else {
        callbacks.onError(
          err instanceof Error ? err : new Error(String(err))
        );
      }
    }
  })();
}

export function createSSEConnection(
  sessionId: number,
  body: {
    content: string;
    skill_id?: string;
    required_evidence?: string[];
  },
  callbacks: SSECallbacks
): AbortController {
  const controller = new AbortController();

  if (USE_MOCK) {
    createMockSSE(callbacks, controller);
  } else {
    createRealSSE(sessionId, body, callbacks, controller);
  }

  return controller;
}
