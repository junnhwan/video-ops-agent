import { useState, useRef, useEffect, useCallback } from "react";
import { cn, formatDate, formatDuration } from "../lib/utils";
import { useSession } from "./SessionProvider";
import { useConnectionStatus } from "../hooks/useConnectionStatus";
import type { AgentMessage } from "../types";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  Send,
  Bot,
  Loader2,
  Sparkles,
  FileText,
  Zap,
  AlertTriangle,
  CornerDownLeft,
  ChevronDown,
  ChevronRight,
  CheckCircle2,
  Lightbulb,
  BarChart3,
  ShieldCheck,
  Link2,
  Clock,
  Copy,
  Check,
  Download,
} from "lucide-react";

const scenarioLabels: Record<string, string> = {
  hot_rank_analysis: "热榜归因",
  comment_risk_analysis: "评论风险",
  author_profile_analysis: "作者画像",
  tag_trend_analysis: "标签趋势",
};

export function ChatArea() {
  const { currentSession, messages, sending, sendMessage, toolCalls } = useSession();
  const connStatus = useConnectionStatus();
  const [input, setInput] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }, 60);
    return () => clearTimeout(timer);
  }, [messages, sending]);

  // Auto-focus input after sending completes
  useEffect(() => {
    if (!sending && inputRef.current) {
      inputRef.current.focus();
    }
  }, [sending]);

  async function handleSend() {
    if (!input.trim() || !currentSession || sending) return;
    const content = input.trim();
    setInput("");
    if (inputRef.current) inputRef.current.style.height = "auto";
    await sendMessage(currentSession.id, { content });
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  if (!currentSession) return null;

  const sessionToolCalls = toolCalls.filter(tc => tc.session_id === currentSession.id);

  return (
    <div className="flex-1 flex flex-col min-h-0">
      <div className="header-bar shrink-0 px-6 py-2.5 bg-surface-raised border-b border-border-subtle flex items-center justify-between animate-fade-in">
        <div className="flex items-center gap-3">
          <h3 className="text-sm font-semibold text-text-primary">{currentSession.title}</h3>
          {currentSession.scenario && (
            <span className="text-[11px] px-2 py-0.5 rounded-full bg-accent-soft text-accent border border-accent-border font-medium">
              {scenarioLabels[currentSession.scenario]}
            </span>
          )}
        </div>
        <div className="flex items-center gap-4 text-[11px] text-text-tertiary">
          <span>{messages.filter(m => m.role === "user").length} 次提问</span>
          <span>{sessionToolCalls.length} 次工具调用</span>
        </div>
      </div>

      {/* Offline banner */}
      {connStatus === "offline" && (
        <div className="shrink-0 px-6 py-2 bg-rose-soft border-b border-rose/15 flex items-center gap-2 animate-fade-in">
          <AlertTriangle size={13} className="text-rose shrink-0" />
          <span className="text-xs text-rose font-medium">后端连接已断开，部分功能可能不可用</span>
        </div>
      )}

      <div className="flex-1 overflow-y-auto px-6 py-6 space-y-5 workspace-bg scroll-smooth">
        {messages.length === 0 ? (
          <EmptyMessagesPlaceholder />
        ) : (
          messages.map((msg, i) => <MessageBlock key={msg.id} message={msg} index={i} toolCalls={sessionToolCalls} />)
        )}
        {sending && <SendingIndicator />}
        <div ref={bottomRef} />
      </div>

      <div className="input-dock shrink-0 px-6 py-3 bg-surface-raised border-t border-border-subtle">
        <div className="max-w-3xl mx-auto relative group">
          <div className="absolute -inset-1 rounded-2xl bg-accent/5 opacity-0 group-focus-within:opacity-100 transition-opacity duration-300 pointer-events-none" />
          <textarea
            ref={inputRef}
            rows={1}
            value={input}
            onChange={(e) => {
              setInput(e.target.value);
              e.target.style.height = "auto";
              e.target.style.height = Math.min(e.target.scrollHeight, 120) + "px";
            }}
            onKeyDown={handleKeyDown}
            placeholder="输入运营诊断问题，例如：分析视频 123 为什么上热榜..."
            className="relative w-full text-sm px-4 py-3 pr-24 rounded-xl bg-surface-overlay border border-border-default text-text-primary placeholder-text-muted focus:outline-none focus:border-accent/50 focus:ring-2 focus:ring-accent/10 resize-none transition-all duration-200 max-h-32"
            disabled={sending}
          />
          <div className="absolute right-2 bottom-2 flex items-center gap-2">
            <span className="text-[11px] text-text-muted flex items-center gap-0.5 opacity-0 group-focus-within:opacity-100 transition-opacity duration-200">
              <CornerDownLeft size={10} /> 发送
            </span>
            <button
              onClick={handleSend}
              disabled={!input.trim() || sending}
              className={cn(
                "btn-press px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 flex items-center gap-1 focus-ring",
                input.trim() && !sending
                  ? "bg-accent text-white hover:bg-accent-hover shadow-sm"
                  : "bg-border-subtle text-text-muted cursor-not-allowed"
              )}
            >
              {sending ? <Loader2 size={13} className="animate-spin" /> : <Send size={13} />}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ---- Structured report card for assistant messages ---- */
function MessageBlock({ message, index, toolCalls }: { message: AgentMessage; index: number; toolCalls: import("../types").AgentToolCall[] }) {
  const isUser = message.role === "user";
  const delay = Math.min(index * 60, 300);

  if (isUser) {
    return (
      <div className="flex justify-end animate-slide-up" style={{ animationDelay: `${delay}ms`, animationFillMode: "both" }}>
        <div className="max-w-[70%] bg-accent text-white rounded-2xl rounded-br-md px-4 py-3 shadow-sm">
          <p className="text-sm whitespace-pre-wrap leading-relaxed">{message.content}</p>
          <span className="text-[11px] text-white/50 mt-1.5 block">{formatDate(message.created_at)}</span>
        </div>
      </div>
    );
  }

  // Find tool calls that are associated with nearby assistant messages
  const relatedCalls = toolCalls; // In a real app, filter by message_id proximity

  return (
    <div className="flex gap-3 items-start animate-slide-up" style={{ animationDelay: `${delay}ms`, animationFillMode: "both" }}>
      <div className="w-8 h-8 rounded-lg bg-cyan-soft border border-cyan/15 flex items-center justify-center shrink-0 mt-0.5">
        <Bot size={14} className="text-cyan" />
      </div>
      <div className="flex-1 min-w-0 max-w-[85%]">
        <DiagnosisReport message={message} toolCalls={relatedCalls} />
        <span className="text-[11px] text-text-muted mt-1.5 ml-1 block">{formatDate(message.created_at)}</span>
      </div>
    </div>
  );
}

function DiagnosisReport({ message, toolCalls }: { message: AgentMessage; toolCalls: import("../types").AgentToolCall[] }) {
  const [evidenceOpen, setEvidenceOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(message.content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch { /* fallback: do nothing */ }
  }, [message.content]);

  const handleExport = useCallback(() => {
    const lines = [`# ${message.content_summary || "诊断报告"}`, "", message.content];
    if (toolCalls.length > 0) {
      lines.push("", "---", `## 证据链 (${toolCalls.length} 次工具调用)`, "");
      toolCalls.forEach((tc, i) => {
        lines.push(`### ${i + 1}. ${toolStepLabels[tc.tool_name] || tc.tool_name}`);
        if (tc.result_summary) lines.push("", tc.result_summary);
        lines.push("", `> 耗时: ${formatDuration(tc.latency_ms)} | 状态: ${tc.status}`, "");
      });
    }
    const blob = new Blob([lines.join("\n")], { type: "text/markdown;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `诊断报告_${formatDate(message.created_at).replace(/[^\w]/g, "_")}.md`;
    a.click();
    URL.revokeObjectURL(url);
  }, [message, toolCalls]);

  return (
    <div className="report-card relative bg-surface-raised border border-border-subtle rounded-xl overflow-hidden">
      {/* Left accent bar */}
      <div className="absolute left-0 top-0 bottom-0 w-[3px] bg-gradient-to-b from-cyan via-accent to-cyan/30 rounded-l-xl" />

      {/* Report header */}
      <div className="flex items-center gap-2 pl-5 pr-4 py-2.5 border-b border-border-subtle bg-surface-overlay/60">
        <FileText size={13} className="text-cyan" />
        <span className="text-xs font-medium text-text-primary">诊断报告</span>
        {message.content_summary && (
          <span className="text-[11px] text-text-tertiary bg-surface px-1.5 py-0.5 rounded ml-1">
            {message.content_summary}
          </span>
        )}
        <div className="flex-1" />
        {/* Action buttons */}
        <div className="flex items-center gap-0.5">
          <button
            onClick={handleCopy}
            className="btn-press p-1 rounded hover:bg-surface-overlay text-text-tertiary hover:text-text-secondary transition-colors focus-ring"
            title="复制报告内容"
            aria-label="复制报告内容"
          >
            {copied ? <Check size={12} className="text-emerald" /> : <Copy size={12} />}
          </button>
          <button
            onClick={handleExport}
            className="btn-press p-1 rounded hover:bg-surface-overlay text-text-tertiary hover:text-text-secondary transition-colors focus-ring"
            title="导出为 Markdown"
            aria-label="导出为 Markdown"
          >
            <Download size={12} />
          </button>
        </div>
        {toolCalls.length > 0 && (
          <div className="flex items-center gap-1 text-[11px] text-emerald font-medium">
            <ShieldCheck size={11} />
            <span>{toolCalls.length} 项证据</span>
          </div>
        )}
      </div>

      {/* Report conclusion - Markdown rendered */}
      <div className="pl-5 pr-4 py-3 border-b border-border-subtle/50">
        <div className="flex items-start gap-2">
          <CheckCircle2 size={14} className="text-accent mt-0.5 shrink-0" />
          <div className="report-prose flex-1">
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{message.content}</ReactMarkdown>
          </div>
        </div>
      </div>

      {/* Evidence chain (collapsible) */}
      {toolCalls.length > 0 && (
        <div>
          <button
            onClick={() => setEvidenceOpen(!evidenceOpen)}
            className="w-full flex items-center gap-2 pl-5 pr-4 py-2 text-xs text-text-tertiary hover:text-text-secondary hover:bg-surface-overlay/40 transition-colors"
          >
            <Link2 size={12} />
            <span className="font-medium">证据链</span>
            <span className="text-text-muted">({toolCalls.length} 次工具调用)</span>
            <div className="flex-1" />
            {evidenceOpen ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
          </button>
          <div className={cn("collapse-wrapper", evidenceOpen && "open")}>
            <div className="collapse-inner">
              <div className="pl-5 pr-4 pb-3 space-y-1.5">
                {toolCalls.map((tc) => (
                  <div key={tc.id} className="flex items-start gap-2 text-xs py-1.5 px-2.5 rounded-md bg-surface-overlay/60">
                    <div className={cn(
                      "w-4 h-4 rounded-full flex items-center justify-center shrink-0 mt-0.5",
                      tc.status === "success" ? "bg-emerald-soft" : tc.status === "error" ? "bg-rose-soft" : "bg-amber-soft"
                    )}>
                      {tc.status === "success" ? <CheckCircle2 size={9} className="text-emerald" /> :
                       tc.status === "error" ? <AlertTriangle size={9} className="text-rose" /> :
                       <Clock size={9} className="text-amber" />}
                    </div>
                    <div className="flex-1 min-w-0">
                      <span className="font-mono font-medium text-text-primary">{toolStepLabels[tc.tool_name] || tc.tool_name}</span>
                      {tc.result_summary && (
                        <p className="text-text-tertiary mt-0.5 line-clamp-2 leading-relaxed">{tc.result_summary}</p>
                      )}
                    </div>
                    <span className="text-[11px] text-text-muted font-mono tabular-nums shrink-0">{formatDuration(tc.latency_ms)}</span>
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

const toolStepLabels: Record<string, string> = {
  get_video_detail: "视频详情",
  get_hot_videos: "热榜数据",
  get_video_comments: "评论采集",
  get_author_profile: "作者信息",
  list_author_videos: "作者作品",
  list_tag_videos: "标签视频",
  analyze_comment_risk: "风险分析",
};

function SendingIndicator() {
  return (
    <div className="flex gap-3 items-start animate-slide-up">
      <div className="w-8 h-8 rounded-lg bg-cyan-soft border border-cyan/15 flex items-center justify-center shrink-0">
        <Bot size={14} className="text-cyan" />
      </div>
      <div className="bg-surface-raised border border-border-subtle rounded-xl px-4 py-4 shadow-sm">
        <div className="space-y-2.5 w-64">
          <div className="flex items-center gap-2 mb-3">
            <Loader2 size={13} className="text-accent animate-spin" />
            <span className="text-xs text-text-secondary">Agent 正在调用工具分析数据...</span>
          </div>
          <div className="h-2.5 skeleton w-full" />
          <div className="h-2.5 skeleton w-4/5" />
          <div className="h-2.5 skeleton w-3/5" />
        </div>
      </div>
    </div>
  );
}

function EmptyMessagesPlaceholder() {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center animate-fade-in">
      <div className="w-14 h-14 rounded-xl bg-accent-soft border border-accent-border flex items-center justify-center mb-5 shadow-sm">
        <Sparkles size={24} className="text-accent" />
      </div>
      <h3 className="text-base font-semibold text-text-primary mb-1.5">开始诊断分析</h3>
      <p className="text-sm text-text-tertiary max-w-xs leading-relaxed mb-6">
        输入运营问题，Agent 将调用平台工具采集数据，<br />并生成基于证据的诊断报告。
      </p>
      <div className="flex items-center gap-6 text-[11px] text-text-tertiary">
        <div className="flex items-center gap-1.5">
          <Zap size={12} className="text-amber" />
          <span>工具调用</span>
        </div>
        <div className="flex items-center gap-1.5">
          <BarChart3 size={12} className="text-cyan" />
          <span>证据追踪</span>
        </div>
        <div className="flex items-center gap-1.5">
          <Lightbulb size={12} className="text-accent" />
          <span>诊断报告</span>
        </div>
      </div>
    </div>
  );
}
