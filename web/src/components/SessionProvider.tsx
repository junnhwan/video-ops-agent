import { createContext, useContext, useState, useCallback, type ReactNode } from "react";
import type { AgentSession, AgentMessage, AgentToolCall, CreateSessionInput, PostMessageInput } from "../types";
import { api } from "../lib/api";
import { mockSessions, mockMessages, mockToolCalls } from "../mock/data";

const USE_MOCK = import.meta.env.VITE_USE_MOCK === "true";

interface SessionContextType {
  sessions: AgentSession[];
  currentSession: AgentSession | null;
  messages: AgentMessage[];
  toolCalls: AgentToolCall[];
  loading: boolean;
  sending: boolean;
  error: string | null;
  selectSession: (id: number) => Promise<void>;
  createSession: (input: CreateSessionInput) => Promise<void>;
  sendMessage: (sessionId: number, input: PostMessageInput) => Promise<void>;
  refreshSessions: () => Promise<void>;
  refreshMessages: (sessionId: number) => Promise<void>;
  refreshToolCalls: (sessionId: number) => Promise<void>;
}

const SessionContext = createContext<SessionContextType | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const [sessions, setSessions] = useState<AgentSession[]>(USE_MOCK ? mockSessions : []);
  const [currentSession, setCurrentSession] = useState<AgentSession | null>(null);
  const [messages, setMessages] = useState<AgentMessage[]>([]);
  const [toolCalls, setToolCalls] = useState<AgentToolCall[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refreshSessions = useCallback(async () => {
    if (USE_MOCK) return;
    try {
      setLoading(true);
      const res = await api.listSessions();
      setSessions(res.sessions);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载会话失败");
    } finally {
      setLoading(false);
    }
  }, []);

  const selectSession = useCallback(async (id: number) => {
    if (USE_MOCK) {
      const session = mockSessions.find((s) => s.id === id) || null;
      setCurrentSession(session);
      setMessages(session ? mockMessages.filter((m) => m.session_id === id) : []);
      setToolCalls(session ? mockToolCalls.filter((t) => t.session_id === id) : []);
      return;
    }
    try {
      setLoading(true);
      const res = await api.getSession(id);
      setCurrentSession(res.session);
      setMessages(res.messages);
      // Also fetch tool calls
      const tcRes = await api.listToolCalls(id);
      setToolCalls(tcRes.tool_calls);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "加载会话详情失败");
    } finally {
      setLoading(false);
    }
  }, []);

  const createSession = useCallback(async (input: CreateSessionInput) => {
    if (USE_MOCK) {
      const newSession: AgentSession = {
        id: Date.now(),
        user_id: input.user_id || "ops-001",
        title: input.title,
        scenario: input.scenario || "",
        status: "active",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setSessions((prev) => [newSession, ...prev]);
      setCurrentSession(newSession);
      setMessages([]);
      setToolCalls([]);
      return;
    }
    try {
      const res = await api.createSession(input);
      setSessions((prev) => [res.session, ...prev]);
      setCurrentSession(res.session);
      setMessages([]);
      setToolCalls([]);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建会话失败");
    }
  }, []);

  const sendMessage = useCallback(async (sessionId: number, input: PostMessageInput) => {
    const userMsg: AgentMessage = {
      id: Date.now(),
      session_id: sessionId,
      role: "user",
      content: input.content,
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, userMsg]);

    if (USE_MOCK) {
      setSending(true);
      await new Promise((r) => setTimeout(r, 1500));
      const assistantMsg: AgentMessage = {
        id: Date.now() + 1,
        session_id: sessionId,
        role: "assistant",
        content: `收到您的提问："${input.content}"。\n\n这是模拟的诊断回复。在实际运行中，Agent 会调用平台工具分析数据并生成基于证据的报告。`,
        content_summary: "模拟诊断回复",
        created_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, assistantMsg]);
      setSending(false);
      return;
    }

    try {
      setSending(true);
      const res = await api.postMessage(sessionId, input);
      const assistantMsg: AgentMessage = {
        id: Date.now(),
        session_id: sessionId,
        role: "assistant",
        content: res.final_answer,
        created_at: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, assistantMsg]);
      // Refresh tool calls after message
      const tcRes = await api.listToolCalls(sessionId);
      setToolCalls(tcRes.tool_calls);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "发送消息失败");
    } finally {
      setSending(false);
    }
  }, []);

  const refreshMessages = useCallback(async (sessionId: number) => {
    if (USE_MOCK) return;
    try {
      const res = await api.getSession(sessionId);
      setMessages(res.messages);
    } catch (err) {
      setError(err instanceof Error ? err.message : "刷新消息失败");
    }
  }, []);

  const refreshToolCalls = useCallback(async (sessionId: number) => {
    if (USE_MOCK) return;
    try {
      const res = await api.listToolCalls(sessionId);
      setToolCalls(res.tool_calls);
    } catch (err) {
      setError(err instanceof Error ? err.message : "刷新工具调用失败");
    }
  }, []);

  return (
    <SessionContext.Provider
      value={{
        sessions,
        currentSession,
        messages,
        toolCalls,
        loading,
        sending,
        error,
        selectSession,
        createSession,
        sendMessage,
        refreshSessions,
        refreshMessages,
        refreshToolCalls,
      }}
    >
      {children}
    </SessionContext.Provider>
  );
}

export function useSession() {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error("useSession must be used within SessionProvider");
  return ctx;
}
