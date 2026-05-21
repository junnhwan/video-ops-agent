import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import {
  Zap,
  Plus,
  Search,
  CheckCircle2,
  XCircle,
  Wrench,
  FileText,
  Loader2,
} from "lucide-react";
import type { Skill } from "../../types";
import { skillApi } from "../../lib/api";
import { cn } from "../../lib/utils";

export function SkillList() {
  const navigate = useNavigate();
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [statusFilter, setStatusFilter] = useState<"all" | "enabled" | "disabled">("all");
  const [search, setSearch] = useState("");
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const fetchSkills = useCallback(async () => {
    try {
      setLoading(true);
      setError("");
      const res = await skillApi.list();
      setSkills(res.skills || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载技能失败");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSkills();
  }, [fetchSkills]);

  const handleToggle = async (e: React.MouseEvent, skill: Skill) => {
    e.stopPropagation();
    setTogglingId(skill.id);
    try {
      if (skill.status === "enabled") {
        await skillApi.disable(skill.id);
      } else {
        await skillApi.enable(skill.id);
      }
      await fetchSkills();
    } catch (err) {
      setError(err instanceof Error ? err.message : "切换失败");
    } finally {
      setTogglingId(null);
    }
  };

  const filtered = skills.filter((s) => {
    if (statusFilter !== "all" && s.status !== statusFilter) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        s.id.toLowerCase().includes(q) ||
        (s.name || "").toLowerCase().includes(q) ||
        (s.description || "").toLowerCase().includes(q)
      );
    }
    return true;
  });

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
            诊断技能
          </h1>
          <p className="text-sm text-[var(--color-text-tertiary)] mt-1">
            管理诊断技能配置
          </p>
        </div>
        <button
          onClick={() => navigate("/skills/new")}
          className="console-btn console-btn-primary"
        >
          <Plus className="w-4 h-4" />
          新建技能
        </button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
          <input
            className="console-input pl-9"
            placeholder="搜索技能..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="flex gap-1 bg-[var(--color-surface-overlay)] rounded-lg p-1">
          {(["all", "enabled", "disabled"] as const).map((s) => (
            <button
              key={s}
              onClick={() => setStatusFilter(s)}
              className={cn(
                "px-3 py-1.5 rounded-md text-sm font-medium transition-colors",
                statusFilter === s
                  ? "bg-[var(--color-surface-raised)] text-[var(--color-text-primary)] shadow-sm"
                  : "text-[var(--color-text-tertiary)] hover:text-[var(--color-text-secondary)]"
              )}
            >
              {s === "all" ? "全部" : s === "enabled" ? "已启用" : "已禁用"}
            </button>
          ))}
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="console-card-static p-4 border-l-4 border-l-[var(--color-rose)] text-sm text-[var(--color-rose)]">
          {error}
        </div>
      )}

      {/* Loading */}
      {loading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="console-card-static p-5 space-y-3">
              <div className="skeleton h-5 w-2/3" />
              <div className="skeleton h-4 w-full" />
              <div className="skeleton h-4 w-4/5" />
              <div className="flex gap-2 mt-4">
                <div className="skeleton h-6 w-16 rounded-full" />
                <div className="skeleton h-6 w-16 rounded-full" />
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Empty */}
      {!loading && filtered.length === 0 && (
        <div className="console-card-static p-16 flex flex-col items-center gap-3">
          <Zap className="w-10 h-10 text-[var(--color-text-muted)]" />
          <p className="text-[var(--color-text-tertiary)]">
            {skills.length === 0
              ? "暂无技能"
              : "没有匹配的技能"}
          </p>
          {skills.length === 0 && (
            <button
              onClick={() => navigate("/skills/new")}
              className="console-btn console-btn-primary mt-2"
            >
              <Plus className="w-4 h-4" />
              创建第一个技能
            </button>
          )}
        </div>
      )}

      {/* Skill grid */}
      {!loading && filtered.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filtered.map((skill) => (
            <div
              key={skill.id}
              onClick={() => navigate(`/skills/${skill.id}`)}
              className={cn(
                "console-card p-5 cursor-pointer border-l-4 transition-all",
                skill.status === "enabled"
                  ? "border-l-[var(--color-emerald)]"
                  : "border-l-[var(--color-border-default)]"
              )}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0 flex-1">
                  <h3 className="font-semibold text-[var(--color-text-primary)] truncate">
                    {skill.name || skill.id}
                  </h3>
                  <p className="text-xs text-[var(--color-text-tertiary)] mt-0.5 font-mono">
                    {skill.id}
                  </p>
                </div>
                <span
                  className={cn(
                    "badge flex-shrink-0",
                    skill.status === "enabled"
                      ? "badge-success"
                      : "badge-neutral"
                  )}
                >
                  {skill.status === "enabled" ? (
                    <CheckCircle2 className="w-3 h-3" />
                  ) : (
                    <XCircle className="w-3 h-3" />
                  )}
                  {skill.status === "enabled" ? "已启用" : "已禁用"}
                </span>
              </div>

              {skill.description && (
                <p className="text-sm text-[var(--color-text-secondary)] mt-2 line-clamp-2">
                  {skill.description}
                </p>
              )}

              <div className="flex items-center gap-4 mt-4 text-xs text-[var(--color-text-tertiary)]">
                {(skill.allowed_tools?.length ?? 0) > 0 && (
                  <span className="flex items-center gap-1">
                    <Wrench className="w-3 h-3" />
                    {skill.allowed_tools!.length} 个工具
                  </span>
                )}
                {(skill.required_evidence?.length ?? 0) > 0 && (
                  <span className="flex items-center gap-1">
                    <FileText className="w-3 h-3" />
                    {skill.required_evidence!.length} 项证据
                  </span>
                )}
                {skill.version && (
                  <span className="ml-auto font-mono">v{skill.version}</span>
                )}
              </div>

              <div className="mt-3 pt-3 border-t border-[var(--color-border-subtle)] flex justify-end">
                <button
                  onClick={(e) => handleToggle(e, skill)}
                  disabled={togglingId === skill.id}
                  className={cn(
                    "console-btn text-xs py-1 px-3",
                    skill.status === "enabled"
                      ? "console-btn-secondary"
                      : "console-btn-primary"
                  )}
                >
                  {togglingId === skill.id ? (
                    <Loader2 className="w-3 h-3 animate-spin" />
                  ) : skill.status === "enabled" ? (
                    "禁用"
                  ) : (
                    "启用"
                  )}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
