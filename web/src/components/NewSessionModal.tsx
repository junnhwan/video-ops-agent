import { useState, useRef, useEffect, useCallback } from "react";
import { cn } from "../lib/utils";
import { useSession } from "./SessionProvider";
import { mockScenarios } from "../mock/data";
import {
  X,
  TrendingUp,
  ShieldAlert,
  UserCircle,
  Tag,
  Sparkles,
  ArrowRight,
  Loader2,
} from "lucide-react";

const iconMap: Record<string, React.ReactNode> = {
  TrendingUp: <TrendingUp size={18} />,
  ShieldAlert: <ShieldAlert size={18} />,
  UserCircle: <UserCircle size={18} />,
  Tag: <Tag size={18} />,
};

const scenarioTheme: Record<string, { bg: string; border: string; text: string; softBg: string }> = {
  hot_rank_analysis:       { bg: "bg-amber-soft",  border: "border-amber/25",  text: "text-amber",  softBg: "bg-amber/5" },
  comment_risk_analysis:   { bg: "bg-rose-soft",   border: "border-rose/25",   text: "text-rose",   softBg: "bg-rose/5" },
  author_profile_analysis: { bg: "bg-cyan-soft",   border: "border-cyan/25",   text: "text-cyan",   softBg: "bg-cyan/5" },
  tag_trend_analysis:      { bg: "bg-emerald-soft", border: "border-emerald/25", text: "text-emerald", softBg: "bg-emerald/5" },
};

interface NewSessionModalProps {
  onClose: () => void;
}

export function NewSessionModal({ onClose }: NewSessionModalProps) {
  const { createSession } = useSession();
  const [title, setTitle] = useState("");
  const [selectedScenario, setSelectedScenario] = useState<string>("");
  const [creating, setCreating] = useState(false);
  const modalRef = useRef<HTMLDivElement>(null);

  // Focus trap: keep Tab cycling inside the modal
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    if (e.key === "Escape") { onClose(); return; }
    if (e.key !== "Tab" || !modalRef.current) return;
    const focusable = modalRef.current.querySelectorAll<HTMLElement>(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    if (focusable.length === 0) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (e.shiftKey) {
      if (document.activeElement === first) { e.preventDefault(); last.focus(); }
    } else {
      if (document.activeElement === last) { e.preventDefault(); first.focus(); }
    }
  }, [onClose]);

  useEffect(() => {
    document.addEventListener("keydown", handleKeyDown);
    // Focus the modal on mount
    modalRef.current?.focus();
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);

  async function handleCreate() {
    if (!title.trim()) return;
    setCreating(true);
    await createSession({ title: title.trim(), scenario: selectedScenario || undefined });
    setCreating(false);
    onClose();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center" onClick={onClose} role="presentation">
      {/* Backdrop with blur */}
      <div className="absolute inset-0 bg-text-primary/25 backdrop-blur-sm animate-fade-in" />

      {/* Modal with slide-up entrance */}
      <div
        ref={modalRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
        tabIndex={-1}
        className="relative bg-surface-raised border border-border-subtle rounded-2xl shadow-2xl w-full max-w-lg mx-4 max-h-[80vh] flex flex-col overflow-hidden animate-slide-up focus:outline-none"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="px-6 py-4 flex items-center justify-between border-b border-border-subtle shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-accent-soft border border-accent-border flex items-center justify-center">
              <Sparkles size={14} className="text-accent" />
            </div>
            <div>
              <h2 id="modal-title" className="text-sm font-semibold text-text-primary">新建诊断会话</h2>
              <p className="text-[11px] text-text-tertiary">选择场景，描述问题，开始分析</p>
            </div>
          </div>
          <button onClick={onClose} className="btn-press p-1.5 rounded-lg hover:bg-surface-overlay text-text-tertiary transition-colors focus-ring">
            <X size={16} />
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto px-6 py-5 space-y-5 scroll-smooth">
          {/* Scenario cards */}
          <div>
            <label className="text-[11px] font-semibold text-text-tertiary uppercase tracking-wider mb-3 block">分析场景</label>
            <div className="grid grid-cols-2 gap-2">
              {mockScenarios.map((s, i) => {
                const theme = scenarioTheme[s.id] || { bg: "bg-surface-overlay", border: "border-border-default", text: "text-text-secondary", softBg: "bg-surface-overlay" };
                const isSelected = selectedScenario === s.id;
                return (
                  <button
                    key={s.id}
                    onClick={() => setSelectedScenario(isSelected ? "" : s.id)}
                    className={cn(
                      "btn-press text-left p-3.5 rounded-xl border transition-all duration-200 focus-ring",
                      isSelected
                        ? `${theme.softBg} ${theme.border} shadow-sm`
                        : "bg-surface-overlay border-border-subtle hover:border-border-default hover:shadow-sm"
                    )}
                    style={{ animationDelay: `${i * 50}ms` }}
                  >
                    <div className={cn("w-8 h-8 rounded-lg flex items-center justify-center mb-2 transition-colors duration-200", isSelected ? theme.bg : "bg-surface-overlay")}>
                      <span className={cn("transition-colors duration-200", isSelected ? theme.text : "text-text-tertiary")}>
                        {iconMap[s.icon]}
                      </span>
                    </div>
                    <div className="text-xs font-medium text-text-primary">
                      {s.label}
                    </div>
                    <div className="text-[11px] text-text-tertiary mt-0.5 line-clamp-2 leading-relaxed">
                      {s.description}
                    </div>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Quick prompts - smooth expand */}
          <div className={cn("collapse-wrapper", selectedScenario && "open")}>
            <div className="collapse-inner">
              <label className="text-[11px] font-semibold text-text-tertiary uppercase tracking-wider mb-2 block">快速提问</label>
              <div className="space-y-1">
              {mockScenarios.find((s) => s.id === selectedScenario)?.quickPrompts.map((qp, i) => (
                <button
                  key={i}
                  onClick={() => setTitle(qp)}
                  className={cn(
                    "btn-press w-full text-left text-xs px-3 py-2.5 rounded-lg border transition-all duration-200 group focus-ring",
                    title === qp
                      ? "bg-accent-soft border-accent-border text-accent"
                      : "bg-surface border-border-subtle text-text-secondary hover:border-border-default hover:text-text-primary"
                  )}
                >
                  <div className="flex items-center gap-2">
                    <ArrowRight size={10} className="shrink-0 text-text-muted group-hover:text-accent transition-colors duration-200" />
                    <span>{qp}</span>
                  </div>
                </button>
              ))}
              </div>
            </div>
          </div>

          {/* Custom input */}
          <div>
            <label className="text-[11px] font-semibold text-text-tertiary uppercase tracking-wider mb-2 block">或自定义问题</label>
            <textarea
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="描述你想分析的运营问题..."
              rows={2}
              className="w-full text-sm px-4 py-3 rounded-xl bg-surface border border-border-default text-text-primary placeholder-text-muted focus:outline-none focus:border-accent/50 focus:ring-2 focus:ring-accent/10 resize-none transition-all duration-200"
            />
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-border-subtle flex items-center justify-end gap-3 shrink-0">
          <button onClick={onClose} className="btn-press px-4 py-2 text-xs text-text-tertiary hover:text-text-secondary hover:bg-surface-overlay rounded-lg transition-colors focus-ring">
            取消
          </button>
          <button
            onClick={handleCreate}
            disabled={!title.trim() || creating}
            className={cn(
              "btn-press px-5 py-2 text-xs font-medium rounded-lg transition-all duration-200 flex items-center gap-2 focus-ring",
              title.trim() && !creating
                ? "bg-accent text-white hover:bg-accent-hover shadow-sm"
                : "bg-border-subtle text-text-muted cursor-not-allowed"
            )}
          >
            {creating ? (
              <><Loader2 size={13} className="animate-spin" /> 创建中...</>
            ) : (
              <>开始诊断 <ArrowRight size={13} /></>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
