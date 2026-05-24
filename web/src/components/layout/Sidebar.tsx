import { useState, useEffect } from "react";
import { NavLink } from "react-router-dom";
import {
  LayoutDashboard,
  MessageSquare,
  Wrench,
  Activity,
  Zap,
  BarChart3,
  Sun,
  Moon,
  KeyRound,
  Check,
} from "lucide-react";
import { cn } from "../../lib/utils";
import { useTheme } from "../../hooks/useTheme";
import { useConnectionStatus } from "../../hooks/useConnectionStatus";
import { getApiKey, setApiKey } from "../../lib/api";

const navItems = [
  { icon: LayoutDashboard, label: "总览", to: "/" },
  { icon: MessageSquare, label: "会话", to: "/sessions" },
  { icon: Wrench, label: "工具网关", to: "/tools" },
  { icon: Activity, label: "调用追踪", to: "/invocations" },
  { icon: Zap, label: "诊断技能", to: "/skills" },
  { icon: BarChart3, label: "评估指标", to: "/eval" },
];

export function Sidebar() {
  const { theme, toggle } = useTheme();
  const status = useConnectionStatus();
  const [apiKeyInput, setApiKeyInput] = useState(getApiKey());
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    setApiKeyInput(getApiKey());
  }, []);

  const handleSaveKey = () => {
    setApiKey(apiKeyInput.trim());
    setSaved(true);
    setTimeout(() => setSaved(false), 1500);
  };

  return (
    <aside className="console-sidebar flex flex-col w-[260px] h-full flex-shrink-0">
      {/* Brand */}
      <div className="flex items-center gap-3 px-5 h-[60px] border-b border-[var(--color-border-subtle)]">
        <div className="w-8 h-8 rounded-lg bg-[var(--color-accent)] flex items-center justify-center">
          <Zap className="w-4 h-4 text-white" />
        </div>
        <div>
          <div className="text-sm font-bold text-[var(--color-text-primary)] leading-tight">
            VideoOps
          </div>
          <div className="text-[11px] text-[var(--color-text-tertiary)] leading-tight">
            Agent 控制台
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 px-3 py-4 space-y-1 scroll-panel">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === "/"}
            className={({ isActive }) =>
              cn("nav-item", isActive && "active")
            }
          >
            <item.icon className="w-[18px] h-[18px] flex-shrink-0" />
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      {/* Bottom: Status + Theme */}
      <div className="px-4 py-4 border-t border-[var(--color-border-subtle)] space-y-3">
        {/* Connection status */}
        <div className="flex items-center gap-2 text-xs text-[var(--color-text-tertiary)]">
          <div
            className={cn(
              "status-dot",
              status === "online" && "status-dot-online",
              status === "offline" && "status-dot-offline",
              status === "checking" && "status-dot-checking"
            )}
          />
          <span>
            {status === "online"
              ? "后端已连接"
              : status === "offline"
                ? "后端离线"
                : "检测中..."}
          </span>
        </div>

        {/* Theme toggle */}
        <button
          onClick={toggle}
          className="console-btn console-btn-ghost w-full justify-start text-xs"
        >
          {theme === "dark" ? (
            <>
              <Sun className="w-3.5 h-3.5" />
              <span>浅色模式</span>
            </>
          ) : (
            <>
              <Moon className="w-3.5 h-3.5" />
              <span>深色模式</span>
            </>
          )}
        </button>

        {/* API Key */}
        <div className="space-y-1.5">
          <div className="flex items-center gap-1.5 text-xs text-[var(--color-text-tertiary)]">
            <KeyRound className="w-3 h-3" />
            <span>API Key</span>
          </div>
          <div className="flex gap-1">
            <input
              type="password"
              value={apiKeyInput}
              onChange={(e) => setApiKeyInput(e.target.value)}
              placeholder="输入 API Key"
              className="flex-1 text-[11px] px-2 py-1 rounded bg-[var(--color-surface)] border border-[var(--color-border-subtle)] text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)] focus:outline-none focus:border-[var(--color-accent)]"
            />
            <button
              onClick={handleSaveKey}
              className={cn(
                "px-1.5 py-1 rounded text-[11px] transition-colors",
                saved
                  ? "bg-[var(--color-emerald-soft)] text-[var(--color-emerald)]"
                  : "bg-[var(--color-surface)] text-[var(--color-text-tertiary)] hover:text-[var(--color-accent)]"
              )}
            >
              {saved ? <Check className="w-3 h-3" /> : "保存"}
            </button>
          </div>
        </div>
      </div>
    </aside>
  );
}
