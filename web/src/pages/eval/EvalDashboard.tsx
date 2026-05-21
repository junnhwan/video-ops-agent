import { useState, useEffect, useCallback } from "react";
import {
  BarChart3,
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  Wrench,
  Zap,
  TrendingUp,
  Play,
  Loader2,
} from "lucide-react";
import type { EvalSummary, EvalRun, EvalMode, Skill } from "../../types";
import { evalApi, skillApi } from "../../lib/api";
import { cn, formatDuration } from "../../lib/utils";
import { Select } from "../../components/ui/Select";

interface MetricCardProps {
  label: string;
  value: string | number;
  icon: React.ReactNode;
  color: "emerald" | "rose" | "amber" | "accent" | "cyan";
  suffix?: string;
}

function MetricCard({ label, value, icon, color, suffix }: MetricCardProps) {
  const colorMap = {
    emerald: "text-[var(--color-emerald)]",
    rose: "text-[var(--color-rose)]",
    amber: "text-[var(--color-amber)]",
    accent: "text-[var(--color-accent)]",
    cyan: "text-[var(--color-cyan)]",
  };
  return (
    <div className="console-card metric-card p-5">
      <div className="flex items-center justify-between mb-2">
        <span className={cn("text-xs font-medium", colorMap[color])}>{icon}</span>
      </div>
      <div className={cn("text-2xl font-bold tabular-nums", colorMap[color])}>
        {value}
        {suffix && <span className="text-sm font-normal ml-1">{suffix}</span>}
      </div>
      <p className="text-xs text-[var(--color-text-tertiary)] mt-1">{label}</p>
    </div>
  );
}

function renderMetrics(summary: EvalSummary | null) {
  if (!summary) return null;
  const fmt = (v: number | null, suffix = "") =>
    v === null ? "---" : v % 1 !== 0 ? v.toFixed(2) + suffix : String(v) + suffix;

  return (
    <>
      <MetricCard
        label="工具调用成功率"
        value={summary.tool_call_success_rate !== null ? `${(summary.tool_call_success_rate * 100).toFixed(1)}` : "---"}
        icon={<CheckCircle2 className="w-4 h-4" />}
        color="emerald"
        suffix={summary.tool_call_success_rate !== null ? "%" : ""}
      />
      <MetricCard
        label="工具调用错误数"
        value={fmt(summary.tool_call_error_count)}
        icon={<XCircle className="w-4 h-4" />}
        color="rose"
      />
      <MetricCard
        label="未授权工具调用"
        value={fmt(summary.unauthorized_tool_call_count)}
        icon={<AlertTriangle className="w-4 h-4" />}
        color="amber"
      />
      <MetricCard
        label="证据完整回答数"
        value={fmt(summary.evidence_complete_final_answer_count)}
        icon={<CheckCircle2 className="w-4 h-4" />}
        color="emerald"
      />
      <MetricCard
        label="平均工具延迟"
        value={summary.average_tool_latency_ms !== null ? formatDuration(summary.average_tool_latency_ms) : "---"}
        icon={<Clock className="w-4 h-4" />}
        color="cyan"
      />
      <MetricCard
        label="平均工具调用数"
        value={fmt(summary.average_tool_call_count)}
        icon={<Wrench className="w-4 h-4" />}
        color="accent"
      />
      <MetricCard
        label="技能成功次数"
        value={fmt(summary.skill_success_count)}
        icon={<TrendingUp className="w-4 h-4" />}
        color="emerald"
      />
      <MetricCard
        label="技能失败次数"
        value={fmt(summary.skill_failure_count)}
        icon={<XCircle className="w-4 h-4" />}
        color="rose"
      />
    </>
  );
}

export function EvalDashboard() {
  const [globalSummary, setGlobalSummary] = useState<EvalSummary | null>(null);
  const [skillSummary, setSkillSummary] = useState<EvalSummary | null>(null);
  const [skills, setSkills] = useState<Skill[]>([]);
  const [selectedSkill, setSelectedSkill] = useState("");
  const [loadingGlobal, setLoadingGlobal] = useState(true);
  const [loadingSkill, setLoadingSkill] = useState(false);
  const [error, setError] = useState("");

  // Eval run state
  const [runMode, setRunMode] = useState<EvalMode>("skill_guard");
  const [runSkillId, setRunSkillId] = useState("");
  const [running, setRunning] = useState(false);
  const [lastRun, setLastRun] = useState<EvalRun | null>(null);

  const fetchGlobal = useCallback(async () => {
    try {
      setLoadingGlobal(true);
      const res = await evalApi.summary();
      setGlobalSummary(res);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载评估概览失败");
    } finally {
      setLoadingGlobal(false);
    }
  }, []);

  useEffect(() => {
    fetchGlobal();
    skillApi.list().then((res) => setSkills(res.skills || [])).catch(() => {});
  }, [fetchGlobal]);

  const handleSkillSelect = async (skillId: string) => {
    setSelectedSkill(skillId);
    if (!skillId) {
      setSkillSummary(null);
      return;
    }
    try {
      setLoadingSkill(true);
      const res = await evalApi.skillSummary(skillId);
      setSkillSummary(res);
    } catch {
      setSkillSummary(null);
    } finally {
      setLoadingSkill(false);
    }
  };

  const handleCreateRun = async () => {
    if (!runSkillId) return;
    try {
      setRunning(true);
      setError("");
      const res = await evalApi.createRun({ mode: runMode, skill_id: runSkillId });
      setLastRun(res.run);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建评估运行失败");
    } finally {
      setRunning(false);
    }
  };

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
          评估指标
        </h1>
        <p className="text-sm text-[var(--color-text-tertiary)] mt-1">
          Agent 性能分析与评估运行
        </p>
      </div>

      {error && (
        <div className="console-card-static p-4 border-l-4 border-l-[var(--color-rose)] text-sm text-[var(--color-rose)]">
          {error}
        </div>
      )}

      {/* Global Summary */}
      <section>
        <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4 flex items-center gap-2">
          <BarChart3 className="w-5 h-5 text-[var(--color-accent)]" />
          全局概览
        </h2>
        {loadingGlobal ? (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="console-card-static p-5 space-y-2">
                <div className="skeleton h-4 w-8" />
                <div className="skeleton h-8 w-20" />
                <div className="skeleton h-3 w-28" />
              </div>
            ))}
          </div>
        ) : globalSummary ? (
          <>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {renderMetrics(globalSummary)}
            </div>
            {globalSummary.unsupported_metrics && globalSummary.unsupported_metrics.length > 0 && (
              <div className="mt-3 p-3 rounded-lg bg-[var(--color-amber-soft)] text-xs text-[var(--color-amber)]">
                不支持的指标：{globalSummary.unsupported_metrics.join(", ")}
              </div>
            )}
          </>
        ) : (
          <div className="console-card-static p-8 text-center text-[var(--color-text-tertiary)]">
            暂无评估数据
          </div>
        )}
      </section>

      {/* Per-Skill Summary */}
      <section>
        <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4 flex items-center gap-2">
          <Zap className="w-5 h-5 text-[var(--color-cyan)]" />
          单技能概览
        </h2>
        <div className="flex items-center gap-3 mb-4">
          <Select
            value={selectedSkill}
            onChange={(v) => handleSkillSelect(v)}
            options={[
              { value: "", label: "选择技能..." },
              ...skills.map((s) => ({ value: s.id, label: s.name || s.id })),
            ]}
          />
        </div>
        {loadingSkill && (
          <div className="flex items-center gap-2 text-sm text-[var(--color-text-tertiary)]">
            <Loader2 className="w-4 h-4 animate-spin" />
            加载技能指标...
          </div>
        )}
        {!loadingSkill && selectedSkill && skillSummary && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            {renderMetrics(skillSummary)}
          </div>
        )}
        {!loadingSkill && selectedSkill && !skillSummary && (
          <div className="console-card-static p-6 text-center text-sm text-[var(--color-text-tertiary)]">
            此技能暂无评估数据
          </div>
        )}
      </section>

      {/* Create Eval Run */}
      <section>
        <h2 className="text-lg font-semibold text-[var(--color-text-primary)] mb-4 flex items-center gap-2">
          <Play className="w-5 h-5 text-[var(--color-emerald)]" />
          创建评估运行
        </h2>
        <div className="console-card-static p-5">
          <div className="flex items-end gap-4 flex-wrap">
            <div className="flex-1 min-w-[200px]">
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
                模式
              </label>
              <Select
                value={runMode}
                onChange={(v) => setRunMode(v as EvalMode)}
                options={[
                  { value: "baseline", label: "基线模式" },
                  { value: "skill_guard", label: "技能守卫模式" },
                ]}
                className="w-full"
              />
            </div>
            <div className="flex-1 min-w-[200px]">
              <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
                技能 ID
              </label>
              <Select
                value={runSkillId}
                onChange={(v) => setRunSkillId(v)}
                options={[
                  { value: "", label: "选择技能..." },
                  ...skills.map((s) => ({ value: s.id, label: s.name || s.id })),
                ]}
                className="w-full"
              />
            </div>
            <button
              onClick={handleCreateRun}
              disabled={running || !runSkillId}
              className="console-btn console-btn-primary"
            >
              {running ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Play className="w-4 h-4" />
              )}
              运行评估
            </button>
          </div>

          {/* Last run result */}
          {lastRun && (
            <div className="mt-4 p-4 rounded-lg bg-[var(--color-surface)] border border-[var(--color-border-subtle)]">
              <div className="flex items-center gap-2 mb-3">
                <span className="badge badge-accent">运行 #{lastRun.id}</span>
                <span className="badge badge-neutral">{lastRun.mode === "baseline" ? "基线模式" : "技能守卫模式"}</span>
                <span className="text-xs text-[var(--color-text-tertiary)]">
                  {lastRun.skill_id}
                </span>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
                {lastRun.summary && Object.entries(lastRun.summary)
                  .filter(([k]) => k !== "unsupported_metrics")
                  .map(([key, value]) => (
                    <div key={key} className="text-xs">
                      <span className="text-[var(--color-text-tertiary)]">
                        {key.replace(/_/g, " ")}
                      </span>
                      <div className="font-mono font-semibold text-[var(--color-text-primary)] mt-0.5">
                        {value === null ? "---" : String(value)}
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          )}
        </div>
      </section>
    </div>
  );
}
