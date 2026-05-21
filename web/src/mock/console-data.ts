import type {
  Tool, Invocation, Skill, EvalSummary, EvalRun,
  AgentSession, PostMessageResult,
} from "../types";
import { mockSessions, mockMessages, mockToolCalls } from "./data";

const MOCK_DELAY = 300;

function delay(ms = MOCK_DELAY) {
  return new Promise((r) => setTimeout(r, ms));
}

// ===== Tools =====
const mockTools: Tool[] = [
  {
    name: "get_video_detail",
    display_name: "视频详情",
    category: "video",
    description: "Get one video detail from video-feed by video_id.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "get_video_detail",
        parameters: {
          type: "object",
          properties: {
            video_id: { type: "integer", description: "视频 ID" },
          },
          required: ["video_id"],
        },
      },
    },
  },
  {
    name: "get_hot_videos",
    display_name: "热榜视频",
    category: "video",
    description: "Get current hot rank video list.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "get_hot_videos",
        parameters: {
          type: "object",
          properties: {
            limit: { type: "integer", description: "返回数量" },
            category: { type: "string", description: "分类筛选", enum: ["tech", "life", "entertainment"] },
          },
        },
      },
    },
  },
  {
    name: "get_video_comments",
    display_name: "评论采集",
    category: "comment",
    description: "Fetch comments for a video.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "get_video_comments",
        parameters: {
          type: "object",
          properties: {
            video_id: { type: "integer", description: "视频 ID" },
            limit: { type: "integer", description: "数量限制" },
            sort: { type: "string", description: "排序方式", enum: ["hot", "time"] },
          },
          required: ["video_id"],
        },
      },
    },
  },
  {
    name: "get_author_profile",
    display_name: "作者信息",
    category: "author",
    description: "Get author profile by ID.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "get_author_profile",
        parameters: {
          type: "object",
          properties: {
            author_id: { type: "integer", description: "作者 ID" },
          },
          required: ["author_id"],
        },
      },
    },
  },
  {
    name: "list_author_videos",
    display_name: "作者作品",
    category: "author",
    description: "List videos by author.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "list_author_videos",
        parameters: {
          type: "object",
          properties: {
            author_id: { type: "integer", description: "作者 ID" },
            limit: { type: "integer", description: "数量限制" },
            period: { type: "string", description: "时间范围" },
          },
          required: ["author_id"],
        },
      },
    },
  },
  {
    name: "list_tag_videos",
    display_name: "标签视频",
    category: "tag",
    description: "List videos by tag.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "list_tag_videos",
        parameters: {
          type: "object",
          properties: {
            tag: { type: "string", description: "标签名称" },
            limit: { type: "integer", description: "数量限制" },
            period: { type: "string", description: "时间范围" },
          },
          required: ["tag"],
        },
      },
    },
  },
  {
    name: "analyze_video_comment_risk",
    display_name: "评论风险分析",
    category: "analysis",
    description: "Analyze comment risk for a video.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "analyze_video_comment_risk",
        parameters: {
          type: "object",
          properties: {
            video_id: { type: "integer", description: "视频 ID" },
            depth: { type: "string", description: "分析深度", enum: ["quick", "full"] },
          },
          required: ["video_id"],
        },
      },
    },
  },
  {
    name: "analyze_comment_risk",
    display_name: "评论风险分析(v1)",
    category: "analysis",
    description: "Legacy comment risk analysis.",
    read_only: true,
    schema: {
      type: "function",
      function: {
        name: "analyze_comment_risk",
        parameters: {
          type: "object",
          properties: {
            video_id: { type: "integer", description: "视频 ID" },
          },
          required: ["video_id"],
        },
      },
    },
  },
];

// ===== Invocations =====
const now = Date.now();
const mockInvocations: Invocation[] = [
  { id: 1, source: "agent_runtime", tool_name: "get_video_detail", status: "success", latency_ms: 156, result_summary: "视频播放量 128.5w，点赞 12.3w", session_id: 1, skill_id: "comment_risk_analysis", created_at: new Date(now - 29 * 60000).toISOString() },
  { id: 2, source: "agent_runtime", tool_name: "get_hot_videos", status: "success", latency_ms: 203, result_summary: "热榜前10中科技类占比40%", session_id: 1, created_at: new Date(now - 29 * 60000).toISOString() },
  { id: 3, source: "agent_runtime", tool_name: "get_video_comments", status: "success", latency_ms: 312, result_summary: "获取50条评论，正面68%", session_id: 1, skill_id: "comment_risk_analysis", created_at: new Date(now - 28 * 60000).toISOString() },
  { id: 4, source: "agent_runtime", tool_name: "analyze_video_comment_risk", status: "success", latency_ms: 89, result_summary: "风险等级: Attention", session_id: 2, skill_id: "comment_risk_analysis", created_at: new Date(now - 1.8 * 3600000).toISOString() },
  { id: 5, source: "manual_console", tool_name: "get_video_detail", status: "success", latency_ms: 120, result_summary: "视频播放量 89w", created_at: new Date(now - 2 * 3600000).toISOString() },
  { id: 6, source: "mcp_client", tool_name: "get_author_profile", status: "timeout", latency_ms: 30000, error_message: "获取作者信息超时", session_id: 3, created_at: new Date(now - 5 * 3600000).toISOString() },
  { id: 7, source: "agent_runtime", tool_name: "list_tag_videos", status: "error", latency_ms: 4500, error_message: "标签服务不可用", session_id: 4, skill_id: "tag_trend_analysis", created_at: new Date(now - 6 * 3600000).toISOString() },
  { id: 8, source: "manual_console", tool_name: "analyze_video_comment_risk", status: "success", latency_ms: 340, result_summary: "未发现明显风险", created_at: new Date(now - 8 * 3600000).toISOString() },
];

// ===== Skills =====
const mockSkills: Skill[] = [
  { id: "comment_risk_analysis", name: "评论风险分析", description: "识别评论区争议、攻击与刷屏风险", version: "1.0.0", status: "enabled", scenario: "comment_risk_analysis", allowed_tools: ["get_video_detail", "get_video_comments", "analyze_video_comment_risk"], required_evidence: ["get_video_detail"], prompt_template: "分析视频评论区的风险...", output_sections: ["结论", "证据", "建议"] },
  { id: "hot_rank_attribution", name: "热榜归因分析", description: "分析视频上热榜的原因与潜在风险", version: "1.0.0", status: "enabled", scenario: "hot_rank_attribution", allowed_tools: ["get_video_detail", "get_hot_videos", "get_video_comments"], required_evidence: ["get_video_detail"], prompt_template: "分析视频进入热榜的原因...", output_sections: ["核心发现", "上榜原因", "风险评估"] },
  { id: "author_support_evaluation", name: "作者扶持评估", description: "评估作者表现与扶持价值", version: "1.0.0", status: "disabled", scenario: "author_support_evaluation", allowed_tools: ["get_author_profile", "list_author_videos"], required_evidence: ["get_author_profile"], prompt_template: "评估作者扶持价值...", output_sections: ["基本信息", "表现分析", "扶持建议"] },
  { id: "tag_trend_analysis", name: "标签趋势分析", description: "洞察标签下的内容风向与趋势", version: "1.0.0", status: "enabled", scenario: "tag_trend_analysis", allowed_tools: ["list_tag_videos", "get_hot_videos"], required_evidence: [], prompt_template: "分析标签趋势...", output_sections: ["标签概览", "内容趋势", "运营建议"] },
  { id: "content_review_summary", name: "内容复盘摘要", description: "综合复盘视频内容表现", version: "0.9.0", status: "disabled", scenario: "content_review_summary", allowed_tools: ["get_video_detail", "get_video_comments", "get_author_profile"], required_evidence: ["get_video_detail"], prompt_template: "复盘视频内容表现...", output_sections: ["数据概览", "表现分析", "改进建议"] },
];

// ===== Eval =====
const mockEvalSummary: EvalSummary = {
  tool_call_success_rate: 0.875,
  tool_call_error_count: 3,
  unauthorized_tool_call_count: 1,
  evidence_complete_final_answer_count: 12,
  average_tool_latency_ms: 245,
  average_tool_call_count: 2.8,
  skill_success_count: 15,
  skill_failure_count: 2,
  unsupported_metrics: ["guard_retry_count", "average_round_count"],
};

// ===== Mock API Handler =====
export const mockApi = {
  // Health
  health: async () => { await delay(100); return { status: "ok" }; },

  // Sessions
  listSessions: async () => { await delay(); return { sessions: mockSessions }; },
  getSession: async (id: number) => {
    await delay();
    const session = mockSessions.find((s) => s.id === id);
    const messages = mockMessages.filter((m) => m.session_id === id);
    return { session: session || mockSessions[0], messages };
  },
  createSession: async (body: Record<string, unknown>) => {
    await delay();
    const newSession: AgentSession = {
      id: mockSessions.length + 1,
      user_id: (body.user_id as string) || "ops-001",
      title: (body.title as string) || "新会话",
      scenario: (body.scenario as string) || "",
      status: "active",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };
    return { session: newSession };
  },
  postMessage: async (sessionId: number) => {
    await delay(1500);
    const result: PostMessageResult = {
      session_id: sessionId,
      final_answer: "## Mock 诊断结果\n\n这是一个模拟的 Agent 回复。实际使用时会调用工具采集数据并生成基于证据的诊断报告。\n\n### 发现\n- 数据指标正常\n- 需要持续关注\n\n### 建议\n1. 持续监控\n2. 定期复查",
      round_count: 2,
      tool_call_count: 3,
    };
    return result;
  },
  listToolCalls: async (sessionId: number) => {
    await delay();
    return { tool_calls: mockToolCalls.filter((tc) => tc.session_id === sessionId) };
  },

  // Gateway
  listTools: async () => { await delay(); return { tools: mockTools }; },
  getTool: async (name: string) => {
    await delay();
    const tool = mockTools.find((t) => t.name === name);
    if (!tool) throw new Error("Tool not found");
    return { tool };
  },
  callTool: async (name: string) => {
    await delay(800);
    return {
      invocation: {
        id: Date.now(),
        source: "manual_console" as const,
        tool_name: name,
        status: "success" as const,
        latency_ms: 234,
        result_summary: `Mock result for ${name}`,
        created_at: new Date().toISOString(),
      },
      result: {
        tool_name: name,
        summary: `Mock ${name} 调用成功`,
        data: { mock: true, note: "这是模拟数据" },
      },
    };
  },
  listInvocations: async (filters?: Record<string, unknown>) => {
    await delay();
    let inv = [...mockInvocations];
    if (filters?.source) inv = inv.filter((i) => i.source === filters.source);
    if (filters?.tool_name) inv = inv.filter((i) => i.tool_name === filters.tool_name);
    if (filters?.status) inv = inv.filter((i) => i.status === filters.status);
    if (filters?.session_id) inv = inv.filter((i) => i.session_id === Number(filters.session_id));
    if (filters?.skill_id) inv = inv.filter((i) => i.skill_id === filters.skill_id);
    const limit = Number(filters?.limit) || 20;
    return { invocations: inv.slice(0, limit) };
  },
  getInvocation: async (id: number) => {
    await delay();
    const inv = mockInvocations.find((i) => i.id === id);
    if (!inv) throw new Error("Invocation not found");
    return { invocation: inv };
  },

  // Skills
  listSkills: async () => { await delay(); return { skills: mockSkills }; },
  getSkill: async (id: string) => {
    await delay();
    const skill = mockSkills.find((s) => s.id === id);
    if (!skill) throw new Error("Skill not found");
    return { skill };
  },
  createSkill: async (body: Record<string, unknown>) => {
    await delay();
    return { skill: { ...body, status: "enabled" } as Skill };
  },
  updateSkill: async (id: string, body: Record<string, unknown>) => {
    await delay();
    return { skill: { ...body, id } as Skill };
  },
  enableSkill: async (id: string) => {
    await delay();
    const skill = mockSkills.find((s) => s.id === id);
    return { skill: { ...skill, status: "enabled" } as Skill };
  },
  disableSkill: async (id: string) => {
    await delay();
    const skill = mockSkills.find((s) => s.id === id);
    return { skill: { ...skill, status: "disabled" } as Skill };
  },

  // Eval
  evalSummary: async () => { await delay(); return mockEvalSummary; },
  evalSkillSummary: async () => {
    await delay();
    return {
      ...mockEvalSummary,
      tool_call_success_rate: 0.92,
      skill_failure_count: 0,
    };
  },
  createEvalRun: async (body: Record<string, unknown>) => {
    await delay();
    return {
      run: {
        id: Date.now(),
        mode: body.mode as string,
        skill_id: body.skill_id as string,
        summary: mockEvalSummary,
      } as EvalRun,
    };
  },
};
