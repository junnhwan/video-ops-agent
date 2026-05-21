// ===== Agent Session =====
export interface AgentSession {
  id: number;
  user_id: string;
  title?: string;
  scenario?: string;
  skill_id?: string;
  skill_version?: string;
  status: SessionStatus;
  context_policy_json?: string;
  last_message_preview?: string;
  created_at: string;
  updated_at: string;
}

export type SessionStatus = "active" | "closed" | "error";

export interface AgentMessage {
  id: number;
  session_id: number;
  role: MessageRole;
  content: string;
  content_summary?: string;
  skill_id?: string;
  required_evidence?: string[];
  created_at: string;
}

export type MessageRole = "user" | "assistant" | "system" | "tool";

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

export type ToolCallStatus = "success" | "error" | "timeout";

// ===== Gateway Tool =====
export interface Tool {
  name: string;
  display_name: string;
  category: string;
  description: string;
  read_only: boolean;
  schema: ToolSchema;
}

export interface ToolSchema {
  type: string;
  function: {
    name: string;
    description?: string;
    parameters: {
      type: string;
      properties?: Record<string, ToolParamDef>;
      required?: string[];
    };
  };
}

export interface ToolParamDef {
  type: string;
  description?: string;
  enum?: string[];
  default?: unknown;
}

export interface ToolCallInput {
  source?: string;
  session_id?: number;
  skill_id?: string;
  arguments: Record<string, unknown>;
}

export interface ToolCallResult {
  invocation: Invocation;
  result: ToolResultData;
}

export interface ToolResultData {
  tool_name: string;
  summary: string;
  data: Record<string, unknown>;
}

// ===== Invocation =====
export interface Invocation {
  id: number;
  source: InvocationSource;
  tool_name: string;
  status: ToolCallStatus;
  latency_ms: number;
  result_summary?: string;
  session_id?: number;
  skill_id?: string;
  arguments_json?: string;
  result_json?: string;
  error_message?: string;
  created_at: string;
}

export type InvocationSource =
  | "manual_console"
  | "agent_runtime"
  | "mcp_client";

export interface InvocationFilters {
  source?: string;
  tool_name?: string;
  session_id?: number;
  skill_id?: string;
  status?: string;
  limit?: number;
}

// ===== Skill =====
export interface Skill {
  id: string;
  name?: string;
  description?: string;
  version?: string;
  status: SkillStatus;
  scenario?: string;
  allowed_tools?: string[];
  required_evidence?: string[];
  prompt_template?: string;
  output_sections?: string[];
  created_at?: string;
  updated_at?: string;
}

export type SkillStatus = "enabled" | "disabled";

export interface SkillInput {
  id: string;
  name: string;
  description: string;
  version: string;
  status: SkillStatus;
  scenario: string;
  allowed_tools: string[];
  required_evidence: string[];
  prompt_template: string;
  output_sections: string[];
}

// ===== Eval =====
export interface EvalSummary {
  tool_call_success_rate: number | null;
  tool_call_error_count: number | null;
  unauthorized_tool_call_count: number | null;
  evidence_complete_final_answer_count: number | null;
  average_tool_latency_ms: number | null;
  average_tool_call_count: number | null;
  skill_success_count: number | null;
  skill_failure_count: number | null;
  unsupported_metrics?: string[];
}

export interface EvalRun {
  id: number;
  mode: EvalMode;
  skill_id: string;
  summary: EvalSummary;
  status?: string;
  created_at?: string;
}

export type EvalMode = "baseline" | "skill_guard";

export interface EvalRunInput {
  mode: EvalMode;
  skill_id: string;
}

// ===== SSE Events =====
export interface SSEEvent {
  type: SSEEventType;
  session_id?: number;
  skill_id?: string;
  tool_name?: string;
  summary?: string;
  status?: string;
  round_count?: number;
  tool_call_count?: number;
  error?: string;
  arguments?: Record<string, unknown>;
  result?: unknown;
  text_delta?: string;
}

export type SSEEventType =
  | "agent_start"
  | "skill_loaded"
  | "llm_round_start"
  | "tool_call"
  | "tool_result"
  | "text_delta"
  | "guard_retry"
  | "final_answer"
  | "error";

// ===== Scenario Template =====
export interface ScenarioTemplate {
  id: string;
  label: string;
  description: string;
  icon: string;
  quickPrompts: string[];
}

// ===== Common =====
export interface CreateSessionInput {
  user_id?: string;
  title?: string;
  scenario?: string;
  skill_id?: string;
  skill_version?: string;
  context_policy?: { max_recent_messages: number };
}

export interface PostMessageInput {
  content: string;
  skill_id?: string;
  required_evidence?: string[];
}

export interface PostMessageResult {
  session_id: number;
  final_answer: string;
  round_count: number;
  tool_call_count: number;
}
