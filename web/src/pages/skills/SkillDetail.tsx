import { useState, useEffect, useRef } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  ArrowLeft,
  Save,
  Loader2,
  CheckCircle2,
  XCircle,
  AlertCircle,
} from "lucide-react";
import type { Skill, SkillInput, Tool } from "../../types";
import { skillApi, gatewayApi } from "../../lib/api";
import { cn } from "../../lib/utils";

export function SkillDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [skill, setSkill] = useState<Skill | null>(null);
  const [form, setForm] = useState<SkillInput | null>(null);
  const [availableTools, setAvailableTools] = useState<Tool[]>([]);
  const [toolsInput, setToolsInput] = useState("");
  const [evidenceInput, setEvidenceInput] = useState("");
  const [sectionsInput, setSectionsInput] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [saveMessage, setSaveMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [togglingStatus, setTogglingStatus] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    if (!id) return;
    Promise.all([skillApi.get(id), gatewayApi.listTools()])
      .then(([skillRes, toolsRes]) => {
        const s = skillRes.skill;
        setSkill(s);
        setAvailableTools(toolsRes.tools || []);
        setForm({
          id: s.id,
          name: s.name || "",
          description: s.description || "",
          version: s.version || "1.0.0",
          status: s.status,
          scenario: s.scenario || "",
          allowed_tools: s.allowed_tools || [],
          required_evidence: s.required_evidence || [],
          prompt_template: s.prompt_template || "",
          output_sections: s.output_sections || [],
        });
        setToolsInput((s.allowed_tools || []).join(", "));
        setEvidenceInput((s.required_evidence || []).join(", "));
        setSectionsInput((s.output_sections || []).join(", "));
      })
      .catch((err) =>
        setError(err instanceof Error ? err.message : "加载技能失败")
      )
      .finally(() => setLoading(false));
  }, [id]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id || !form) return;

    try {
      setSaving(true);
      setSaveMessage(null);
      const body: SkillInput = {
        ...form,
        allowed_tools: toolsInput.split(",").map((s) => s.trim()).filter(Boolean),
        required_evidence: evidenceInput.split(",").map((s) => s.trim()).filter(Boolean),
        output_sections: sectionsInput.split(",").map((s) => s.trim()).filter(Boolean),
      };
      const res = await skillApi.update(id, body);
      setSkill(res.skill);
      setSaveMessage({ type: "success", text: "技能更新成功" });
    } catch (err) {
      setSaveMessage({
        type: "error",
        text: err instanceof Error ? err.message : "更新失败",
      });
    } finally {
      setSaving(false);
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setSaveMessage(null), 3000);
    }
  };

  const handleToggleStatus = async () => {
    if (!id || !skill) return;
    setTogglingStatus(true);
    try {
      const res =
        skill.status === "enabled"
          ? await skillApi.disable(id)
          : await skillApi.enable(id);
      setSkill(res.skill);
      if (form) setForm({ ...form, status: res.skill.status });
    } catch (err) {
      setSaveMessage({
        type: "error",
        text: err instanceof Error ? err.message : "切换失败",
      });
    } finally {
      setTogglingStatus(false);
    }
  };

  const updateField = <K extends keyof SkillInput>(key: K, value: SkillInput[K]) => {
    if (!form) return;
    setForm({ ...form, [key]: value });
  };

  if (loading) {
    return (
      <div className="max-w-2xl mx-auto space-y-4">
        <div className="skeleton h-8 w-48" />
        <div className="console-card-static p-6 space-y-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="skeleton h-10 w-full" />
          ))}
        </div>
      </div>
    );
  }

  if (error || !skill || !form) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="console-card-static p-8 flex flex-col items-center gap-3">
          <AlertCircle className="w-8 h-8 text-[var(--color-rose)]" />
          <p className="text-[var(--color-text-secondary)]">{error || "技能未找到"}</p>
          <button onClick={() => navigate("/skills")} className="console-btn console-btn-secondary mt-2">
            返回技能列表
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button onClick={() => navigate("/skills")} className="console-btn console-btn-ghost p-2">
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold text-[var(--color-text-primary)] truncate">
              {skill.name || skill.id}
            </h1>
            <span
              className={cn(
                "badge flex-shrink-0",
                skill.status === "enabled" ? "badge-success" : "badge-neutral"
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
          <p className="text-xs text-[var(--color-text-tertiary)] font-mono mt-0.5">
            {skill.id} {skill.version && `• v${skill.version}`}
          </p>
        </div>
        <button
          onClick={handleToggleStatus}
          disabled={togglingStatus}
          className={cn(
            "console-btn text-sm",
            skill.status === "enabled" ? "console-btn-secondary" : "console-btn-primary"
          )}
        >
          {togglingStatus ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : skill.status === "enabled" ? (
            "禁用"
          ) : (
            "启用"
          )}
        </button>
      </div>

      {/* Save notification */}
      {saveMessage && (
        <div
          className={cn(
            "mb-4 p-3 rounded-lg text-sm flex items-center gap-2 animate-[fade-in_0.2s_ease]",
            saveMessage.type === "success"
              ? "bg-[var(--color-emerald-soft)] text-[var(--color-emerald)]"
              : "bg-[var(--color-rose-soft)] text-[var(--color-rose)]"
          )}
        >
          {saveMessage.type === "success" ? (
            <CheckCircle2 className="w-4 h-4" />
          ) : (
            <AlertCircle className="w-4 h-4" />
          )}
          {saveMessage.text}
        </div>
      )}

      {/* Edit form */}
      <form onSubmit={handleSave} className="console-card-static p-6 space-y-5">
        {/* ID (read-only) */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            技能 ID
          </label>
          <input className="console-input bg-[var(--color-surface-overlay)] opacity-60" value={form.id} readOnly />
        </div>

        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            名称
          </label>
          <input
            className="console-input"
            value={form.name}
            onChange={(e) => updateField("name", e.target.value)}
          />
        </div>

        {/* Description */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            描述
          </label>
          <textarea
            className="console-input resize-none"
            rows={3}
            value={form.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>

        {/* Version + Scenario */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
              版本
            </label>
            <input
              className="console-input font-mono"
              value={form.version}
              onChange={(e) => updateField("version", e.target.value)}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
              场景
            </label>
            <input
              className="console-input"
              value={form.scenario}
              onChange={(e) => updateField("scenario", e.target.value)}
            />
          </div>
        </div>

        {/* Allowed Tools */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            允许的工具
          </label>
          <input
            className="console-input"
            placeholder="逗号分隔的工具名称"
            value={toolsInput}
            onChange={(e) => setToolsInput(e.target.value)}
          />
          {availableTools.length > 0 && (
            <div className="flex flex-wrap gap-1 mt-2">
              {availableTools.map((t) => (
                <button
                  key={t.name}
                  type="button"
                  onClick={() => {
                    const current = toolsInput.split(",").map((s) => s.trim()).filter(Boolean);
                    if (!current.includes(t.name)) {
                      setToolsInput(current.concat(t.name).join(", "));
                    }
                  }}
                  className="text-xs px-2 py-0.5 rounded bg-[var(--color-surface-overlay)] text-[var(--color-text-tertiary)] hover:bg-[var(--color-accent-soft)] hover:text-[var(--color-accent)] transition-colors"
                >
                  + {t.name}
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Required Evidence */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            必要证据
          </label>
          <input
            className="console-input"
            placeholder="逗号分隔的工具名称"
            value={evidenceInput}
            onChange={(e) => setEvidenceInput(e.target.value)}
          />
        </div>

        {/* Output Sections */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            输出段落
          </label>
          <input
            className="console-input"
            placeholder="逗号分隔的段落名称"
            value={sectionsInput}
            onChange={(e) => setSectionsInput(e.target.value)}
          />
        </div>

        {/* Prompt Template */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            提示词模板
          </label>
          <textarea
            className="console-input resize-none font-mono text-sm"
            rows={8}
            value={form.prompt_template}
            onChange={(e) => updateField("prompt_template", e.target.value)}
          />
        </div>

        {/* Save */}
        <div className="flex justify-end pt-2">
          <button type="submit" disabled={saving} className="console-btn console-btn-primary">
            {saving ? <Loader2 className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
            保存修改
          </button>
        </div>
      </form>
    </div>
  );
}
