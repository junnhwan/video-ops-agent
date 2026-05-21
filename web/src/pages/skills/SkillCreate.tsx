import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, Loader2, Save } from "lucide-react";
import type { SkillInput, Tool } from "../../types";
import { skillApi, gatewayApi } from "../../lib/api";
import { Select } from "../../components/ui/Select";

const EMPTY_FORM: SkillInput = {
  id: "",
  name: "",
  description: "",
  version: "1.0.0",
  status: "enabled",
  scenario: "",
  allowed_tools: [],
  required_evidence: [],
  prompt_template: "",
  output_sections: [],
};

export function SkillCreate() {
  const navigate = useNavigate();
  const [form, setForm] = useState<SkillInput>({ ...EMPTY_FORM });
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [validationErrors, setValidationErrors] = useState<Record<string, string>>({});
  const [availableTools, setAvailableTools] = useState<Tool[]>([]);
  const [toolsInput, setToolsInput] = useState("");
  const [evidenceInput, setEvidenceInput] = useState("");
  const [sectionsInput, setSectionsInput] = useState("");

  useEffect(() => {
    gatewayApi.listTools().then((res) => setAvailableTools(res.tools || [])).catch(() => {});
  }, []);

  const validate = () => {
    const errors: Record<string, string> = {};
    if (!form.id.trim()) errors.id = "技能 ID 为必填项";
    if (!form.name.trim()) errors.name = "名称为必填项";
    if (!form.scenario.trim()) errors.scenario = "场景为必填项";
    if (!/^[a-z][a-z0-9_]*$/.test(form.id))
      errors.id = "ID 必须为小写字母、数字、下划线，且以字母开头";
    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    try {
      setSaving(true);
      setError("");
      const body: SkillInput = {
        ...form,
        allowed_tools: toolsInput
          .split(",")
          .map((s) => s.trim())
          .filter(Boolean),
        required_evidence: evidenceInput
          .split(",")
          .map((s) => s.trim())
          .filter(Boolean),
        output_sections: sectionsInput
          .split(",")
          .map((s) => s.trim())
          .filter(Boolean),
      };
      const res = await skillApi.create(body);
      navigate(`/skills/${res.skill.id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建技能失败");
    } finally {
      setSaving(false);
    }
  };

  const updateField = <K extends keyof SkillInput>(key: K, value: SkillInput[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    if (validationErrors[key]) {
      setValidationErrors((prev) => {
        const next = { ...prev };
        delete next[key];
        return next;
      });
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <button
          onClick={() => navigate("/skills")}
          className="console-btn console-btn-ghost p-2"
        >
          <ArrowLeft className="w-4 h-4" />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
            新建诊断技能
          </h1>
          <p className="text-sm text-[var(--color-text-tertiary)]">
            定义一个新的诊断技能配置
          </p>
        </div>
      </div>

      {/* Form */}
      <form onSubmit={handleSubmit} className="console-card-static p-6 space-y-5">
        {error && (
          <div className="p-3 rounded-lg bg-[var(--color-rose-soft)] text-[var(--color-rose)] text-sm">
            {error}
          </div>
        )}

        {/* ID */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            技能 ID <span className="text-[var(--color-rose)]">*</span>
          </label>
          <input
            className={cn("console-input font-mono", validationErrors.id && "border-[var(--color-rose)]")}
            placeholder="e.g. custom_comment_review"
            value={form.id}
            onChange={(e) => updateField("id", e.target.value)}
          />
          {validationErrors.id && (
            <p className="text-xs text-[var(--color-rose)] mt-1">{validationErrors.id}</p>
          )}
        </div>

        {/* Name */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            名称 <span className="text-[var(--color-rose)]">*</span>
          </label>
          <input
            className={cn("console-input", validationErrors.name && "border-[var(--color-rose)]")}
            placeholder="e.g. 自定义评论复盘"
            value={form.name}
            onChange={(e) => updateField("name", e.target.value)}
          />
          {validationErrors.name && (
            <p className="text-xs text-[var(--color-rose)] mt-1">{validationErrors.name}</p>
          )}
        </div>

        {/* Description */}
        <div>
          <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
            描述
          </label>
          <textarea
            className="console-input resize-none"
            rows={3}
            placeholder="描述此技能的功能..."
            value={form.description}
            onChange={(e) => updateField("description", e.target.value)}
          />
        </div>

        {/* Version + Status + Scenario row */}
        <div className="grid grid-cols-3 gap-4">
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
              状态
            </label>
            <Select
              value={form.status}
              onChange={(v) => updateField("status", v as "enabled" | "disabled")}
              options={[
                { value: "enabled", label: "已启用" },
                { value: "disabled", label: "已禁用" },
              ]}
              className="w-full"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-[var(--color-text-secondary)] mb-1.5">
              场景 <span className="text-[var(--color-rose)]">*</span>
            </label>
            <input
              className={cn("console-input", validationErrors.scenario && "border-[var(--color-rose)]")}
              placeholder="e.g. comment_risk_analysis"
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
            rows={6}
            placeholder="输入此技能的提示词模板..."
            value={form.prompt_template}
            onChange={(e) => updateField("prompt_template", e.target.value)}
          />
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={() => navigate("/skills")}
            className="console-btn console-btn-secondary"
          >
            取消
          </button>
          <button
            type="submit"
            disabled={saving}
            className="console-btn console-btn-primary"
          >
            {saving ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            新建技能
          </button>
        </div>
      </form>
    </div>
  );
}
