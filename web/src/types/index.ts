export interface AgentSession {
  id: number;
  user_id: string;
  title: string;
  scenario: string;
  status: SessionStatus;
  context_policy_json?: string;
  last_message_preview?: string;
  created_at: string;
  updated_at: string;
}

export type SessionStatus = 'active' | 'closed' | 'error';

export interface AgentMessage {
  id: number;
  session_id: number;
  role: MessageRole;
  content: string;
  content_summary?: string;
  created_at: string;
}

export type MessageRole = 'user' | 'assistant' | 'system' | 'tool';

export interface AgentToolCall {
  id: number;
  session_id: number;
  message_id?: number;
  tool_name: string;
  arguments_json: string;
  result_json?: string;
  result_summary?: string;
  latency_ms: number;
  status: ToolCallStatus;
  error_message?: string;
  created_at: string;
}

export type ToolCallStatus = 'success' | 'error' | 'timeout';

export interface CreateSessionInput {
  user_id?: string;
  title: string;
  scenario?: string;
  context_policy_json?: string;
}

export interface PostMessageInput {
  content: string;
  required_evidence?: string[];
}

export interface PostMessageResult {
  session_id: number;
  final_answer: string;
  round_count: number;
  tool_call_count: number;
}

export interface ScenarioTemplate {
  id: string;
  label: string;
  description: string;
  icon: string;
  quickPrompts: string[];
}
