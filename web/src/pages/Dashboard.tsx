import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  Activity,
  MessageSquare,
  Wrench,
  Zap,
  Clock,
  ArrowRight,
  AlertCircle,
  Server,
} from "lucide-react";
import type { AgentSession, Invocation } from "../types";
import { sessionApi, gatewayApi, skillApi } from "../lib/api";
import { cn, formatDate } from "../lib/utils";
import { useConnectionStatus } from "../hooks/useConnectionStatus";

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(2)}s`;
}

interface DashboardData {
  sessions: AgentSession[];
  tools: { name: string }[];
  skills: { status: string }[];
  invocations: Invocation[];
  recentSessions: AgentSession[];
  recentInvocations: Invocation[];
}

function SessionStatusBadge({ status }: { status: string }) {
  const cls =
    status === "active"
      ? "badge badge-success"
      : status === "error"
        ? "badge badge-error"
        : "badge badge-neutral";
  const statusMap: Record<string, string> = {
    active: "活跃",
    closed: "已关闭",
    error: "错误",
  };
  return <span className={cls}>{statusMap[status] ?? status}</span>;
}

function InvocationStatusBadge({ status }: { status: string }) {
  const cls =
    status === "success"
      ? "badge badge-success"
      : status === "error"
        ? "badge badge-error"
        : "badge badge-neutral";
  const statusMap: Record<string, string> = {
    success: "成功",
    error: "错误",
    timeout: "超时",
  };
  return <span className={cls}>{statusMap[status] ?? status}</span>;
}

function SourceBadge({ source }: { source: string }) {
  const cls =
    source === "manual_console"
      ? "badge badge-accent"
      : source === "agent_runtime"
        ? "badge badge-cyan"
        : "badge badge-neutral";
  const sourceMap: Record<string, string> = {
    manual_console: "手动控制台",
    agent_runtime: "Agent 运行时",
    mcp_client: "MCP 客户端",
  };
  return <span className={cls}>{sourceMap[source] ?? source.replace(/_/g, " ")}</span>;
}

function MetricCard({
  icon: Icon,
  label,
  value,
  loading,
}: {
  icon: React.ElementType;
  label: string;
  value: number;
  loading: boolean;
}) {
  return (
    <div className="console-card metric-card p-5">
      <div className="flex items-center gap-3 mb-3">
        <div className="flex items-center justify-center w-9 h-9 rounded-lg bg-[var(--color-accent-soft)]">
          <Icon className="w-[18px] h-[18px] text-[var(--color-accent)]" />
        </div>
        <span className="text-sm font-medium text-[var(--color-text-secondary)]">
          {label}
        </span>
      </div>
      {loading ? (
        <div className="skeleton h-9 w-20" />
      ) : (
        <span className="text-3xl font-bold text-[var(--color-text-primary)] tabular-nums">
          {value}
        </span>
      )}
    </div>
  );
}

function SkeletonRow({ cols }: { cols: number }) {
  return (
    <tr>
      {Array.from({ length: cols }).map((_, i) => (
        <td key={i}>
          <div className="skeleton h-4 w-full max-w-[140px]" />
        </td>
      ))}
    </tr>
  );
}

export function Dashboard() {
  const navigate = useNavigate();
  const connectionStatus = useConnectionStatus();

  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function fetchData() {
      setLoading(true);
      setError(null);
      try {
        const [
          sessionsRes,
          toolsRes,
          skillsRes,
          invocationsRes,
          recentSessionsRes,
          recentInvocationsRes,
        ] = await Promise.all([
          sessionApi.list({ limit: 100 }),
          gatewayApi.listTools(),
          skillApi.list(),
          gatewayApi.listInvocations({ limit: 1000 }),
          sessionApi.list({ limit: 5 }),
          gatewayApi.listInvocations({ limit: 5 }),
        ]);

        if (cancelled) return;

        setData({
          sessions: sessionsRes.sessions,
          tools: toolsRes.tools,
          skills: skillsRes.skills,
          invocations: invocationsRes.invocations,
          recentSessions: recentSessionsRes.sessions,
          recentInvocations: recentInvocationsRes.invocations,
        });
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "加载失败");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    fetchData();
    return () => { cancelled = true; };
  }, []);

  const activeSessions =
    data?.sessions.filter((s) => s.status === "active").length ?? 0;
  const totalTools = data?.tools.length ?? 0;
  const enabledSkills =
    data?.skills.filter((s) => s.status === "enabled").length ?? 0;
  const totalInvocations = data?.invocations.length ?? 0;

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
          总览
        </h1>
        <p className="text-sm text-[var(--color-text-secondary)] mt-1">
          VideoOps Agent 控制台总览
        </p>
      </div>

      {/* Backend Status */}
      <div className="console-card p-4 flex items-center gap-3">
        <div
          className={cn(
            "status-dot",
            connectionStatus === "online" && "status-dot-online",
            connectionStatus === "offline" && "status-dot-offline",
            connectionStatus === "checking" && "status-dot-checking",
          )}
        />
        <Server className="w-4 h-4 text-[var(--color-text-tertiary)]" />
        <span className="text-sm font-medium text-[var(--color-text-primary)]">
          {connectionStatus === "online" && "后端在线"}
          {connectionStatus === "offline" && "后端离线"}
          {connectionStatus === "checking" && "检测中..."}
        </span>
      </div>

      {/* Error State */}
      {error && (
        <div className="console-card p-4 flex items-start gap-3 border-[var(--color-rose)]">
          <AlertCircle className="w-5 h-5 text-[var(--color-rose)] shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-[var(--color-rose)]">
              加载失败
            </p>
            <p className="text-sm text-[var(--color-text-secondary)] mt-1">
              {error}
            </p>
          </div>
        </div>
      )}

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard
          icon={MessageSquare}
          label="活跃会话"
          value={activeSessions}
          loading={loading}
        />
        <MetricCard
          icon={Wrench}
          label="工具总数"
          value={totalTools}
          loading={loading}
        />
        <MetricCard
          icon={Zap}
          label="已启用技能"
          value={enabledSkills}
          loading={loading}
        />
        <MetricCard
          icon={Activity}
          label="调用总数"
          value={totalInvocations}
          loading={loading}
        />
      </div>

      {/* Recent Sessions */}
      <div className="console-card overflow-hidden">
        <div className="flex items-center justify-between p-5 border-b border-[var(--color-border-subtle)]">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] flex items-center gap-2">
            <Clock className="w-4 h-4 text-[var(--color-text-tertiary)]" />
            最近会话
          </h2>
          <button
            onClick={() => navigate("/sessions")}
            className="console-btn console-btn-ghost text-xs"
          >
            查看全部
            <ArrowRight className="w-3.5 h-3.5" />
          </button>
        </div>

        <table className="console-table">
          <thead>
            <tr>
              <th>ID</th>
              <th>标题</th>
              <th>状态</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <>
                <SkeletonRow cols={4} />
                <SkeletonRow cols={4} />
                <SkeletonRow cols={4} />
              </>
            ) : data && data.recentSessions.length === 0 ? (
              <tr>
                <td colSpan={4} className="text-center py-8">
                  <p className="text-sm text-[var(--color-text-tertiary)]">
                    暂无数据
                  </p>
                </td>
              </tr>
            ) : (
              data?.recentSessions.map((session) => (
                <tr
                  key={session.id}
                  className="cursor-pointer hover:bg-[var(--color-surface-overlay)] transition-colors"
                  onClick={() => navigate(`/sessions/${session.id}`)}
                >
                  <td className="font-mono text-[var(--color-text-primary)]">
                    #{session.id}
                  </td>
                  <td className="text-[var(--color-text-primary)] max-w-[240px] truncate">
                    {session.title || session.scenario || "无标题"}
                  </td>
                  <td>
                    <SessionStatusBadge status={session.status} />
                  </td>
                  <td className="text-[var(--color-text-tertiary)] whitespace-nowrap">
                    {formatDate(session.created_at)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Recent Invocations */}
      <div className="console-card overflow-hidden">
        <div className="flex items-center justify-between p-5 border-b border-[var(--color-border-subtle)]">
          <h2 className="text-base font-semibold text-[var(--color-text-primary)] flex items-center gap-2">
            <Activity className="w-4 h-4 text-[var(--color-text-tertiary)]" />
            最近调用
          </h2>
          <button
            onClick={() => navigate("/invocations")}
            className="console-btn console-btn-ghost text-xs"
          >
            查看全部
            <ArrowRight className="w-3.5 h-3.5" />
          </button>
        </div>

        <table className="console-table">
          <thead>
            <tr>
              <th>工具</th>
              <th>来源</th>
              <th>状态</th>
              <th>耗时</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <>
                <SkeletonRow cols={5} />
                <SkeletonRow cols={5} />
                <SkeletonRow cols={5} />
              </>
            ) : data && data.recentInvocations.length === 0 ? (
              <tr>
                <td colSpan={5} className="text-center py-8">
                  <p className="text-sm text-[var(--color-text-tertiary)]">
                    暂无数据
                  </p>
                </td>
              </tr>
            ) : (
              data?.recentInvocations.map((inv) => (
                <tr key={inv.id}>
                  <td className="font-mono text-[var(--color-text-primary)]">
                    {inv.tool_name}
                  </td>
                  <td>
                    <SourceBadge source={inv.source} />
                  </td>
                  <td>
                    <InvocationStatusBadge status={inv.status} />
                  </td>
                  <td className="text-[var(--color-text-secondary)] tabular-nums">
                    {formatDuration(inv.latency_ms)}
                  </td>
                  <td className="text-[var(--color-text-tertiary)] whitespace-nowrap">
                    {formatDate(inv.created_at)}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
