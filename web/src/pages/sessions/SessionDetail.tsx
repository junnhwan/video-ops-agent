import { useState, useEffect, useRef, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  ArrowLeft,
  Send,
  Wrench,
  Bot,
  Loader2,
  Zap,
  AlertTriangle,
  FileText,
  Activity,
  Sparkles,
  ChevronDown,
  ChevronRight,
  CheckCircle2,
  Clock,
  XCircle,
  Copy,
  Check,
  Download,
  Link2,
  ShieldCheck,
  Terminal,
  CornerDownLeft,
  X,
} from "lucide-react";
import type { AgentSession, AgentMessage, AgentToolCall, SSEEvent, Skill } from "../../types";
import { sessionApi, skillApi } from "../../lib/api";
import { createSSEConnection } from "../../lib/sse-client";
import { cn, formatDate, formatDuration, safeJSONParse } from "../../lib/utils";

const toolStepLabels: Record<string, string> = {
  get_video_detail: "视频详情",
  get_hot_videos: "热榜数据",
  get_video_comments: "评论采集",
  get_author_profile: "作者信息",
  list_author_videos: "作者作品",
  list_tag_videos: "标签视频",
  analyze_video_comment_risk: "风险分析",
  analyze_comment_risk: "风险分析",
};

const scenarioLabels: Record<string, string> = {
  hot_rank_attribution: "热榜归因",
  comment_risk_analysis: "评论风险",
  author_support_evaluation: "作者评估",
  tag_trend_analysis: "标签趋势",
  content_review_summary: "内容复盘",
};

/* ========== Diagnosis Report Card (old style) ========== */

function DiagnosisReport({
  message,
  toolCalls,
}: {
  message: AgentMessage;
  toolCalls: AgentToolCall[];
}) {
  const [evidenceOpen, setEvidenceOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(message.content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {}
  }, [message.content]);

  const handleExport = useCallback(() => {
    const lines = [
      `# ${message.content_summary || "诊断报告"}`,
      "",
      message.content,
    ];
    if (toolCalls.length > 0) {
      lines.push("", "---", `## 证据链 (${toolCalls.length} 次工具调用)`, "");
      toolCalls.forEach((tc, i) => {
        lines.push(
          `### ${i + 1}. ${toolStepLabels[tc.tool_name] || tc.tool_name}`
        );
        if (tc.result_summary) lines.push("", tc.result_summary);
        lines.push(
          "",
          `> 耗时: ${formatDuration(tc.latency_ms)} | 状态: ${tc.status}`,
          ""
        );
      });
    }
    const blob = new Blob([lines.join("\n")], {
      type: "text/markdown;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `诊断报告.md`;
    a.click();
    URL.revokeObjectURL(url);
  }, [message, toolCalls]);

  return (
    <div className="report-card relative bg-[var(--color-surface-raised)] border border-[var(--color-border-subtle)] rounded-xl overflow-hidden">
      {/* Left accent bar */}
      <div className="absolute left-0 top-0 bottom-0 w-[3px] bg-gradient-to-b from-[var(--color-cyan)] via-[var(--color-accent)] to-[var(--color-cyan)] opacity-50 rounded-l-xl" />

      {/* Report header */}
      <div className="flex items-center gap-2 pl-5 pr-4 py-2.5 border-b border-[var(--color-border-subtle)] bg-[var(--color-surface-overlay)]">
        <FileText size={13} className="text-[var(--color-cyan)]" />
        <span className="text-xs font-medium text-[var(--color-text-primary)]">
          诊断报告
        </span>
        {message.content_summary && (
          <span className="text-[11px] text-[var(--color-text-tertiary)] bg-[var(--color-surface)] px-1.5 py-0.5 rounded ml-1">
            {message.content_summary}
          </span>
        )}
        <div className="flex-1" />
        <button
          onClick={handleCopy}
          className="btn-press p-1 rounded hover:bg-[var(--color-surface)] text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] transition-colors"
          title="复制"
        >
          {copied ? (
            <Check size={12} className="text-[var(--color-emerald)]" />
          ) : (
            <Copy size={12} />
          )}
        </button>
        <button
          onClick={handleExport}
          className="btn-press p-1 rounded hover:bg-[var(--color-surface)] text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] transition-colors"
          title="导出 Markdown"
        >
          <Download size={12} />
        </button>
        {toolCalls.length > 0 && (
          <div className="flex items-center gap-1 text-[11px] text-[var(--color-emerald)] font-medium ml-2">
            <ShieldCheck size={11} />
            <span>{toolCalls.length} 项证据</span>
          </div>
        )}
      </div>

      {/* Report body */}
      <div className="pl-5 pr-4 py-3">
        <div className="flex items-start gap-2">
          <CheckCircle2
            size={14}
            className="text-[var(--color-accent)] mt-0.5 shrink-0"
          />
          <div className="report-prose flex-1">
            <Markdown remarkPlugins={[remarkGfm]}>
              {message.content}
            </Markdown>
          </div>
        </div>
      </div>

      {/* Evidence chain (collapsible) */}
      {toolCalls.length > 0 && (
        <div>
          <button
            onClick={() => setEvidenceOpen(!evidenceOpen)}
            className="w-full flex items-center gap-2 pl-5 pr-4 py-2 text-xs text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-overlay)] transition-colors"
          >
            <Link2 size={12} />
            <span className="font-medium">证据链</span>
            <span className="text-[var(--color-text-muted)]">
              ({toolCalls.length} 次工具调用)
            </span>
            <div className="flex-1" />
            {evidenceOpen ? (
              <ChevronDown size={12} />
            ) : (
              <ChevronRight size={12} />
            )}
          </button>
          <div className={cn("collapse-wrapper", evidenceOpen && "open")}>
            <div className="collapse-inner">
              <div className="pl-5 pr-4 pb-3 space-y-1.5">
                {toolCalls.map((tc) => (
                  <div
                    key={tc.id}
                    className="flex items-start gap-2 text-xs py-1.5 px-2.5 rounded-md bg-[var(--color-surface-overlay)]"
                  >
                    <div
                      className={cn(
                        "w-4 h-4 rounded-full flex items-center justify-center shrink-0 mt-0.5",
                        tc.status === "success"
                          ? "bg-[var(--color-emerald-soft)]"
                          : tc.status === "error"
                            ? "bg-[var(--color-rose-soft)]"
                            : "bg-[var(--color-amber-soft)]"
                      )}
                    >
                      {tc.status === "success" ? (
                        <CheckCircle2
                          size={9}
                          className="text-[var(--color-emerald)]"
                        />
                      ) : tc.status === "error" ? (
                        <AlertTriangle
                          size={9}
                          className="text-[var(--color-rose)]"
                        />
                      ) : (
                        <Clock
                          size={9}
                          className="text-[var(--color-amber)]"
                        />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <span className="font-mono font-medium text-[var(--color-text-primary)]">
                        {toolStepLabels[tc.tool_name] || tc.tool_name}
                      </span>
                      {tc.result_summary && (
                        <p className="text-[var(--color-text-tertiary)] mt-0.5 line-clamp-2 leading-relaxed">
                          {tc.result_summary}
                        </p>
                      )}
                    </div>
                    <span className="text-[11px] text-[var(--color-text-muted)] font-mono tabular-nums shrink-0">
                      {formatDuration(tc.latency_ms)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

/* ========== Sending Indicator ========== */

function SendingIndicator() {
  return (
    <div className="flex gap-3 items-start animate-slide-up">
      <div className="w-8 h-8 rounded-lg bg-[var(--color-cyan-soft)] border border-[var(--color-cyan)] border-opacity-15 flex items-center justify-center shrink-0">
        <Bot size={14} className="text-[var(--color-cyan)]" />
      </div>
      <div className="bg-[var(--color-surface-raised)] border border-[var(--color-border-subtle)] rounded-xl px-4 py-4 shadow-sm">
        <div className="space-y-2.5 w-64">
          <div className="flex items-center gap-2 mb-3">
            <Loader2
              size={13}
              className="text-[var(--color-accent)] animate-spin"
            />
            <span className="text-xs text-[var(--color-text-secondary)]">
              Agent 正在调用工具分析数据...
            </span>
          </div>
          <div className="h-2.5 skeleton w-full" />
          <div className="h-2.5 skeleton w-4/5" />
          <div className="h-2.5 skeleton w-3/5" />
        </div>
      </div>
    </div>
  );
}

/* ========== Empty State ========== */

function EmptyMessagesPlaceholder() {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center animate-fade-in">
      <div className="w-14 h-14 rounded-xl bg-[var(--color-accent-soft)] border border-[var(--color-accent-border)] flex items-center justify-center mb-5 shadow-sm">
        <Sparkles size={24} className="text-[var(--color-accent)]" />
      </div>
      <h3 className="text-base font-semibold text-[var(--color-text-primary)] mb-1.5">
        开始诊断分析
      </h3>
      <p className="text-sm text-[var(--color-text-tertiary)] max-w-xs leading-relaxed mb-6">
        输入运营问题，Agent 将调用平台工具采集数据，
        <br />
        并生成基于证据的诊断报告。
      </p>
      <div className="flex items-center gap-6 text-[11px] text-[var(--color-text-tertiary)]">
        <div className="flex items-center gap-1.5">
          <Zap size={12} className="text-[var(--color-amber)]" />
          <span>工具调用</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Activity size={12} className="text-[var(--color-cyan)]" />
          <span>证据追踪</span>
        </div>
        <div className="flex items-center gap-1.5">
          <FileText size={12} className="text-[var(--color-accent)]" />
          <span>诊断报告</span>
        </div>
      </div>
    </div>
  );
}

/* ========== Tool Trace Panel (old style) ========== */

function ToolTracePanel({
  events,
  toolCalls,
  onClose,
}: {
  events: SSEEvent[];
  toolCalls: AgentToolCall[];
  onClose: () => void;
}) {
  const [expandedId, setExpandedId] = useState<number | null>(null);

  return (
    <aside className="w-[360px] flex flex-col bg-[var(--color-surface-raised)] border-l border-[var(--color-border-subtle)] h-full">
      {/* Header */}
      <div className="h-12 shrink-0 px-4 flex items-center justify-between border-b border-[var(--color-border-subtle)]">
        <div className="flex items-center gap-2">
          <Zap size={14} className="text-[var(--color-cyan)]" />
          <span className="text-xs font-semibold text-[var(--color-text-primary)]">
            工具调用追踪
          </span>
          <span className="text-[11px] text-[var(--color-text-tertiary)] bg-[var(--color-surface-overlay)] px-1.5 py-0.5 rounded">
            {toolCalls.length} 次调用
          </span>
        </div>
        <button
          onClick={onClose}
          className="btn-press p-1 rounded hover:bg-[var(--color-surface-overlay)] text-[var(--color-text-tertiary)] transition-colors"
        >
          <X size={14} />
        </button>
      </div>

      {/* SSE Pipeline progress */}
      {events.length > 0 && (
        <div className="px-4 py-3 border-b border-[var(--color-border-subtle)] bg-[var(--color-surface-overlay)]">
          <div className="text-[11px] font-medium text-[var(--color-text-tertiary)] uppercase tracking-wider mb-2">
            实时事件
          </div>
          <div className="flex flex-wrap items-center gap-1.5">
            {events.map((evt, i) => {
              const color =
                evt.type === "error"
                  ? "bg-[var(--color-rose)]"
                  : evt.type === "tool_call"
                    ? "bg-[var(--color-amber)]"
                    : evt.type === "final_answer"
                      ? "bg-[var(--color-emerald)]"
                      : "bg-[var(--color-accent)]";
              const label =
                evt.type === "agent_start"
                  ? "启动"
                  : evt.type === "skill_loaded"
                    ? "技能"
                    : evt.type === "llm_round_start"
                      ? `R${evt.round_count || i}`
                      : evt.type === "tool_call"
                        ? evt.tool_name?.split("_").pop() || "调用"
                        : evt.type === "tool_result"
                          ? "结果"
                          : evt.type === "guard_retry"
                            ? "重试"
                            : evt.type === "final_answer"
                              ? "完成"
                              : evt.type === "error"
                                ? "错误"
                                : evt.type;
              return (
                <span
                  key={i}
                  className={cn(
                    "text-[10px] font-medium text-white px-1.5 py-0.5 rounded-full",
                    color
                  )}
                >
                  {label}
                </span>
              );
            })}
          </div>
        </div>
      )}

      {/* Tool call list */}
      <div className="flex-1 overflow-y-auto p-3">
        {toolCalls.length === 0 && events.length === 0 ? (
          <div className="flex flex-col items-center py-10 text-center animate-fade-in">
            <Terminal
              size={20}
              className="text-[var(--color-text-muted)] mb-3"
            />
            <p className="text-xs text-[var(--color-text-tertiary)]">
              暂无工具调用
            </p>
            <p className="text-[11px] text-[var(--color-text-muted)] mt-1 max-w-[200px]">
              Agent 分析时调用的工具将在这里以时间线形式展示
            </p>
          </div>
        ) : (
          <div className="relative">
            <div className="absolute left-[15px] top-3 bottom-3 w-px bg-gradient-to-b from-[var(--color-border-default)] via-[var(--color-border-subtle)] to-transparent" />
            <div className="space-y-1.5">
              {toolCalls.map((tc, i) => (
                <ToolCallNode
                  key={tc.id}
                  toolCall={tc}
                  index={i + 1}
                  isExpanded={expandedId === tc.id}
                  onToggle={() =>
                    setExpandedId(expandedId === tc.id ? null : tc.id)
                  }
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}

function ToolCallNode({
  toolCall,
  index,
  isExpanded,
  onToggle,
}: {
  toolCall: AgentToolCall;
  index: number;
  isExpanded: boolean;
  onToggle: () => void;
}) {
  const args = safeJSONParse<Record<string, unknown>>(
    toolCall.arguments_json,
    {}
  );
  const result = safeJSONParse<unknown>(toolCall.result_json, null);

  const dotColor =
    toolCall.status === "success"
      ? "bg-[var(--color-emerald)]"
      : toolCall.status === "error"
        ? "bg-[var(--color-rose)]"
        : "bg-[var(--color-amber)]";

  const statusIcon =
    toolCall.status === "success" ? (
      <CheckCircle2 size={10} className="text-[var(--color-emerald)]" />
    ) : toolCall.status === "error" ? (
      <XCircle size={10} className="text-[var(--color-rose)]" />
    ) : (
      <Clock size={10} className="text-[var(--color-amber)]" />
    );

  const friendlyName =
    toolStepLabels[toolCall.tool_name] || toolCall.tool_name;

  return (
    <div
      className="relative pl-8 animate-fade-in"
      style={{
        animationDelay: `${(index - 1) * 60}ms`,
        animationFillMode: "both",
      }}
    >
      <div className="absolute left-[11px] top-3.5 z-10">
        <div
          className={cn(
            "w-[9px] h-[9px] rounded-full border-[2.5px] border-[var(--color-surface-raised)] transition-colors duration-300",
            dotColor
          )}
        />
      </div>

      <button
        onClick={onToggle}
        className={cn(
          "btn-press w-full text-left rounded-lg transition-all duration-200 mb-0.5",
          isExpanded
            ? "bg-[var(--color-surface-overlay)] border border-[var(--color-border-default)] shadow-sm"
            : "hover:bg-[var(--color-surface-overlay)] border border-transparent"
        )}
      >
        <div className="px-3 py-2.5 flex items-center gap-2">
          {isExpanded ? (
            <ChevronDown
              size={12}
              className="text-[var(--color-text-tertiary)]"
            />
          ) : (
            <ChevronRight
              size={12}
              className="text-[var(--color-text-muted)]"
            />
          )}
          <span className="text-[11px] font-mono text-[var(--color-text-tertiary)] bg-[var(--color-surface)] px-1 rounded shrink-0">
            #{index}
          </span>
          <span className="text-xs font-medium text-[var(--color-text-primary)] truncate">
            {friendlyName}
          </span>
          <span className="text-[11px] font-mono text-[var(--color-text-muted)] truncate hidden sm:inline">
            {toolCall.tool_name}
          </span>
          <div className="flex-1" />
          {statusIcon}
          <span className="text-[11px] text-[var(--color-text-tertiary)] font-mono tabular-nums">
            {formatDuration(toolCall.latency_ms)}
          </span>
        </div>
      </button>

      <div className={cn("collapse-wrapper", isExpanded && "open")}>
        <div className="collapse-inner">
          <div className="ml-2 border-l-2 border-[var(--color-accent)] border-opacity-20 pl-3 py-1 space-y-2.5 mb-1.5">
            <div>
              <div className="flex items-center gap-1 mb-1">
                <Wrench size={10} className="text-[var(--color-text-tertiary)]" />
                <span className="text-[11px] font-medium text-[var(--color-text-tertiary)]">
                  参数
                </span>
              </div>
              <pre className="text-[11px] text-[var(--color-text-secondary)] bg-[var(--color-surface)] border border-[var(--color-border-subtle)] rounded-md p-2.5 overflow-x-auto">
                {JSON.stringify(args, null, 2)}
              </pre>
            </div>

            {toolCall.result_summary && (
              <div>
                <div className="flex items-center gap-1 mb-1">
                  <Activity size={10} className="text-[var(--color-cyan)]" />
                  <span className="text-[11px] font-medium text-[var(--color-text-tertiary)]">
                    结果摘要
                  </span>
                </div>
                <p className="text-xs text-[var(--color-text-secondary)] bg-[var(--color-cyan-soft)] border border-[var(--color-cyan)] border-opacity-10 rounded-md p-2.5 leading-relaxed">
                  {toolCall.result_summary}
                </p>
              </div>
            )}

            {result != null && (
              <div>
                <div className="flex items-center gap-1 mb-1">
                  <Terminal
                    size={10}
                    className="text-[var(--color-text-tertiary)]"
                  />
                  <span className="text-[11px] font-medium text-[var(--color-text-tertiary)]">
                    原始结果
                  </span>
                </div>
                <pre className="text-[11px] text-[var(--color-text-tertiary)] bg-[var(--color-surface)] border border-[var(--color-border-subtle)] rounded-md p-2.5 overflow-x-auto max-h-36">
                  {JSON.stringify(result, null, 2)}
                </pre>
              </div>
            )}

            {toolCall.error_message && (
              <div>
                <div className="flex items-center gap-1 mb-1">
                  <AlertTriangle size={10} className="text-[var(--color-rose)]" />
                  <span className="text-[11px] font-medium text-[var(--color-rose)]">
                    错误
                  </span>
                </div>
                <p className="text-xs text-[var(--color-rose)] bg-[var(--color-rose-soft)] border border-[var(--color-rose)] border-opacity-10 rounded-md p-2.5">
                  {toolCall.error_message}
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

/* ========== Main Session Detail ========== */

export function SessionDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const sessionId = Number(id);

  const [session, setSession] = useState<AgentSession | null>(null);
  const [messages, setMessages] = useState<AgentMessage[]>([]);
  const [toolCalls, setToolCalls] = useState<AgentToolCall[]>([]);
  const [events, setEvents] = useState<SSEEvent[]>([]);
  const [inputText, setInputText] = useState("");
  const [skillId, setSkillId] = useState("");
  const [skills, setSkills] = useState<Skill[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [traceOpen, setTraceOpen] = useState(true);

  const sseController = useRef<AbortController | null>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Auto-scroll
  useEffect(() => {
    const timer = setTimeout(() => {
      chatEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, 60);
    return () => clearTimeout(timer);
  }, [messages, isStreaming]);

  // Auto-focus after send
  useEffect(() => {
    if (!isStreaming && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [isStreaming]);

  // Load session
  const loadSession = useCallback(async () => {
    if (isNaN(sessionId)) {
      setError("无效的会话 ID");
      setLoading(false);
      return;
    }
    try {
      const [data, tcData] = await Promise.all([
        sessionApi.get(sessionId),
        sessionApi.listToolCalls(sessionId),
      ]);
      setSession(data.session);
      setMessages(data.messages ?? []);
      setToolCalls(tcData.tool_calls ?? []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "会话未找到");
    } finally {
      setLoading(false);
    }
  }, [sessionId]);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  useEffect(() => {
    skillApi
      .list()
      .then((data) => setSkills((data.skills ?? []).filter((s) => s.status === "enabled")))
      .catch(() => setSkills([]));
  }, []);

  useEffect(() => {
    return () => sseController.current?.abort();
  }, []);

  // Send via SSE
  const handleSend = () => {
    const content = inputText.trim();
    if (!content || isStreaming) return;

    setIsStreaming(true);
    setEvents([]);

    const optimisticMsg: AgentMessage = {
      id: Date.now(),
      session_id: sessionId,
      role: "user",
      content,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, optimisticMsg]);
    setInputText("");
    if (textareaRef.current) textareaRef.current.style.height = "auto";

    const body: { content: string; skill_id?: string } = { content };
    if (skillId.trim()) body.skill_id = skillId.trim();

    sseController.current = createSSEConnection(sessionId, body, {
      onEvent: (event) => {
        setEvents((prev) => [...prev, event]);

        if (event.type === "final_answer") {
          const assistantMsg: AgentMessage = {
            id: Date.now() + 1,
            session_id: sessionId,
            role: "assistant",
            content: event.summary ?? "",
            created_at: new Date().toISOString(),
          };
          setMessages((prev) => [...prev, assistantMsg]);
          setIsStreaming(false);
          // Reload tool calls after completion
          sessionApi.listToolCalls(sessionId).then((res) =>
            setToolCalls(res.tool_calls ?? [])
          );
        }

        if (event.type === "error") {
          setIsStreaming(false);
        }
      },
      onError: (err) => {
        setEvents((prev) => [
          ...prev,
          { type: "error", error: err.message },
        ]);
        setIsStreaming(false);
        const errorMsg: AgentMessage = {
          id: Date.now() + 2,
          session_id: sessionId,
          role: "assistant",
          content: `**错误：** ${err.message}`,
          created_at: new Date().toISOString(),
        };
        setMessages((prev) => [...prev, errorMsg]);
      },
      onClose: () => {
        setIsStreaming(false);
      },
    });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleTextareaChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInputText(e.target.value);
    e.target.style.height = "auto";
    e.target.style.height = Math.min(e.target.scrollHeight, 120) + "px";
  };

  const sessionToolCalls = toolCalls.filter(
    (tc) => tc.session_id === sessionId
  );

  // ---- Render ----

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-120px)]">
        <Loader2 className="w-6 h-6 animate-spin text-[var(--color-accent)]" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-[calc(100vh-120px)] text-center">
        <AlertTriangle className="w-10 h-10 text-[var(--color-rose)] mb-3" />
        <p className="text-sm font-medium text-[var(--color-text-primary)]">
          {error}
        </p>
        <button
          onClick={() => navigate("/sessions")}
          className="console-btn console-btn-secondary mt-4"
        >
          <ArrowLeft className="w-4 h-4" />
          返回会话列表
        </button>
      </div>
    );
  }

  return (
    <div className="-m-6">
      <div className="flex h-[calc(100vh-48px)]">
        {/* ---- Chat Area ---- */}
        <div className="flex-1 flex flex-col min-h-0">
          {/* Chat header */}
          <div className="header-bar shrink-0 px-6 py-2.5 bg-[var(--color-surface-raised)] border-b border-[var(--color-border-subtle)] flex items-center justify-between animate-fade-in">
            <div className="flex items-center gap-3">
              <button
                onClick={() => navigate("/sessions")}
                className="btn-press p-1 rounded hover:bg-[var(--color-surface-overlay)] text-[var(--color-text-tertiary)] transition-colors"
              >
                <ArrowLeft size={16} />
              </button>
              <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
                {session?.title || `会话 #${sessionId}`}
              </h3>
              {session?.scenario && (
                <span className="text-[11px] px-2 py-0.5 rounded-full bg-[var(--color-accent-soft)] text-[var(--color-accent)] border border-[var(--color-accent-border)] font-medium">
                  {scenarioLabels[session.scenario] || session.scenario}
                </span>
              )}
            </div>
            <div className="flex items-center gap-3">
              <span className="text-[11px] text-[var(--color-text-tertiary)]">
                {messages.filter((m) => m.role === "user").length} 次提问
              </span>
              <span className="text-[11px] text-[var(--color-text-tertiary)]">
                {sessionToolCalls.length} 次工具调用
              </span>
              <button
                onClick={() => setTraceOpen(!traceOpen)}
                className={cn(
                  "btn-press p-1.5 rounded transition-colors",
                  traceOpen
                    ? "bg-[var(--color-accent-soft)] text-[var(--color-accent)]"
                    : "text-[var(--color-text-tertiary)] hover:bg-[var(--color-surface-overlay)]"
                )}
                title="工具追踪面板"
              >
                <Zap size={14} />
              </button>
            </div>
          </div>

          {/* Messages */}
          <div className="flex-1 overflow-y-auto px-6 py-6 space-y-5 workspace-bg scroll-smooth">
            {messages.length === 0 ? (
              <EmptyMessagesPlaceholder />
            ) : (
              messages.map((msg, i) => {
                const delay = Math.min(i * 60, 300);
                if (msg.role === "user") {
                  return (
                    <div
                      key={msg.id}
                      className="flex justify-end animate-slide-up"
                      style={{
                        animationDelay: `${delay}ms`,
                        animationFillMode: "both",
                      }}
                    >
                      <div className="max-w-[70%] bg-[var(--color-accent)] text-white rounded-2xl rounded-br-md px-4 py-3 shadow-sm">
                        <p className="text-sm whitespace-pre-wrap leading-relaxed">
                          {msg.content}
                        </p>
                        <span className="text-[11px] text-white/50 mt-1.5 block">
                          {formatDate(msg.created_at)}
                        </span>
                      </div>
                    </div>
                  );
                }

                // Assistant → DiagnosisReport
                return (
                  <div
                    key={msg.id}
                    className="flex gap-3 items-start animate-slide-up"
                    style={{
                      animationDelay: `${delay}ms`,
                      animationFillMode: "both",
                    }}
                  >
                    <div className="w-8 h-8 rounded-lg bg-[var(--color-cyan-soft)] border border-[var(--color-cyan)] border-opacity-15 flex items-center justify-center shrink-0 mt-0.5">
                      <Bot size={14} className="text-[var(--color-cyan)]" />
                    </div>
                    <div className="flex-1 min-w-0 max-w-[85%]">
                      <DiagnosisReport
                        message={msg}
                        toolCalls={sessionToolCalls}
                      />
                      <span className="text-[11px] text-[var(--color-text-muted)] mt-1.5 ml-1 block">
                        {formatDate(msg.created_at)}
                      </span>
                    </div>
                  </div>
                );
              })
            )}
            {isStreaming && <SendingIndicator />}
            <div ref={chatEndRef} />
          </div>

          {/* Input dock */}
          <div className="input-dock shrink-0 px-6 py-3 bg-[var(--color-surface-raised)] border-t border-[var(--color-border-subtle)]">
            {skillId !== undefined && (
              <div className="mb-2">
                <select
                  value={skillId}
                  onChange={(e) => setSkillId(e.target.value)}
                  className="console-input text-xs py-1.5"
                  disabled={isStreaming}
                >
                  <option value="">不指定技能（通用模式）</option>
                  {skills.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name ? `${s.name} · ${s.id}` : s.id}
                    </option>
                  ))}
                </select>
              </div>
            )}
            <div className="max-w-3xl mx-auto relative group">
              <div className="absolute -inset-1 rounded-2xl bg-[var(--color-accent)] opacity-0 group-focus-within:opacity-5 transition-opacity duration-300 pointer-events-none" />
              <textarea
                ref={textareaRef}
                rows={1}
                value={inputText}
                onChange={handleTextareaChange}
                onKeyDown={handleKeyDown}
                placeholder="输入运营诊断问题，例如：分析视频 123 为什么上热榜..."
                className="relative w-full text-sm px-4 py-3 pr-24 rounded-xl bg-[var(--color-surface-overlay)] border border-[var(--color-border-default)] text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] focus:outline-none focus:border-[var(--color-accent)] focus:border-opacity-50 focus:ring-2 focus:ring-[var(--color-accent)] focus:ring-opacity-10 resize-none transition-all duration-200 max-h-32"
                disabled={isStreaming}
              />
              <div className="absolute right-2 bottom-2 flex items-center gap-2">
                <span className="text-[11px] text-[var(--color-text-muted)] flex items-center gap-0.5 opacity-0 group-focus-within:opacity-100 transition-opacity duration-200">
                  <CornerDownLeft size={10} /> 发送
                </span>
                <button
                  onClick={handleSend}
                  disabled={!inputText.trim() || isStreaming}
                  className={cn(
                    "btn-press px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 flex items-center gap-1",
                    inputText.trim() && !isStreaming
                      ? "bg-[var(--color-accent)] text-white hover:bg-[var(--color-accent-hover)] shadow-sm"
                      : "bg-[var(--color-border-subtle)] text-[var(--color-text-muted)] cursor-not-allowed"
                  )}
                >
                  {isStreaming ? (
                    <Loader2 size={13} className="animate-spin" />
                  ) : (
                    <Send size={13} />
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* ---- Tool Trace Panel (right) ---- */}
        {traceOpen && (
          <ToolTracePanel
            events={events}
            toolCalls={sessionToolCalls}
            onClose={() => setTraceOpen(false)}
          />
        )}
      </div>
    </div>
  );
}
