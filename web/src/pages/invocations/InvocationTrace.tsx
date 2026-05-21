import { useState, useEffect, useCallback } from "react";
import { Activity, ChevronDown, ChevronRight, RotateCcw } from "lucide-react";
import type { Invocation, InvocationFilters } from "../../types";
import { gatewayApi } from "../../lib/api";
import { Select } from "../../components/ui/Select";
import { cn, formatDate, formatDuration, safeJSONParse } from "../../lib/utils";

export function InvocationTrace() {
  const [invocations, setInvocations] = useState<Invocation[]>([]);
  const [filters, setFilters] = useState<InvocationFilters>({ limit: 20 });
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  // Filter form state (separate from applied filters)
  const [filterSource, setFilterSource] = useState("");
  const [filterStatus, setFilterStatus] = useState("");
  const [filterToolName, setFilterToolName] = useState("");
  const [filterSessionId, setFilterSessionId] = useState("");
  const [filterSkillId, setFilterSkillId] = useState("");

  const fetchInvocations = useCallback(
    async (appliedFilters: InvocationFilters) => {
      setLoading(true);
      try {
        const res = await gatewayApi.listInvocations(appliedFilters);
        setInvocations(res.invocations);
      } catch {
        // Silently handle — table will just be empty
      } finally {
        setLoading(false);
      }
    },
    []
  );

  useEffect(() => {
    fetchInvocations(filters);
  }, [filters, fetchInvocations]);

  const handleApply = () => {
    const newFilters: InvocationFilters = { limit: filters.limit ?? 20 };
    if (filterSource) newFilters.source = filterSource;
    if (filterStatus) newFilters.status = filterStatus;
    if (filterToolName) newFilters.tool_name = filterToolName;
    if (filterSessionId) newFilters.session_id = Number(filterSessionId);
    if (filterSkillId) newFilters.skill_id = filterSkillId;
    setFilters(newFilters);
  };

  const handleReset = () => {
    setFilterSource("");
    setFilterStatus("");
    setFilterToolName("");
    setFilterSessionId("");
    setFilterSkillId("");
    setFilters({ limit: 20 });
  };

  const handleLoadMore = () => {
    const newLimit = (filters.limit ?? 20) + 20;
    setFilters({ ...filters, limit: newLimit });
  };

  const toggleExpand = (id: number) => {
    setExpandedId(expandedId === id ? null : id);
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
          调用追踪
        </h1>
        <p className="text-sm text-[var(--color-text-tertiary)] mt-1">
          统一工具调用历史
        </p>
      </div>

      {/* Filter bar */}
      <div className="console-card-static p-4">
        <div className="flex flex-wrap items-end gap-3">
          {/* Source */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-[var(--color-text-tertiary)]">
              来源
            </label>
            <Select
              value={filterSource}
              onChange={(v) => setFilterSource(v)}
              options={[
                { value: "", label: "全部" },
                { value: "manual_console", label: "手动控制台" },
                { value: "agent_runtime", label: "Agent 运行时" },
                { value: "mcp_client", label: "MCP 客户端" },
              ]}
              placeholder="全部"
            />
          </div>

          {/* Status */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-[var(--color-text-tertiary)]">
              状态
            </label>
            <Select
              value={filterStatus}
              onChange={(v) => setFilterStatus(v)}
              options={[
                { value: "", label: "全部" },
                { value: "success", label: "成功" },
                { value: "error", label: "错误" },
                { value: "timeout", label: "超时" },
              ]}
              placeholder="全部"
            />
          </div>

          {/* Tool name */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-[var(--color-text-tertiary)]">
              工具名称
            </label>
            <input
              type="text"
              value={filterToolName}
              onChange={(e) => setFilterToolName(e.target.value)}
              placeholder="按工具筛选..."
              className="console-input w-40"
            />
          </div>

          {/* Session ID */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-[var(--color-text-tertiary)]">
              会话 ID
            </label>
            <input
              type="number"
              value={filterSessionId}
              onChange={(e) => setFilterSessionId(e.target.value)}
              placeholder="会话 ID"
              className="console-input w-28"
            />
          </div>

          {/* Skill ID */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-[var(--color-text-tertiary)]">
              技能 ID
            </label>
            <input
              type="text"
              value={filterSkillId}
              onChange={(e) => setFilterSkillId(e.target.value)}
              placeholder="技能 ID"
              className="console-input w-36"
            />
          </div>

          {/* Action buttons */}
          <div className="flex items-center gap-2">
            <button
              onClick={handleApply}
              className="console-btn console-btn-primary btn-press"
            >
              应用
            </button>
            <button
              onClick={handleReset}
              className="console-btn console-btn-secondary btn-press inline-flex items-center gap-1"
            >
              <RotateCcw className="w-3 h-3" />
              重置
            </button>
          </div>
        </div>
      </div>

      {/* Loading skeleton */}
      {loading && (
        <div className="console-card-static overflow-hidden">
          <div className="p-4 space-y-3">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="flex items-center gap-4">
                <div className="skeleton h-4 w-12" />
                <div className="skeleton h-4 w-24" />
                <div className="skeleton h-4 w-20" />
                <div className="skeleton h-4 flex-1" />
                <div className="skeleton h-4 w-24" />
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Empty state */}
      {!loading && invocations.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <Activity className="w-10 h-10 text-[var(--color-text-muted)] mb-3" />
          <p className="text-sm text-[var(--color-text-tertiary)]">
            暂无调用记录
          </p>
        </div>
      )}

      {/* Data table */}
      {!loading && invocations.length > 0 && (
        <div className="console-card-static overflow-hidden">
          <div className="overflow-x-auto">
            <table className="console-table">
              <thead>
                <tr>
                  <th style={{ width: 32 }} />
                  <th>ID</th>
                  <th>来源</th>
                  <th>工具</th>
                  <th>状态</th>
                  <th>耗时</th>
                  <th>摘要</th>
                  <th>时间</th>
                </tr>
              </thead>
              <tbody>
                {invocations.map((inv) => (
                  <InvocationRow
                    key={inv.id}
                    invocation={inv}
                    expanded={expandedId === inv.id}
                    onToggle={() => toggleExpand(inv.id)}
                  />
                ))}
              </tbody>
            </table>
          </div>

          {/* Load more */}
          {invocations.length >= (filters.limit ?? 20) && (
            <div className="flex justify-center py-4 border-t border-[var(--color-border-subtle)]">
              <button
                onClick={handleLoadMore}
                className="console-btn console-btn-secondary btn-press text-sm"
              >
                加载更多
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function InvocationRow({
  invocation,
  expanded,
  onToggle,
}: {
  invocation: Invocation;
  expanded: boolean;
  onToggle: () => void;
}) {
  const summary = invocation.result_summary || "-";
  const truncatedSummary =
    summary.length > 50 ? summary.slice(0, 50) + "..." : summary;

  return (
    <>
      <tr className="cursor-pointer" onClick={onToggle}>
        <td className="!py-3">
          {expanded ? (
            <ChevronDown className="w-4 h-4 text-[var(--color-text-muted)]" />
          ) : (
            <ChevronRight className="w-4 h-4 text-[var(--color-text-muted)]" />
          )}
        </td>
        <td className="font-mono text-xs">#{invocation.id}</td>
        <td>
          <SourceBadge source={invocation.source} />
        </td>
        <td className="font-medium text-[var(--color-text-primary)] text-sm">
          {invocation.tool_name}
        </td>
        <td>
          <StatusBadge status={invocation.status} />
        </td>
        <td className="font-mono text-xs">
          {formatDuration(invocation.latency_ms)}
        </td>
        <td className="text-xs max-w-[200px] truncate">{truncatedSummary}</td>
        <td className="text-xs whitespace-nowrap">
          {formatDate(invocation.created_at)}
        </td>
      </tr>

      {/* Expanded detail row */}
      {expanded && (
        <tr>
          <td colSpan={8} className="!py-0 !px-0">
            <div className="px-12 py-4 bg-[var(--color-surface)] space-y-3 animate-expand">
              {/* Full summary */}
              {invocation.result_summary && (
                <div className="space-y-1">
                  <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                    结果摘要
                  </span>
                  <p className="text-sm text-[var(--color-text-secondary)]">
                    {invocation.result_summary}
                  </p>
                </div>
              )}

              {/* Error message */}
              {invocation.status === "error" && invocation.error_message && (
                <div className="space-y-1">
                  <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                    错误
                  </span>
                  <p className="text-sm text-[var(--color-rose)]">
                    {invocation.error_message}
                  </p>
                </div>
              )}

              {/* Arguments JSON */}
              <div className="space-y-1">
                <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                  参数
                </span>
                <div className="console-code">
                  {JSON.stringify(
                    safeJSONParse(invocation.arguments_json, {}),
                    null,
                    2
                  )}
                </div>
              </div>

              {/* Result JSON */}
              {invocation.result_json && (
                <div className="space-y-1">
                  <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                    结果数据
                  </span>
                  <div className="console-code">
                    {JSON.stringify(
                      safeJSONParse(invocation.result_json, {}),
                      null,
                      2
                    )}
                  </div>
                </div>
              )}
            </div>
          </td>
        </tr>
      )}
    </>
  );
}

const sourceLabels: Record<string, string> = {
  manual_console: "手动控制台",
  agent_runtime: "Agent 运行时",
  mcp_client: "MCP 客户端",
};

function SourceBadge({ source }: { source: string }) {
  const cls =
    source === "manual_console"
      ? "badge-accent"
      : source === "agent_runtime"
        ? "badge-cyan"
        : "badge-neutral";
  return <span className={cn("badge", cls)}>{sourceLabels[source] || source}</span>;
}

const statusLabels: Record<string, string> = {
  success: "成功",
  error: "错误",
  timeout: "超时",
};

function StatusBadge({ status }: { status: string }) {
  const cls =
    status === "success"
      ? "badge-success"
      : status === "error"
        ? "badge-error"
        : "badge-warning";
  return <span className={cn("badge", cls)}>{statusLabels[status] || status}</span>;
}
