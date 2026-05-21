import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, MessageSquare, Search } from "lucide-react";
import type { AgentSession, Skill } from "../../types";
import { sessionApi, skillApi } from "../../lib/api";
import { cn, formatDate } from "../../lib/utils";
import { Select } from "../../components/ui/Select";

const SCENARIOS = [
  { value: "hot_rank_attribution", label: "热榜归因" },
  { value: "comment_risk_analysis", label: "评论风险分析" },
  { value: "author_support_evaluation", label: "作者扶持评估" },
  { value: "tag_trend_analysis", label: "标签趋势分析" },
  { value: "content_review_summary", label: "内容复盘摘要" },
] as const;

const statusBadge: Record<string, string> = {
  active: "badge badge-success",
  closed: "badge badge-neutral",
  error: "badge badge-error",
};

function SkeletonRow() {
  return (
    <tr>
      <td><div className="skeleton h-4 w-24" /></td>
      <td><div className="skeleton h-4 w-40" /></td>
      <td><div className="skeleton h-4 w-36" /></td>
      <td><div className="skeleton h-4 w-28" /></td>
      <td><div className="skeleton h-4 w-16" /></td>
      <td><div className="skeleton h-4 w-28" /></td>
    </tr>
  );
}

export function SessionList() {
  const navigate = useNavigate();
  const [sessions, setSessions] = useState<AgentSession[]>([]);
  const [skills, setSkills] = useState<Skill[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [showModal, setShowModal] = useState(false);
  const [submitting, setSubmitting] = useState(false);

  // New session form
  const [title, setTitle] = useState("");
  const [scenario, setScenario] = useState("");
  const [skillId, setSkillId] = useState("");
  const [skillVersion, setSkillVersion] = useState("");

  const loadSessions = useCallback(async () => {
    try {
      const data = await sessionApi.list({ limit: 50 });
      setSessions(data.sessions ?? []);
    } catch {
      /* silently fail */
    } finally {
      setLoading(false);
    }
  }, []);

  const loadSkills = useCallback(async () => {
    try {
      const data = await skillApi.list();
      setSkills(data.skills ?? []);
    } catch {
      /* silently fail */
    }
  }, []);

  useEffect(() => {
    loadSessions();
    loadSkills();
  }, [loadSessions, loadSkills]);

  const filtered = sessions.filter((s) => {
    if (!search) return true;
    const q = search.toLowerCase();
    return (
      s.title?.toLowerCase().includes(q) ||
      s.scenario?.toLowerCase().includes(q) ||
      s.skill_id?.toLowerCase().includes(q) ||
      String(s.id).includes(q)
    );
  });

  const handleCreate = async () => {
    if (submitting) return;
    setSubmitting(true);
    try {
      const input: Parameters<typeof sessionApi.create>[0] = {};
      if (title.trim()) input.title = title.trim();
      if (scenario) input.scenario = scenario;
      if (skillId.trim()) input.skill_id = skillId.trim();
      if (skillVersion.trim()) input.skill_version = skillVersion.trim();

      const { session } = await sessionApi.create(input);
      setShowModal(false);
      resetForm();
      navigate(`/sessions/${session.id}`);
    } catch (err) {
      console.error("Failed to create session:", err);
    } finally {
      setSubmitting(false);
    }
  };

  const resetForm = () => {
    setTitle("");
    setScenario("");
    setSkillId("");
    setSkillVersion("");
  };

  const closeModal = () => {
    setShowModal(false);
    resetForm();
  };

  return (
    <div className="space-y-5">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-[var(--color-text-primary)]">
            会话列表
          </h1>
          <p className="text-sm text-[var(--color-text-tertiary)] mt-0.5">
            Agent 对话会话
          </p>
        </div>
        <button
          onClick={() => setShowModal(true)}
          className="console-btn console-btn-primary btn-press"
        >
          <Plus className="w-4 h-4" />
          新建会话
        </button>
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-tertiary)]" />
        <input
          type="text"
          placeholder="搜索会话..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="console-input pl-9"
        />
      </div>

      {/* Table */}
      <div className="console-card overflow-hidden">
        {loading ? (
          <table className="console-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>标题</th>
                <th>场景</th>
                <th>技能</th>
                <th>状态</th>
                <th>更新时间</th>
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: 6 }).map((_, i) => (
                <SkeletonRow key={i} />
              ))}
            </tbody>
          </table>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-[var(--color-text-tertiary)]">
            <MessageSquare className="w-10 h-10 mb-3 opacity-40" />
            <p className="text-sm font-medium">暂无会话</p>
            <p className="text-xs mt-1">
              {search
                ? "请尝试其他搜索关键词"
                : '点击「新建会话」开始使用'}
            </p>
          </div>
        ) : (
          <table className="console-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>标题</th>
                <th>场景</th>
                <th>技能</th>
                <th>状态</th>
                <th>更新时间</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((session) => (
                <tr
                  key={session.id}
                  onClick={() => navigate(`/sessions/${session.id}`)}
                  className="cursor-pointer"
                >
                  <td className="font-mono text-xs text-[var(--color-text-tertiary)]">
                    #{session.id}
                  </td>
                  <td className="font-medium text-[var(--color-text-primary)]">
                    {session.title || (
                      <span className="text-[var(--color-text-muted)] italic">
                        无标题
                      </span>
                    )}
                  </td>
                  <td>
                    {session.scenario ? (
                      <span className="text-xs font-medium">
                        {SCENARIOS.find(s => s.value === session.scenario)?.label ?? session.scenario.replace(/_/g, " ")}
                      </span>
                    ) : (
                      <span className="text-[var(--color-text-muted)]">--</span>
                    )}
                  </td>
                  <td>
                    {session.skill_id ? (
                      <span className="text-xs font-mono">
                        {session.skill_id}
                      </span>
                    ) : (
                      <span className="text-[var(--color-text-muted)]">--</span>
                    )}
                  </td>
                  <td>
                    <span className={statusBadge[session.status] ?? "badge badge-neutral"}>
                      {session.status === "active" ? "活跃" : session.status === "closed" ? "已关闭" : session.status === "error" ? "错误" : session.status}
                    </span>
                  </td>
                  <td className="text-xs text-[var(--color-text-tertiary)] whitespace-nowrap">
                    {formatDate(session.updated_at)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* New Session Modal */}
      {showModal && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
          onClick={closeModal}
        >
          <div
            className="console-card-static w-full max-w-md mx-4 p-0 animate-[fade-in_0.2s_ease] overflow-hidden"
            onClick={(e) => e.stopPropagation()}
          >
            {/* Modal Header */}
            <div className="px-6 py-4 border-b border-[var(--color-border-subtle)]">
              <h2 className="text-base font-semibold text-[var(--color-text-primary)]">
                新建会话
              </h2>
              <p className="text-xs text-[var(--color-text-tertiary)] mt-0.5">
                开始一个新的 Agent 对话
              </p>
            </div>

            {/* Modal Body */}
            <div className="px-6 py-5 space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1.5">
                  标题
                </label>
                <input
                  type="text"
                  placeholder="例如：游戏趋势分析"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="console-input"
                  autoFocus
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1.5">
                  场景
                </label>
                <Select
                  value={scenario}
                  onChange={(v) => setScenario(v)}
                  options={[
                    { value: "", label: "选择场景..." },
                    ...SCENARIOS.map((s) => ({ value: s.value, label: s.label })),
                  ]}
                  className="w-full"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1.5">
                  技能 ID{" "}
                  <span className="text-[var(--color-text-muted)] font-normal">
                    （可选）
                  </span>
                </label>
                <input
                  type="text"
                  placeholder="例如：hot_rank_skill"
                  value={skillId}
                  onChange={(e) => setSkillId(e.target.value)}
                  className="console-input"
                  list="skill-suggestions"
                />
                {skills.length > 0 && (
                  <datalist id="skill-suggestions">
                    {skills.map((s) => (
                      <option key={s.id} value={s.id} />
                    ))}
                  </datalist>
                )}
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-secondary)] mb-1.5">
                  技能版本{" "}
                  <span className="text-[var(--color-text-muted)] font-normal">
                    （可选）
                  </span>
                </label>
                <input
                  type="text"
                  placeholder="例如：1.0.0"
                  value={skillVersion}
                  onChange={(e) => setSkillVersion(e.target.value)}
                  className="console-input"
                />
              </div>
            </div>

            {/* Modal Footer */}
            <div className="px-6 py-4 border-t border-[var(--color-border-subtle)] flex items-center justify-end gap-3">
              <button
                onClick={closeModal}
                className="console-btn console-btn-secondary"
              >
                取消
              </button>
              <button
                onClick={handleCreate}
                disabled={submitting}
                className={cn(
                  "console-btn console-btn-primary btn-press",
                  submitting && "opacity-60 pointer-events-none"
                )}
              >
                {submitting ? "创建中..." : "创建会话"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
