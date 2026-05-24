import { useState } from "react";
import { cn, formatDate } from "../lib/utils";
import { useSession } from "./SessionProvider";
import type { AgentSession } from "../types";
import {
  TrendingUp,
  ShieldAlert,
  UserCircle,
  Tag,
  Activity,
  Search,
  Plus,
  Filter,
} from "lucide-react";

const scenarioConfig: Record<string, { icon: React.ReactNode; bg: string; text: string; label: string; shortLabel: string }> = {
  hot_rank_analysis:       { icon: <TrendingUp size={14} />, bg: "bg-amber-soft",  text: "text-amber",  label: "热榜归因", shortLabel: "热榜" },
  comment_risk_analysis:   { icon: <ShieldAlert size={14} />, bg: "bg-rose-soft",   text: "text-rose",   label: "评论风险", shortLabel: "评论" },
  author_profile_analysis: { icon: <UserCircle size={14} />, bg: "bg-cyan-soft",    text: "text-cyan",   label: "作者画像", shortLabel: "作者" },
  tag_trend_analysis:      { icon: <Tag size={14} />,        bg: "bg-emerald-soft", text: "text-emerald",label: "标签趋势", shortLabel: "标签" },
};

const filterTabs = [
  { key: "", label: "全部" },
  { key: "hot_rank_analysis", label: "热榜" },
  { key: "comment_risk_analysis", label: "评论" },
  { key: "author_profile_analysis", label: "作者" },
  { key: "tag_trend_analysis", label: "标签" },
];

const statusBadge: Record<string, { text: string; className: string }> = {
  active:  { text: "进行中", className: "bg-accent-soft text-accent border-accent-border" },
  error:   { text: "异常",   className: "bg-rose-soft text-rose border-rose/20" },
  closed:  { text: "已结束", className: "bg-surface-overlay text-text-tertiary border-border-subtle" },
};

interface SessionSidebarProps {
  onNewSession: () => void;
}

export function SessionSidebar({ onNewSession }: SessionSidebarProps) {
  const { sessions, currentSession, selectSession } = useSession();
  const [search, setSearch] = useState("");
  const [scenarioFilter, setScenarioFilter] = useState("");

  const filtered = sessions.filter((s) => {
    if (search && !(s.title ?? "").toLowerCase().includes(search.toLowerCase())) return false;
    if (scenarioFilter && s.scenario !== scenarioFilter) return false;
    return true;
  });

  return (
    <aside className="w-64 flex flex-col bg-sidebar border-r border-border-subtle h-full shadow-lg md:shadow-none">
      {/* Header */}
      <div className="h-12 shrink-0 flex items-center justify-between px-4 border-b border-border-subtle">
        <span className="text-xs font-semibold text-text-tertiary uppercase tracking-widest">会话</span>
        <button
          onClick={onNewSession}
          className="btn-press p-1 rounded hover:bg-surface-overlay text-text-tertiary hover:text-accent transition-colors focus-ring"
          aria-label="新建会话"
        >
          <Plus size={15} />
        </button>
      </div>

      {/* Search */}
      <div className="px-3 py-2">
        <div className="relative">
          <Search size={13} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            type="text"
            placeholder="搜索..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            aria-label="搜索会话"
            className="w-full text-xs pl-8 pr-3 py-1.5 rounded-md bg-surface-overlay border border-border-subtle text-text-primary placeholder-text-muted focus:outline-none focus:border-accent/40 focus:ring-2 focus:ring-accent/10 transition-all"
          />
        </div>
      </div>

      {/* Scenario filter tabs */}
      <div className="px-3 pb-2">
        <div className="flex items-center gap-0.5 overflow-x-auto">
          <Filter size={11} className="text-text-muted shrink-0 mr-1" />
          {filterTabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setScenarioFilter(tab.key)}
              className={cn(
                "btn-press text-[11px] font-medium px-2 py-1 rounded-md transition-colors whitespace-nowrap focus-ring",
                scenarioFilter === tab.key
                  ? "bg-accent-soft text-accent"
                  : "text-text-tertiary hover:bg-surface-overlay hover:text-text-secondary"
              )}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* List */}
      <div className="flex-1 overflow-y-auto min-h-0" role="listbox" aria-label="会话列表">
        {filtered.length === 0 ? (
          <div className="flex flex-col items-center py-10 text-text-muted animate-fade-in">
            <Activity size={20} className="mb-2" />
            <p className="text-xs">暂无会话</p>
            <p className="text-[11px] text-text-muted mt-1">尝试更换筛选条件</p>
          </div>
        ) : (
          filtered.map((s, i) => (
            <SessionItem key={s.id} session={s} isActive={currentSession?.id === s.id} onClick={() => selectSession(s.id)} index={i} />
          ))
        )}
      </div>
    </aside>
  );
}

function SessionItem({ session, isActive, onClick, index }: { session: AgentSession; isActive: boolean; onClick: () => void; index: number }) {
  const config = scenarioConfig[session.scenario ?? ""];
  const badge = statusBadge[session.status];

  return (
    <button
      onClick={onClick}
      role="option"
      aria-selected={isActive}
      className={cn(
        "session-item relative w-[calc(100%-16px)] text-left px-3 py-2.5 mx-2 rounded-lg flex items-center gap-2.5 mb-0.5 group overflow-hidden",
        isActive
          ? "bg-accent/5"
          : "hover:bg-surface-overlay"
      )}
      style={{ animationDelay: `${index * 40}ms` }}
    >
      {/* Active indicator bar */}
      <div className={cn(
        "absolute left-0 top-1 bottom-1 w-[3px] rounded-r-full bg-accent transition-all duration-300",
        isActive ? "opacity-100 scale-y-100" : "opacity-0 scale-y-0"
      )} />
      <div className={cn(
        "w-7 h-7 rounded-md shrink-0 flex items-center justify-center transition-colors duration-200",
        isActive ? "bg-accent/10" : (config?.bg || "bg-surface-overlay")
      )}>
        <span className={cn("transition-colors duration-200", isActive ? "text-accent" : (config?.text || "text-text-tertiary"))}>
          {config?.icon || <Activity size={14} />}
        </span>
      </div>
      <div className="flex-1 min-w-0">
        <div className={cn("text-xs font-medium truncate transition-colors duration-200", isActive ? "text-accent" : "text-text-primary")}>
          {session.title}
        </div>
        {session.last_message_preview && (
          <p className="text-[11px] text-text-muted truncate mt-0.5 leading-snug">{session.last_message_preview}</p>
        )}
        <div className="flex items-center gap-1.5 mt-0.5">
          <span className="text-[11px] text-text-tertiary">{formatDate(session.updated_at)}</span>
          {config && (
            <span className="text-[11px] text-text-muted">· {config.shortLabel}</span>
          )}
          {badge && session.status !== "active" && (
            <>
              <span className="text-[11px] text-text-muted">·</span>
              <span className={cn("text-[11px] px-1 py-px rounded border", badge.className)}>
                {badge.text}
              </span>
            </>
          )}
        </div>
      </div>
    </button>
  );
}
