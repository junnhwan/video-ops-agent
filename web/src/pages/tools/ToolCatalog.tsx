import { useState, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Wrench, Search } from "lucide-react";
import type { Tool } from "../../types";
import { gatewayApi } from "../../lib/api";
import { cn } from "../../lib/utils";

export function ToolCatalog() {
  const navigate = useNavigate();
  const [tools, setTools] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [search, setSearch] = useState("");
  const [activeCategory, setActiveCategory] = useState<string | null>(null);

  useEffect(() => {
    gatewayApi
      .listTools()
      .then((res) => setTools(res.tools))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  const categories = useMemo(() => {
    const cats = Array.from(new Set(tools.map((t) => t.category)));
    cats.sort();
    return cats;
  }, [tools]);

  const filteredTools = useMemo(() => {
    let result = tools;
    if (activeCategory) {
      result = result.filter((t) => t.category === activeCategory);
    }
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        (t) =>
          t.name.toLowerCase().includes(q) ||
          t.display_name.toLowerCase().includes(q) ||
          t.description.toLowerCase().includes(q)
      );
    }
    return result;
  }, [tools, activeCategory, search]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-[var(--color-text-primary)]">
          工具网关
        </h1>
        <p className="text-sm text-[var(--color-text-tertiary)] mt-1">
          浏览和调用可用工具
        </p>
      </div>

      {/* Search */}
      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[var(--color-text-muted)]" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="搜索工具名称或描述..."
          className="console-input pl-9"
        />
      </div>

      {/* Category pills */}
      {!loading && categories.length > 0 && (
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => setActiveCategory(null)}
            className={cn(
              "badge cursor-pointer transition-colors",
              activeCategory === null
                ? "badge-accent"
                : "badge-neutral hover:bg-[var(--color-surface-overlay)]"
            )}
          >
            全部
          </button>
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() =>
                setActiveCategory(activeCategory === cat ? null : cat)
              }
              className={cn(
                "badge cursor-pointer transition-colors",
                activeCategory === cat
                  ? "badge-accent"
                  : "badge-neutral hover:bg-[var(--color-surface-overlay)]"
              )}
            >
              {cat}
            </button>
          ))}
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="badge badge-error text-sm">
          加载工具失败: {error}
        </div>
      )}

      {/* Loading skeleton */}
      {loading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="console-card p-5 space-y-3">
              <div className="skeleton h-5 w-2/3" />
              <div className="skeleton h-4 w-1/4" />
              <div className="skeleton h-4 w-full" />
              <div className="skeleton h-4 w-4/5" />
            </div>
          ))}
        </div>
      )}

      {/* Empty state */}
      {!loading && !error && filteredTools.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-center">
          <Wrench className="w-10 h-10 text-[var(--color-text-muted)] mb-3" />
          <p className="text-sm text-[var(--color-text-tertiary)]">
            {search || activeCategory
              ? "没有匹配的工具"
              : "暂无工具"}
          </p>
        </div>
      )}

      {/* Tool grid */}
      {!loading && filteredTools.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredTools.map((tool) => (
            <div
              key={tool.name}
              onClick={() => navigate(`/tools/${tool.name}`)}
              className="console-card p-5 cursor-pointer relative group"
            >
              {/* Wrench icon in top-right */}
              <Wrench className="absolute top-4 right-4 w-4 h-4 text-[var(--color-text-muted)] opacity-0 group-hover:opacity-100 transition-opacity" />

              {/* Display name */}
              <h3 className="text-base font-semibold text-[var(--color-text-primary)] mb-2 pr-6">
                {tool.display_name}
              </h3>

              {/* Badges */}
              <div className="flex items-center gap-2 mb-3">
                <span className="badge badge-accent">{tool.category}</span>
                <span
                  className={cn(
                    "badge",
                    tool.read_only ? "badge-success" : "badge-warning"
                  )}
                >
                  {tool.read_only ? "只读" : "可写"}
                </span>
              </div>

              {/* Description */}
              <p className="text-sm text-[var(--color-text-secondary)] leading-relaxed line-clamp-3">
                {tool.description}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
