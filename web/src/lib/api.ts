import type { AgentSession, AgentMessage, AgentToolCall, PostMessageResult } from "../types";

const API_BASE = import.meta.env.VITE_API_BASE_URL || "/api";

async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export const api = {
  health: () => fetchJSON<{ status: string }>("/health"),

  createSession: (body: { user_id?: string; title: string; scenario?: string; context_policy_json?: string }) =>
    fetchJSON<{ session: AgentSession }>("/agent/sessions", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  listSessions: (params?: { user_id?: string; limit?: number }) => {
    const q = new URLSearchParams();
    if (params?.user_id) q.set("user_id", params.user_id);
    if (params?.limit) q.set("limit", String(params.limit));
    return fetchJSON<{ sessions: AgentSession[] }>(`/agent/sessions?${q.toString()}`);
  },

  getSession: (id: number) =>
    fetchJSON<{ session: AgentSession; messages: AgentMessage[] }>(`/agent/sessions/${id}`),

  postMessage: (id: number, body: { content: string; required_evidence?: string[] }) =>
    fetchJSON<PostMessageResult>(`/agent/sessions/${id}/messages`, {
      method: "POST",
      body: JSON.stringify(body),
    }),

  listToolCalls: (id: number) =>
    fetchJSON<{ tool_calls: AgentToolCall[] }>(`/agent/sessions/${id}/tool-calls`),
};
