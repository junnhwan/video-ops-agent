import { useState, useEffect } from "react";
import { useSession } from "./SessionProvider";
import { ChatArea } from "./ChatArea";
import { ToolTracePanel } from "./ToolTracePanel";
import { SessionSidebar } from "./SessionSidebar";
import { NewSessionModal } from "./NewSessionModal";
import { useTheme } from "../hooks/useTheme";
import { useConnectionStatus } from "../hooks/useConnectionStatus";
import { Plus, PanelLeftClose, PanelLeft, Terminal, Sun, Moon, ArrowRight, TrendingUp, ShieldAlert, UserCircle, Tag } from "lucide-react";
import { cn } from "../lib/utils";

const quickStartScenarios = [
  { id: "hot_rank_analysis", icon: TrendingUp, label: "热榜归因", desc: "分析视频上热榜的原因与风险", color: "text-amber", bg: "bg-amber-soft", border: "border-amber/20", prompt: "分析一下视频 123 为什么上热榜，有没有运营风险？" },
  { id: "comment_risk_analysis", icon: ShieldAlert, label: "评论风险", desc: "识别评论区争议与攻击风险", color: "text-rose", bg: "bg-rose-soft", border: "border-rose/20", prompt: "分析视频 123 的评论区有没有争议或攻击风险？" },
  { id: "author_profile_analysis", icon: UserCircle, label: "作者画像", desc: "评估作者表现与扶持价值", color: "text-cyan", bg: "bg-cyan-soft", border: "border-cyan/20", prompt: "作者 8 最近表现怎么样，值得扶持吗？" },
  { id: "tag_trend_analysis", icon: Tag, label: "标签趋势", desc: "洞察标签下的内容风向趋势", color: "text-emerald", bg: "bg-emerald-soft", border: "border-emerald/20", prompt: "#Go 后端 这个标签最近内容表现怎么样？" },
];

export function Layout() {
  const { currentSession, error, createSession } = useSession();
  const { theme, toggle: toggleTheme } = useTheme();
  const connStatus = useConnectionStatus();
  const [showSidebar, setShowSidebar] = useState(true);
  const [showToolPanel, setShowToolPanel] = useState(true);
  const [showNewModal, setShowNewModal] = useState(false);

  async function handleQuickStart(scenario: typeof quickStartScenarios[number]) {
    await createSession({ title: scenario.prompt, scenario: scenario.id });
  }

  // Keyboard shortcuts
  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      // Ctrl+N / Cmd+N -> new session
      if ((e.ctrlKey || e.metaKey) && e.key === "n") {
        e.preventDefault();
        setShowNewModal(true);
      }
      // Esc -> close modal or panel
      if (e.key === "Escape") {
        setShowNewModal(false);
      }
    }
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  return (
    <div className="flex h-screen w-screen bg-surface text-text-primary overflow-hidden">
      {/* Mobile sidebar backdrop */}
      {showSidebar && (
        <div
          className="fixed inset-0 z-20 bg-text-primary/25 backdrop-blur-sm md:hidden animate-fade-in"
          onClick={() => setShowSidebar(false)}
        />
      )}

      {/* Sidebar: overlay on mobile, inline on md+ */}
      <div
        className={cn(
          "shrink-0 overflow-hidden panel-transition",
          "fixed inset-y-0 left-0 z-30 md:relative md:z-auto"
        )}
        style={{ width: showSidebar ? 256 : 0, opacity: showSidebar ? 1 : 0 }}
      >
        <SessionSidebar onNewSession={() => { setShowNewModal(true); setShowSidebar(false); }} />
      </div>

      <div className="flex-1 flex flex-col min-w-0 min-w-[320px]">
        <header className="header-bar h-12 shrink-0 flex items-center justify-between px-4 bg-surface-raised/80 backdrop-blur-sm border-b border-border-subtle z-10">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setShowSidebar(!showSidebar)}
              className="btn-press p-1.5 rounded-md hover:bg-surface-overlay text-text-tertiary focus-ring"
              aria-label={showSidebar ? "关闭侧边栏" : "打开侧边栏"}
            >
              {showSidebar ? <PanelLeftClose size={16} /> : <PanelLeft size={16} />}
            </button>
            <div className="flex items-center gap-2.5">
              <div className="w-6 h-6 rounded-md bg-accent flex items-center justify-center shadow-sm">
                <span className="text-[11px] font-bold text-white tracking-wider">VO</span>
              </div>
              <div className="flex flex-col">
                <span className="text-sm font-semibold text-text-primary leading-none">VideoOps Agent</span>
                <span className="text-[11px] text-text-tertiary leading-none mt-0.5 hidden sm:block">短视频内容运营诊断</span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {/* Connection status */}
            <div className="flex items-center gap-1.5 px-2 py-1 rounded-md" title={connStatus === "online" ? "后端已连接" : connStatus === "offline" ? "后端未连接" : "检测中..."}>
              <div className={cn(
                "w-[6px] h-[6px] rounded-full transition-colors duration-500",
                connStatus === "online" && "bg-emerald animate-pulse-soft",
                connStatus === "offline" && "bg-rose",
                connStatus === "checking" && "bg-amber"
              )} />
              <span className="text-[11px] text-text-tertiary hidden sm:inline">
                {connStatus === "online" ? "已连接" : connStatus === "offline" ? "未连接" : "检测中"}
              </span>
            </div>

            {error && (
              <span className="text-xs text-rose bg-rose-soft px-2.5 py-1 rounded-md border border-rose/15 animate-fade-in">
                {error}
              </span>
            )}
            {currentSession && (
              <button
                onClick={() => setShowToolPanel(!showToolPanel)}
                className={`btn-press flex items-center gap-1.5 px-2.5 py-1.5 rounded-md text-xs font-medium transition-colors focus-ring ${
                  showToolPanel
                    ? "bg-accent-soft text-accent border border-accent-border"
                    : "hover:bg-surface-overlay text-text-tertiary border border-transparent"
                }`}
              >
                <Terminal size={13} />
                <span className="hidden sm:inline">追踪</span>
              </button>
            )}
            {/* Theme toggle */}
            <button
              onClick={toggleTheme}
              className="btn-press p-1.5 rounded-md hover:bg-surface-overlay text-text-tertiary focus-ring transition-colors"
              title={theme === "light" ? "切换暗色模式" : "切换亮色模式"}
              aria-label={theme === "light" ? "切换暗色模式" : "切换亮色模式"}
            >
              {theme === "light" ? <Moon size={15} /> : <Sun size={15} />}
            </button>
            <button
              onClick={() => setShowNewModal(true)}
              className="btn-press flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-white bg-accent hover:bg-accent-hover rounded-md shadow-sm focus-ring transition-colors"
            >
              <Plus size={13} />
              <span className="hidden sm:inline">新建诊断</span>
            </button>
          </div>
        </header>

        <div className="flex-1 flex min-h-0">
          <main className="flex-1 min-w-0 flex flex-col">
            {currentSession ? <ChatArea /> : <EmptyWorkspace onQuickStart={handleQuickStart} />}
          </main>

          {/* Mobile tool panel backdrop */}
          {showToolPanel && currentSession && (
            <div
              className="fixed inset-0 z-20 bg-text-primary/25 backdrop-blur-sm lg:hidden animate-fade-in"
              onClick={() => setShowToolPanel(false)}
            />
          )}

          {/* Tool panel: overlay on < lg, inline on lg+ */}
          <div
            className={cn(
              "shrink-0 overflow-hidden panel-transition",
              "fixed inset-y-0 right-0 z-30 lg:relative lg:z-auto"
            )}
            style={{ width: showToolPanel && currentSession ? 360 : 0, opacity: showToolPanel && currentSession ? 1 : 0 }}
          >
            <ToolTracePanel onClose={() => setShowToolPanel(false)} />
          </div>
        </div>
      </div>

      {showNewModal && <NewSessionModal onClose={() => setShowNewModal(false)} />}
    </div>
  );
}

function EmptyWorkspace({ onQuickStart }: { onQuickStart: (s: typeof quickStartScenarios[number]) => Promise<void> }) {
  return (
    <div className="flex-1 flex flex-col items-center justify-center p-8 workspace-bg animate-fade-in">
      <div className="w-14 h-14 rounded-2xl bg-accent-soft border border-accent-border flex items-center justify-center mb-4 shadow-sm">
        <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="text-accent">
          <path d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          <path d="M9 12l2 2 4-4" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      </div>
      <h2 className="text-base font-semibold text-text-primary mb-1">短视频运营诊断工作台</h2>
      <p className="text-sm text-text-tertiary max-w-sm text-center leading-relaxed mb-8">
        选择一个分析场景快速开始，或自定义问题
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-lg w-full mb-6">
        {quickStartScenarios.map((s, i) => (
          <button
            key={s.id}
            onClick={() => onQuickStart(s)}
            className={cn(
              "btn-press text-left p-4 rounded-xl bg-surface-raised border transition-all duration-300 group hover:shadow-lg hover:-translate-y-0.5 focus-ring animate-slide-up",
              s.border
            )}
            style={{ animationDelay: `${i * 80}ms`, animationFillMode: "both" }}
          >
            <div className={cn("w-9 h-9 rounded-lg flex items-center justify-center mb-3 transition-all duration-300 group-hover:scale-110", s.bg)}>
              <s.icon size={18} className={s.color} />
            </div>
            <div className="text-sm font-medium text-text-primary mb-0.5">{s.label}</div>
            <div className="text-[11px] text-text-tertiary leading-relaxed mb-3">{s.desc}</div>
            <div className="flex items-center gap-1 text-[11px] text-accent font-medium opacity-0 translate-x-[-4px] group-hover:opacity-100 group-hover:translate-x-0 transition-all duration-300">
              <span>开始分析</span>
              <ArrowRight size={11} />
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
