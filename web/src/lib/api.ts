import type {
  AgentSession,
  AgentMessage,
  AgentToolCall,
  CreateSessionInput,
  PostMessageResult,
  Tool,
  ToolCallInput,
  ToolCallResult,
  Invocation,
  InvocationFilters,
  Skill,
  SkillInput,
  EvalSummary,
  EvalRun,
  EvalRunInput,
} from "../types";
import { mockApi } from "../mock/console-data";

const USE_MOCK = import.meta.env.VITE_USE_MOCK === "true";
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

// ===== Health =====
export const healthApi = {
  check: () =>
    USE_MOCK
      ? mockApi.health()
      : fetchJSON<{ status: string }>("/health"),
};

// ===== Agent Sessions =====
export const sessionApi = {
  create: (body: CreateSessionInput) =>
    USE_MOCK
      ? mockApi.createSession(body as unknown as Record<string, unknown>)
      : fetchJSON<{ session: AgentSession }>("/agent/sessions", {
          method: "POST",
          body: JSON.stringify(body),
        }),

  list: (params?: { user_id?: string; limit?: number }) => {
    if (USE_MOCK) return mockApi.listSessions();
    const q = new URLSearchParams();
    if (params?.user_id) q.set("user_id", params.user_id);
    if (params?.limit) q.set("limit", String(params.limit));
    const qs = q.toString();
    return fetchJSON<{ sessions: AgentSession[] }>(
      `/agent/sessions${qs ? `?${qs}` : ""}`
    );
  },

  get: (id: number) =>
    USE_MOCK
      ? mockApi.getSession(id)
      : fetchJSON<{ session: AgentSession; messages: AgentMessage[] }>(
          `/agent/sessions/${id}`
        ),

  postMessage: (
    id: number,
    body: {
      content: string;
      skill_id?: string;
      required_evidence?: string[];
    }
  ) =>
    USE_MOCK
      ? mockApi.postMessage(id)
      : fetchJSON<PostMessageResult>(`/agent/sessions/${id}/messages`, {
          method: "POST",
          body: JSON.stringify(body),
        }),

  listToolCalls: (id: number) =>
    USE_MOCK
      ? mockApi.listToolCalls(id)
      : fetchJSON<{ tool_calls: AgentToolCall[] }>(
          `/agent/sessions/${id}/tool-calls`
        ),
};

// ===== Tool Gateway =====
export const gatewayApi = {
  listTools: () =>
    USE_MOCK
      ? mockApi.listTools()
      : fetchJSON<{ tools: Tool[] }>("/gateway/tools"),

  getTool: (name: string) =>
    USE_MOCK
      ? mockApi.getTool(name)
      : fetchJSON<{ tool: Tool }>(
          `/gateway/tools/${encodeURIComponent(name)}`
        ),

  callTool: (name: string, body: ToolCallInput) =>
    USE_MOCK
      ? mockApi.callTool(name)
      : fetchJSON<ToolCallResult>(
          `/gateway/tools/${encodeURIComponent(name)}/call`,
          {
            method: "POST",
            body: JSON.stringify(body),
          }
        ),

  listInvocations: (filters?: InvocationFilters) => {
    if (USE_MOCK) return mockApi.listInvocations(filters as Record<string, unknown>);
    const q = new URLSearchParams();
    if (filters?.source) q.set("source", filters.source);
    if (filters?.tool_name) q.set("tool_name", filters.tool_name);
    if (filters?.session_id)
      q.set("session_id", String(filters.session_id));
    if (filters?.skill_id) q.set("skill_id", filters.skill_id);
    if (filters?.status) q.set("status", filters.status);
    if (filters?.limit) q.set("limit", String(filters.limit));
    const qs = q.toString();
    return fetchJSON<{ invocations: Invocation[] }>(
      `/gateway/invocations${qs ? `?${qs}` : ""}`
    );
  },

  getInvocation: (id: number) =>
    USE_MOCK
      ? mockApi.getInvocation(id)
      : fetchJSON<{ invocation: Invocation }>(`/gateway/invocations/${id}`),
};

// ===== Skills =====
export const skillApi = {
  list: () =>
    USE_MOCK
      ? mockApi.listSkills()
      : fetchJSON<{ skills: Skill[] }>("/skills"),

  get: (id: string) =>
    USE_MOCK
      ? mockApi.getSkill(id)
      : fetchJSON<{ skill: Skill }>(
          `/skills/${encodeURIComponent(id)}`
        ),

  create: (body: SkillInput) =>
    USE_MOCK
      ? mockApi.createSkill(body as unknown as Record<string, unknown>)
      : fetchJSON<{ skill: Skill }>("/skills", {
          method: "POST",
          body: JSON.stringify(body),
        }),

  update: (id: string, body: SkillInput) =>
    USE_MOCK
      ? mockApi.updateSkill(id, body as unknown as Record<string, unknown>)
      : fetchJSON<{ skill: Skill }>(
          `/skills/${encodeURIComponent(id)}`,
          {
            method: "PUT",
            body: JSON.stringify(body),
          }
        ),

  enable: (id: string) =>
    USE_MOCK
      ? mockApi.enableSkill(id)
      : fetchJSON<{ skill: Skill }>(
          `/skills/${encodeURIComponent(id)}/enable`,
          { method: "POST" }
        ),

  disable: (id: string) =>
    USE_MOCK
      ? mockApi.disableSkill(id)
      : fetchJSON<{ skill: Skill }>(
          `/skills/${encodeURIComponent(id)}/disable`,
          { method: "POST" }
        ),
};

// ===== Eval =====
export const evalApi = {
  summary: () =>
    USE_MOCK
      ? mockApi.evalSummary()
      : fetchJSON<EvalSummary>("/eval/summary"),

  skillSummary: (skillId: string) =>
    USE_MOCK
      ? mockApi.evalSkillSummary()
      : fetchJSON<EvalSummary>(
          `/eval/skills/${encodeURIComponent(skillId)}/summary`
        ),

  createRun: (body: EvalRunInput) =>
    USE_MOCK
      ? mockApi.createEvalRun(body as unknown as Record<string, unknown>)
      : fetchJSON<{ run: EvalRun }>("/eval/runs", {
          method: "POST",
          body: JSON.stringify(body),
        }),

  getRun: (id: number) =>
    USE_MOCK
      ? mockApi.createEvalRun({ mode: "baseline", skill_id: "mock" })
      : fetchJSON<{ run: EvalRun }>(`/eval/runs/${id}`),
};

// ===== Combined API (backward compat) =====
export const api = {
  // session
  listSessions: sessionApi.list,
  createSession: sessionApi.create,
  getSession: sessionApi.get,
  postMessage: sessionApi.postMessage,
  listToolCalls: sessionApi.listToolCalls,
  // gateway
  listTools: gatewayApi.listTools,
  getTool: gatewayApi.getTool,
  callTool: gatewayApi.callTool,
  listInvocations: gatewayApi.listInvocations,
  getInvocation: gatewayApi.getInvocation,
  // skills
  listSkills: skillApi.list,
  getSkill: skillApi.get,
  createSkill: skillApi.create,
  updateSkill: skillApi.update,
  enableSkill: skillApi.enable,
  disableSkill: skillApi.disable,
  // eval
  evalSummary: evalApi.summary,
  evalSkillSummary: evalApi.skillSummary,
  createEvalRun: evalApi.createRun,
  getEvalRun: evalApi.getRun,
};
