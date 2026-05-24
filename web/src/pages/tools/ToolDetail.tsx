import { useState, useEffect, useCallback } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import {
  ArrowLeft,
  Wrench,
  Play,
  Clock,
  Shield,
  ShieldOff,
} from "lucide-react";
import type { Tool, ToolCallResult, Invocation, ToolParamDef } from "../../types";
import { gatewayApi } from "../../lib/api";
import { Select } from "../../components/ui/Select";
import { cn, formatDuration, formatDate } from "../../lib/utils";

export function ToolDetail() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();

  const [tool, setTool] = useState<Tool | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  // Manual call state
  const [args, setArgs] = useState<Record<string, unknown>>({});
  const [source, setSource] = useState("manual_console");
  const [sessionId, setSessionId] = useState("");
  const [skillId, setSkillId] = useState("");
  const [calling, setCalling] = useState(false);
  const [callResult, setCallResult] = useState<ToolCallResult | null>(null);
  const [callError, setCallError] = useState("");

  // Recent invocations
  const [recentInvocations, setRecentInvocations] = useState<Invocation[]>([]);

  const loadTool = useCallback(() => {
    if (!name) return;
    setLoading(true);
    setError("");
    gatewayApi
      .getTool(name)
      .then((res) => {
        setTool(res.tool);
        // Initialize args with defaults
        const defaults: Record<string, unknown> = {};
        const props = res.tool.schema.function.parameters.properties;
        if (props) {
          for (const [key, def] of Object.entries(props)) {
            if (def.default !== undefined) {
              defaults[key] = def.default;
            }
          }
        }
        setArgs(defaults);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [name]);

  const loadRecentInvocations = useCallback(() => {
    if (!name) return;
    gatewayApi
      .listInvocations({ tool_name: name, limit: 5 })
      .then((res) => setRecentInvocations(res.invocations))
      .catch(() => {});
  }, [name]);

  useEffect(() => {
    loadTool();
  }, [loadTool]);

  useEffect(() => {
    loadRecentInvocations();
  }, [loadRecentInvocations]);

  const handleInvoke = async () => {
    if (!name) return;
    setCalling(true);
    setCallError("");
    setCallResult(null);

    try {
      const body: {
        source?: string;
        session_id?: number;
        skill_id?: string;
        arguments: Record<string, unknown>;
      } = { source, arguments: args };
      if (sessionId) body.session_id = Number(sessionId);
      if (skillId) body.skill_id = skillId;

      const result = await gatewayApi.callTool(name, body);
      setCallResult(result);
      loadRecentInvocations();
    } catch (err: unknown) {
      setCallError(
        err instanceof Error ? err.message : "调用失败"
      );
    } finally {
      setCalling(false);
    }
  };

  const handleArgChange = (key: string, value: unknown) => {
    setArgs((prev) => ({ ...prev, [key]: value }));
  };

  const parameters = tool?.schema.function.parameters;
  const properties = parameters?.properties ?? {};
  const required = parameters?.required ?? [];

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="skeleton h-8 w-48" />
        <div className="skeleton h-4 w-64" />
        <div className="skeleton h-40 w-full" />
      </div>
    );
  }

  if (error || !tool) {
    return (
      <div className="space-y-4">
        <Link
          to="/tools"
          className="console-btn console-btn-ghost text-sm inline-flex items-center gap-1"
        >
          <ArrowLeft className="w-4 h-4" />
          返回工具列表
        </Link>
        <div className="badge badge-error text-sm">
          {error || "工具未找到"}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Back link */}
      <Link
        to="/tools"
        className="console-btn console-btn-ghost text-sm inline-flex items-center gap-1"
      >
        <ArrowLeft className="w-4 h-4" />
        返回工具列表
      </Link>

      {/* 工具信息 */}
      <div className="console-card-static p-6 space-y-4">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
              {tool.display_name}
            </h1>
            <div className="flex items-center gap-2">
              <span className="badge badge-accent">{tool.category}</span>
              <span
                className={cn(
                  "badge inline-flex items-center gap-1",
                  tool.read_only ? "badge-success" : "badge-warning"
                )}
              >
                {tool.read_only ? (
                  <Shield className="w-3 h-3" />
                ) : (
                  <ShieldOff className="w-3 h-3" />
                )}
                {tool.read_only ? "只读" : "可写"}
              </span>
            </div>
          </div>
          <Wrench className="w-6 h-6 text-[var(--color-text-muted)]" />
        </div>
        <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          {tool.description}
        </p>
      </div>

      {/* JSON Schema */}
      <div className="console-card-static p-6 space-y-3">
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)] uppercase tracking-wide">
          JSON Schema
        </h2>
        <div className="console-code">
          {JSON.stringify(tool.schema.function.parameters, null, 2)}
        </div>
      </div>

      {/* Manual Call Form */}
      <div className="console-card-static p-6 space-y-4">
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)] uppercase tracking-wide">
          手动调用
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {/* Parameter fields */}
          {Object.entries(properties).map(([paramName, paramDef]) => {
            const def = paramDef as ToolParamDef;
            const isRequired = required.includes(paramName);
            return (
              <div key={paramName} className="space-y-1">
                <label className="text-sm font-medium text-[var(--color-text-primary)] flex items-center gap-1">
                  {paramName}
                  <span className="text-[10px] text-[var(--color-text-tertiary)]">
                    ({def.type})
                  </span>
                  {isRequired && (
                    <span className="text-[var(--color-rose)]">*</span>
                  )}
                </label>
                {def.enum ? (
                  <select
                    value={String(args[paramName] ?? "")}
                    onChange={(e) => handleArgChange(paramName, e.target.value)}
                    className="console-select w-full"
                  >
                    <option value="">-- 请选择 --</option>
                    {def.enum.map((opt) => (
                      <option key={opt} value={opt}>
                        {opt}
                      </option>
                    ))}
                  </select>
                ) : def.type === "integer" || def.type === "number" ? (
                  <input
                    type="number"
                    value={String(args[paramName] ?? "")}
                    onChange={(e) =>
                      handleArgChange(
                        paramName,
                        def.type === "integer"
                          ? parseInt(e.target.value, 10) || 0
                          : parseFloat(e.target.value) || 0
                      )
                    }
                    placeholder={def.description || paramName}
                    className="console-input"
                  />
                ) : (
                  <input
                    type="text"
                    value={String(args[paramName] ?? "")}
                    onChange={(e) => handleArgChange(paramName, e.target.value)}
                    placeholder={def.description || paramName}
                    className="console-input"
                  />
                )}
                {def.description && (
                  <p className="text-[11px] text-[var(--color-text-tertiary)]">
                    {def.description}
                  </p>
                )}
              </div>
            );
          })}

          {/* Source selector */}
          <div className="space-y-1">
            <label className="text-sm font-medium text-[var(--color-text-primary)]">
              来源
            </label>
            <Select
              value={source}
              onChange={(v) => setSource(v)}
              options={[
                { value: "manual_console", label: "手动控制台" },
                { value: "agent_runtime", label: "Agent 运行时" },
                { value: "mcp_client", label: "MCP 客户端" },
              ]}
              className="w-full"
            />
          </div>

          {/* Optional session_id */}
          <div className="space-y-1">
            <label className="text-sm font-medium text-[var(--color-text-primary)]">
              会话 ID{" "}
              <span className="text-[10px] text-[var(--color-text-tertiary)]">
                (可选)
              </span>
            </label>
            <input
              type="number"
              value={sessionId}
              onChange={(e) => setSessionId(e.target.value)}
              placeholder="会话 ID"
              className="console-input"
            />
          </div>

          {/* Optional skill_id */}
          <div className="space-y-1">
            <label className="text-sm font-medium text-[var(--color-text-primary)]">
              技能 ID{" "}
              <span className="text-[10px] text-[var(--color-text-tertiary)]">
                (可选)
              </span>
            </label>
            <input
              type="text"
              value={skillId}
              onChange={(e) => setSkillId(e.target.value)}
              placeholder="技能 ID"
              className="console-input"
            />
          </div>
        </div>

        {/* Invoke button */}
        <div className="flex items-center gap-3 pt-2">
          <button
            onClick={handleInvoke}
            disabled={calling}
            className={cn(
              "console-btn console-btn-primary btn-press",
              calling && "opacity-60 cursor-not-allowed"
            )}
          >
            <Play className="w-4 h-4" />
            {calling ? "调用中..." : "调用工具"}
          </button>
          {calling && (
            <span className="text-sm text-[var(--color-text-tertiary)] animate-pulse-soft">
              正在调用 {tool.display_name}...
            </span>
          )}
        </div>
      </div>

      {/* Result Section */}
      {(callResult || callError) && (
        <div className="console-card-static p-6 space-y-4 animate-slide-up">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-semibold text-[var(--color-text-primary)] uppercase tracking-wide">
              结果
            </h2>
            {callResult && (
              <div className="flex items-center gap-3">
                <span
                  className={cn(
                    "badge",
                    callResult.invocation.status === "success"
                      ? "badge-success"
                      : "badge-error"
                  )}
                >
                  {callResult.invocation.status === "success" ? "成功" : callResult.invocation.status === "error" ? "错误" : callResult.invocation.status === "timeout" ? "超时" : callResult.invocation.status}
                </span>
                <span className="text-xs text-[var(--color-text-tertiary)] flex items-center gap-1">
                  <Clock className="w-3 h-3" />
                  {formatDuration(callResult.invocation.latency_ms)}
                </span>
              </div>
            )}
          </div>

          {callError && (
            <div className="badge badge-error text-sm">{callError}</div>
          )}

          {callResult && (
            <>
              {/* Summary */}
              <div className="space-y-1">
                <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                  摘要
                </span>
                <p className="text-sm text-[var(--color-text-secondary)]">
                  {callResult.result.summary}
                </p>
              </div>

              {/* Result data */}
              <div className="space-y-2">
                <span className="text-xs font-medium text-[var(--color-text-tertiary)] uppercase tracking-wide">
                  结果数据
                </span>
                <div className="console-code">
                  {JSON.stringify(callResult.result.data, null, 2)}
                </div>
              </div>

              {/* Invocation metadata */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3 text-sm">
                <div>
                  <span className="text-[var(--color-text-tertiary)] text-xs">
                    调用 ID
                  </span>
                  <p className="text-[var(--color-text-primary)] font-medium">
                    #{callResult.invocation.id}
                  </p>
                </div>
                <div>
                  <span className="text-[var(--color-text-tertiary)] text-xs">
                    来源
                  </span>
                  <p className="text-[var(--color-text-primary)] font-medium">
                    {callResult.invocation.source}
                  </p>
                </div>
                <div>
                  <span className="text-[var(--color-text-tertiary)] text-xs">
                    耗时
                  </span>
                  <p className="text-[var(--color-text-primary)] font-medium">
                    {formatDuration(callResult.invocation.latency_ms)}
                  </p>
                </div>
                <div>
                  <span className="text-[var(--color-text-tertiary)] text-xs">
                    时间
                  </span>
                  <p className="text-[var(--color-text-primary)] font-medium">
                    {formatDate(callResult.invocation.created_at)}
                  </p>
                </div>
              </div>
            </>
          )}
        </div>
      )}

      {/* Recent Invocations */}
      <div className="console-card-static p-6 space-y-4">
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)] uppercase tracking-wide">
          最近调用
        </h2>

        {recentInvocations.length === 0 ? (
          <p className="text-sm text-[var(--color-text-tertiary)] py-4 text-center">
            该工具暂无调用记录
          </p>
        ) : (
          <div className="overflow-x-auto">
            <table className="console-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>来源</th>
                  <th>状态</th>
                  <th>耗时</th>
                  <th>摘要</th>
                  <th>时间</th>
                </tr>
              </thead>
              <tbody>
                {recentInvocations.map((inv) => (
                  <tr
                    key={inv.id}
                    className="cursor-pointer"
                    onClick={() => navigate(`/invocations`)}
                  >
                    <td className="font-mono text-xs">#{inv.id}</td>
                    <td>
                      <SourceBadge source={inv.source} />
                    </td>
                    <td>
                      <StatusBadge status={inv.status} />
                    </td>
                    <td className="font-mono text-xs">
                      {formatDuration(inv.latency_ms)}
                    </td>
                    <td className="max-w-[200px] truncate text-xs">
                      {inv.result_summary || "-"}
                    </td>
                    <td className="text-xs whitespace-nowrap">
                      {formatDate(inv.created_at)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
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
