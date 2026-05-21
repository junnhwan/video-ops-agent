import { useState, useMemo } from "react";
import { cn, formatDuration, safeJSONParse } from "../lib/utils";
import { useSession } from "./SessionProvider";
import type { AgentToolCall } from "../types";
import {
  X,
  CheckCircle2,
  XCircle,
  Clock,
  ChevronDown,
  ChevronRight,
  Terminal,
  Wrench,
  AlertTriangle,
  ArrowRightLeft,
  Zap,
} from "lucide-react";

/* Map tool names to human-friendly pipeline steps */
const toolStepLabels: Record<string, string> = {
  get_video_detail: "视频详情",
  get_hot_videos: "热榜数据",
  get_video_comments: "评论采集",
  get_author_profile: "作者信息",
  list_author_videos: "作者作品",
  list_tag_videos: "标签视频",
  analyze_comment_risk: "风险分析",
};

interface ToolTracePanelProps {
  onClose: () => void;
}

export function ToolTracePanel({ onClose }: ToolTracePanelProps) {
  const { currentSession, toolCalls, messages } = useSession();
  const [expandedId, setExpandedId] = useState<number | null>(null);

  if (!currentSession) return null;

  const sessionCalls = toolCalls.filter(tc => tc.session_id === currentSession.id);
  const hasAssistantReply = messages.some(m => m.session_id === currentSession.id && m.role === "assistant");

  return (
    <aside className="w-[360px] flex flex-col bg-surface-raised border-l border-border-subtle h-full shadow-lg lg:shadow-none">
      {/* Header */}
      <div className="header-bar h-12 shrink-0 px-4 flex items-center justify-between border-b border-border-subtle">
        <div className="flex items-center gap-2">
          <Zap size={14} className="text-cyan" />
          <span className="text-xs font-semibold text-text-primary">工具调用追踪</span>
          <span className="text-[11px] text-text-tertiary bg-surface-overlay px-1.5 py-0.5 rounded">
            {sessionCalls.length} 次调用
          </span>
        </div>
        <button onClick={onClose} className="btn-press p-1 rounded hover:bg-surface-overlay text-text-tertiary transition-colors focus-ring">
          <X size={14} />
        </button>
      </div>

      {/* Pipeline progress bar */}
      {sessionCalls.length > 0 && (
        <div className="px-4 py-3 border-b border-border-subtle bg-surface-overlay/40">
          <div className="text-[11px] font-medium text-text-tertiary uppercase tracking-wider mb-2">执行流程</div>
          <PipelineProgress toolCalls={sessionCalls} hasFinalAnswer={hasAssistantReply} />
        </div>
      )}

      {/* List */}
      <div className="flex-1 overflow-y-auto p-3 scroll-smooth">
        {sessionCalls.length === 0 ? (
          <EmptyTrace />
        ) : (
          <div className="relative">
            <div className="absolute left-[15px] top-3 bottom-3 w-px bg-gradient-to-b from-border-default via-border-subtle to-transparent" />
            <div className="space-y-1.5">
              {sessionCalls.map((tc, i) => (
                <ToolCallNode
                  key={tc.id}
                  toolCall={tc}
                  index={i + 1}
                  isExpanded={expandedId === tc.id}
                  onToggle={() => setExpandedId(expandedId === tc.id ? null : tc.id)}
                />
              ))}
            </div>
          </div>
        )}
      </div>
    </aside>
  );
}

function PipelineProgress({ toolCalls, hasFinalAnswer }: { toolCalls: AgentToolCall[]; hasFinalAnswer: boolean }) {
  const steps = useMemo(() => {
    const seen = new Set<string>();
    const result: { name: string; label: string; status: "done" | "error" | "pending" }[] = [];
    for (const tc of toolCalls) {
      if (!seen.has(tc.tool_name)) {
        seen.add(tc.tool_name);
        result.push({
          name: tc.tool_name,
          label: toolStepLabels[tc.tool_name] || tc.tool_name,
          status: tc.status === "error" ? "error" : "done",
        });
      }
    }
    if (hasFinalAnswer) {
      result.push({ name: "_report", label: "诊断报告", status: "done" });
    }
    return result;
  }, [toolCalls, hasFinalAnswer]);

  if (steps.length === 0) return null;

  return (
    <div className="flex flex-wrap items-center gap-y-1.5 gap-x-0">
      {steps.map((step, i) => (
        <div key={step.name + i} className="flex items-center">
          {/* Step node */}
          <div className="flex items-center gap-1.5 group" title={step.name}>
            <div className={cn(
              "w-5 h-5 rounded-full flex items-center justify-center transition-colors duration-300",
              step.status === "done" && "bg-accent text-white",
              step.status === "error" && "bg-rose text-white",
              step.status === "pending" && "bg-surface-overlay border border-border-default text-text-muted"
            )}>
              {step.status === "done" ? <CheckCircle2 size={11} /> : step.status === "error" ? <XCircle size={11} /> : <Clock size={11} />}
            </div>
            <span className={cn(
              "text-[11px] font-medium transition-colors duration-200",
              step.status === "done" ? "text-text-primary" : "text-text-tertiary"
            )}>
              {step.label}
            </span>
          </div>
          {/* Connector line */}
          {i < steps.length - 1 && (
            <div className={cn(
              "w-4 h-[2px] mx-0.5 rounded-full transition-colors duration-500",
              steps[i + 1]?.status === "done" || steps[i + 1]?.status === "error" ? "bg-accent/30" : "bg-border-default"
            )} />
          )}
        </div>
      ))}
    </div>
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
  const args = safeJSONParse<Record<string, unknown>>(toolCall.arguments_json, {});
  const result = safeJSONParse<unknown>(toolCall.result_json, null);

  const statusConfig: Record<string, { dot: string; icon: React.ReactNode }> = {
    success: { dot: "bg-emerald", icon: <CheckCircle2 size={10} className="text-emerald" /> },
    error:   { dot: "bg-rose",    icon: <XCircle size={10} className="text-rose" /> },
    timeout: { dot: "bg-amber",   icon: <Clock size={10} className="text-amber" /> },
  };
  const status = statusConfig[toolCall.status] || statusConfig.success;
  const friendlyName = toolStepLabels[toolCall.tool_name] || toolCall.tool_name;

  return (
    <div className="relative pl-8 animate-fade-in" style={{ animationDelay: `${(index - 1) * 60}ms`, animationFillMode: "both" }}>
      <div className="absolute left-[11px] top-3.5 z-10">
        <div className={cn(
          "w-[9px] h-[9px] rounded-full border-[2.5px] border-surface-raised transition-colors duration-300",
          status.dot
        )} />
      </div>

      <button
        onClick={onToggle}
        className={cn(
          "btn-press w-full text-left rounded-lg transition-all duration-200 mb-0.5",
          isExpanded
            ? "bg-surface-overlay border border-border-default shadow-sm"
            : "hover:bg-surface-overlay/60 border border-transparent"
        )}
      >
        <div className="px-3 py-2.5 flex items-center gap-2">
          {isExpanded ? <ChevronDown size={12} className="text-text-tertiary" /> : <ChevronRight size={12} className="text-text-muted" />}
          <span className="text-[11px] font-mono text-text-tertiary bg-surface px-1 rounded shrink-0">#{index}</span>
          <span className="text-xs font-medium text-text-primary truncate">{friendlyName}</span>
          <span className="text-[11px] font-mono text-text-muted truncate hidden sm:inline">{toolCall.tool_name}</span>
          <div className="flex-1" />
          {status.icon}
          <span className="text-[11px] text-text-tertiary font-mono tabular-nums">{formatDuration(toolCall.latency_ms)}</span>
        </div>
      </button>

      <div className={cn("collapse-wrapper", isExpanded && "open")}>
        <div className="collapse-inner">
          <div className="ml-2 border-l-2 border-accent/20 pl-3 py-1 space-y-2.5 mb-1.5">
            <div>
              <div className="flex items-center gap-1 mb-1">
                <Wrench size={10} className="text-text-tertiary" />
                <span className="text-[11px] font-medium text-text-tertiary">参数</span>
              </div>
              <pre className="text-[11px] text-text-secondary bg-surface border border-border-subtle rounded-md p-2.5 overflow-x-auto">
                {JSON.stringify(args, null, 2)}
              </pre>
            </div>

            {toolCall.result_summary && (
              <div>
                <div className="flex items-center gap-1 mb-1">
                  <ArrowRightLeft size={10} className="text-cyan" />
                  <span className="text-[11px] font-medium text-text-tertiary">结果摘要</span>
                </div>
                <p className="text-xs text-text-secondary bg-cyan-soft border border-cyan/10 rounded-md p-2.5 leading-relaxed">
                  {toolCall.result_summary}
                </p>
              </div>
            )}

            {result != null && (
              <div>
                <div className="flex items-center gap-1 mb-1">
                  <Terminal size={10} className="text-text-tertiary" />
                  <span className="text-[11px] font-medium text-text-tertiary">原始结果</span>
                </div>
                <pre className="text-[11px] text-text-tertiary bg-surface border border-border-subtle rounded-md p-2.5 overflow-x-auto max-h-36">
                  {JSON.stringify(result, null, 2)}
                </pre>
              </div>
          )}

          {toolCall.error_message && (
            <div>
              <div className="flex items-center gap-1 mb-1">
                <AlertTriangle size={10} className="text-rose" />
                <span className="text-[11px] font-medium text-rose">错误</span>
              </div>
              <p className="text-xs text-rose bg-rose-soft border border-rose/10 rounded-md p-2.5">
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

function EmptyTrace() {
  return (
    <div className="flex flex-col items-center py-10 text-center animate-fade-in">
      <Terminal size={20} className="text-text-muted mb-3" />
      <p className="text-xs text-text-tertiary">暂无工具调用</p>
      <p className="text-[11px] text-text-muted mt-1 max-w-[200px]">Agent 分析时调用的工具将在这里以时间线形式展示</p>
    </div>
  );
}
