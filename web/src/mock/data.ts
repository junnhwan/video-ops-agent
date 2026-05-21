import type { AgentSession, AgentMessage, AgentToolCall, ScenarioTemplate } from "../types";

export const mockScenarios: ScenarioTemplate[] = [
  {
    id: "hot_rank_analysis",
    label: "热榜归因",
    description: "分析视频上热榜的原因与潜在风险",
    icon: "TrendingUp",
    quickPrompts: [
      "分析一下视频 123 为什么上热榜，有没有运营风险？",
      "最近热榜前三的视频有什么共同特征？",
      "视频 456 突然进入热榜，帮我分析原因。",
    ],
  },
  {
    id: "comment_risk_analysis",
    label: "评论风险",
    description: "识别评论区争议、攻击与刷屏风险",
    icon: "ShieldAlert",
    quickPrompts: [
      "分析视频 123 的评论区有没有争议或攻击风险？",
      "看看这篇热门视频评论区是否有异常行为？",
      "视频 789 的评论是否存在刷屏现象？",
    ],
  },
  {
    id: "author_profile_analysis",
    label: "作者画像",
    description: "评估作者表现与扶持价值",
    icon: "UserCircle",
    quickPrompts: [
      "作者 8 最近表现怎么样，值得扶持吗？",
      "分析一下这个账号的内容风格变化趋势。",
      "这个创作者最近的作品有什么特征？",
    ],
  },
  {
    id: "tag_trend_analysis",
    label: "标签趋势",
    description: "洞察标签下的内容风向与趋势",
    icon: "Tag",
    quickPrompts: [
      "#Go 后端 这个标签最近内容表现怎么样？",
      "分析 #机器学习 话题下的热门内容趋势。",
      "#生活记录 标签最近的评论区情绪如何？",
    ],
  },
];

const now = Date.now();
const hour = 3600_000;
const min = 60_000;

export const mockSessions: AgentSession[] = [
  {
    id: 1,
    user_id: "ops-001",
    title: "视频 123 热榜归因分析",
    scenario: "hot_rank_analysis",
    status: "active",
    last_message_preview: "该视频进入热榜主要由互动密度驱动，评论区情绪偏正面",
    created_at: new Date(now - 30 * min).toISOString(),
    updated_at: new Date(now - 5 * min).toISOString(),
  },
  {
    id: 2,
    user_id: "ops-001",
    title: "视频 456 评论风险评估",
    scenario: "comment_risk_analysis",
    status: "active",
    last_message_preview: "评论区存在 3 条疑似刷屏评论，建议关注",
    created_at: new Date(now - 2 * hour).toISOString(),
    updated_at: new Date(now - 45 * min).toISOString(),
  },
  {
    id: 3,
    user_id: "ops-001",
    title: "作者 8 近期表现诊断",
    scenario: "author_profile_analysis",
    status: "closed",
    last_message_preview: "作者近 30 天发布 12 条内容，平均互动率 4.2%",
    created_at: new Date(now - 24 * hour).toISOString(),
    updated_at: new Date(now - 23 * hour).toISOString(),
  },
  {
    id: 4,
    user_id: "ops-001",
    title: "#Go后端 标签趋势洞察",
    scenario: "tag_trend_analysis",
    status: "active",
    last_message_preview: "该标签近 7 天内容量增长 18%，技术分享类占比最高",
    created_at: new Date(now - 3 * hour).toISOString(),
    updated_at: new Date(now - 10 * min).toISOString(),
  },
  {
    id: 5,
    user_id: "ops-001",
    title: "视频 789 热榜异常检测",
    scenario: "hot_rank_analysis",
    status: "error",
    last_message_preview: "数据采集超时，无法获取热榜排名数据",
    created_at: new Date(now - 48 * hour).toISOString(),
    updated_at: new Date(now - 47.5 * hour).toISOString(),
  },
  {
    id: 6,
    user_id: "ops-001",
    title: "作者 15 深度画像评估",
    scenario: "author_profile_analysis",
    status: "closed",
    last_message_preview: "该作者粉丝质量较高，建议纳入重点扶持名单",
    created_at: new Date(now - 72 * hour).toISOString(),
    updated_at: new Date(now - 71 * hour).toISOString(),
  },
];

export const mockMessages: AgentMessage[] = [
  // ── Session 1: hot_rank_analysis ──
  {
    id: 1,
    session_id: 1,
    role: "user",
    content: "分析一下视频 123 为什么上热榜，有没有运营风险？",
    created_at: new Date(now - 30 * min).toISOString(),
  },
  {
    id: 2,
    session_id: 1,
    role: "assistant",
    content: `## 热榜归因诊断：视频 123

### 核心发现

视频 123 在过去 24 小时内进入平台热榜 **Top 5**，主要驱动力如下：

| 指标 | 数值 | 同类排名百分位 |
|------|------|---------------|
| 播放量 | 128.5w | Top 3% |
| 点赞率 | 9.6% | Top 5% |
| 完播率 | 67.3% | Top 8% |
| 分享率 | 6.9% | Top 2% |

### 上榜原因分析

1. **内容质量** — 科技类科普内容，选题精准切中 Go 后端微服务热点话题，信息密度高
2. **发布时机** — 工作日早高峰发布，科技受众活跃时段匹配度高达 89%
3. **传播路径** — 首波由 3 个万粉科技博主转发引爆，二次传播覆盖约 45w 用户
4. **标签协同** — \`#Go后端\` + \`#微服务\` 双标签叠加，进入两个标签池推荐流

### 风险评估

> 风险等级：**Attention** (需持续关注)

- **评论争议** — 评论区出现个别引战言论（约占总评论 3.2%），围绕「Go vs Java」技术选型产生对立
- **内容合规** — 视频中引用的开源项目协议为 GPL-3.0，存在潜在版权争议
- **刷量嫌疑** — 分享路径中有约 1.2% 的异常高频转发账号，需关注

### 建议操作

- 开启评论预警，设置关键词「Java」「抄袭」自动提醒
- 准备舆情回应话术模板
- 持续监控异常分享账号后续行为`,
    content_summary: "热榜归因与风险初判",
    created_at: new Date(now - 28 * min).toISOString(),
  },
  {
    id: 3,
    session_id: 1,
    role: "user",
    content: "具体有哪些风险？评论区争议的详细情况能展开吗？",
    created_at: new Date(now - 25 * min).toISOString(),
  },
  {
    id: 4,
    session_id: 1,
    role: "assistant",
    content: `## 评论争议详细分析

### 情感分布

对 50 条热门评论进行情感分析，结果如下：

- 正面评价：**68%** — 集中于「讲得好」「干货」「收藏了」
- 中性讨论：**28%** — 技术讨论，如「goroutine 这段可以再深入」
- 负面/引战：**4%** — 主要是技术选型争论

### 高风险评论示例

| 评论内容 | 风险类型 | 情感分 |
|---------|---------|--------|
| 「Go 写微服务不如 Java，作者不懂装懂」 | 对立言论 | 0.12 |
| 「又一个抄 Spring Boot 的」 | 恶意贬低 | 0.08 |
| 「这种内容也能上热榜？平台审核干嘛吃的」 | 平台攻击 | 0.05 |

### 建议措施

1. **暂不干预** — 当前争议规模可控，过度干预可能引发反效果
2. **设置预警** — 对「Java」「抄袭」「平台」关键词设置实时提醒
3. **准备话术** — 如争议升级，以「技术多元、各有场景」方向温和回应`,
    content_summary: "评论风险详细分析",
    created_at: new Date(now - 23 * min).toISOString(),
  },

  // ── Session 2: comment_risk_analysis ──
  {
    id: 5,
    session_id: 2,
    role: "user",
    content: "帮我分析视频 456 的评论区，看看有没有刷屏或攻击行为？",
    created_at: new Date(now - 2 * hour).toISOString(),
  },
  {
    id: 6,
    session_id: 2,
    role: "assistant",
    content: `## 评论风险诊断：视频 456

### 总体评估

> 风险等级：**Warning** (需要介入处理)

视频 456 的评论区检测到 **明显的刷屏行为** 和少量人身攻击，建议尽快处理。

### 检测结果

| 风险类型 | 检出数量 | 占比 | 严重程度 |
|---------|---------|------|---------|
| 刷屏评论 | 23 条 | 4.6% | 高 |
| 人身攻击 | 5 条 | 1.0% | 高 |
| 引战言论 | 12 条 | 2.4% | 中 |
| 广告引流 | 3 条 | 0.6% | 低 |

### 刷屏模式分析

检测到 **两种刷屏模式**：

1. **重复内容刷屏** — 同一用户在 30 分钟内发布 8 条内容高度相似的评论，均为「关注我回粉」类引流话术
2. **复制粘贴攻击** — 多个账号（疑似水军）复制同一段负面评论，在短时间内集中发布

### 已触发的风控规则

- 同用户 10 分钟内评论数 > 3 条
- 评论内容相似度 > 85% 的重复检测
- 包含人身攻击关键词
- 账号注册时间 < 7 天 + 评论频率异常

### 建议操作

1. **立即处理** — 删除 23 条刷屏评论 + 5 条攻击评论
2. **账号处置** — 对刷屏账号禁言 7 天，攻击账号禁言 30 天
3. **防护策略** — 开启评论频率限制（同一用户 1 分钟内最多 1 条）`,
    content_summary: "评论风险诊断 - 刷屏与攻击检测",
    created_at: new Date(now - 1.8 * hour).toISOString(),
  },

  // ── Session 3: author_profile_analysis ──
  {
    id: 7,
    session_id: 3,
    role: "user",
    content: "作者 8 最近表现怎么样，值得扶持吗？",
    created_at: new Date(now - 24 * hour).toISOString(),
  },
  {
    id: 8,
    session_id: 3,
    role: "assistant",
    content: `## 作者画像：作者 8

### 基本信息

| 维度 | 数据 |
|-----|------|
| 粉丝数 | 12.8w |
| 总作品数 | 156 |
| 平均播放量 | 8.3w |
| 创作周期 | 2.3 年 |
| 主领域 | 科技/编程 |

### 近 30 天表现

- **发布频率**：8 条，周均 1.9 条（稳定更新）
- **平均播放量**：11.2w（较历史均值 +34%）
- **最高播放**：28.7w（视频「Go 并发编程实战」）
- **互动率**：8.7%（高于同领域平均 6.2%）
- **完播率**：72.1%（优秀）

### 内容风格分析

作者内容集中在 **后端编程** 领域，风格特征：

1. **教程型** 占 60% — 系统性讲解技术方案，信息密度高
2. **实战型** 占 25% — 项目演示 + 代码走读，实操性强
3. **观点型** 占 15% — 技术选型对比，有一定争议性但互动效果好

### 扶持价值评估

> 综合评分：**A 级（推荐扶持）**

- 内容质量稳定，有明确的专业壁垒
- 粉丝增长曲线健康，无异常波动
- 完播率和互动率均高于同类创作者
- 建议扶持方向：流量加权 + 独家内容合作

### 风险提示

- 近期有 2 条视频评论区出现技术争论，但未升级
- 内容高度依赖个人 IP，需关注作者持续性`,
    content_summary: "作者画像与扶持价值评估",
    created_at: new Date(now - 23.5 * hour).toISOString(),
  },

  // ── Session 4: tag_trend_analysis ──
  {
    id: 9,
    session_id: 4,
    role: "user",
    content: "#Go后端 这个标签最近内容表现怎么样？有什么趋势？",
    created_at: new Date(now - 3 * hour).toISOString(),
  },
  {
    id: 10,
    session_id: 4,
    role: "assistant",
    content: `## 标签趋势：#Go后端

### 标签概览

| 指标 | 近 7 天 | 环比变化 |
|-----|--------|---------|
| 关联视频数 | 1,247 | +12.3% |
| 总播放量 | 3,280w | +18.7% |
| 平均播放量 | 2.63w | +5.6% |
| 新增创作者 | 89 | +8.2% |

### 内容趋势

1. **热门子话题** — 微服务架构（32%）、并发编程（24%）、性能优化（18%）
2. **上升话题** — \`泛型\` 相关内容环比增长 **47%**，Go 1.22 泛型改进引发关注
3. **下降话题** — \`\`RPC 框架对比\` 类内容持续走低，市场饱和

### 内容质量分布

- **优质内容**（播放 > 5w）：占比 15%，以系统教程为主
- **中等内容**（播放 1-5w）：占比 45%，实战分享居多
- **低质内容**（播放 < 1w）：占比 40%，含大量搬运和低信息量内容

### 运营建议

1. **扶持方向** — 重点扶持泛型、性能优化类原创内容
2. **话题运营** — 策划「Go 2024 技术演进」专题，聚合优质创作者
3. **内容治理** — 标签下低质搬运内容占比偏高，建议加强审核`,
    content_summary: "标签趋势洞察与运营建议",
    created_at: new Date(now - 2.7 * hour).toISOString(),
  },

  // ── Session 5: hot_rank_analysis (error case) ──
  {
    id: 11,
    session_id: 5,
    role: "user",
    content: "视频 789 是不是刷量上热榜的？帮我查一下。",
    created_at: new Date(now - 48 * hour).toISOString(),
  },
  {
    id: 12,
    session_id: 5,
    role: "assistant",
    content: `## 热榜异常检测：视频 789

### 检测结果

> ⚠️ 部分数据采集失败，以下分析基于已获取的不完整数据。

### 已获取的数据

视频 789 在过去 6 小时内从热榜第 47 名飙升至第 3 名，上升速度异常：

| 指标 | 数值 | 异常判断 |
|-----|------|---------|
| 播放量增长 | +320w / 6h | ⚠️ 异常偏高 |
| 点赞增长 | +1.2w / 6h | 正常 |
| 评论增长 | +89 / 6h | ⚠️ 明显偏低 |

### 初步判断

播放量与互动量的严重不匹配（播放/评论比 = 35,955:1，正常范围约 5,000-15,000:1）**高度疑似刷量**。

> 由于 \`get_author_profile\` 接口超时，无法获取作者历史数据做交叉验证。建议重新发起分析。`,
    content_summary: "热榜异常检测 - 部分数据采集失败",
    created_at: new Date(now - 47.5 * hour).toISOString(),
  },
];

export const mockToolCalls: AgentToolCall[] = [
  // ── Session 1 tool calls ──
  {
    id: 1,
    session_id: 1,
    message_id: 2,
    tool_name: "get_video_detail",
    arguments_json: JSON.stringify({ video_id: 123 }),
    result_json: JSON.stringify({
      video_id: 123,
      title: "Go 微服务实战：从零到生产",
      play_count: 1285000,
      like_count: 123000,
      share_count: 8900,
      comment_count: 4560,
      duration: "12:34",
      tags: ["Go后端", "微服务", "架构设计"],
      publish_time: "2024-12-15T08:30:00Z",
      author_id: 8,
    }),
    result_summary: "视频播放量 128.5w，点赞 12.3w，分享 8.9k，完播率 67.3%。内容为科技类科普，标签 #Go后端 #微服务。",
    latency_ms: 156,
    status: "success",
    created_at: new Date(now - 29 * min).toISOString(),
  },
  {
    id: 2,
    session_id: 1,
    message_id: 2,
    tool_name: "get_hot_videos",
    arguments_json: JSON.stringify({ limit: 10, category: "tech" }),
    result_json: JSON.stringify({
      rank: [
        { position: 1, video_id: 99, title: "...", play_count: 5200000 },
        { position: 5, video_id: 123, title: "Go 微服务实战", play_count: 1285000 },
      ],
      tech_ratio: 0.4,
      life_ratio: 0.3,
    }),
    result_summary: "当前热榜前 10 中，科技类占比 40%，生活类 30%。视频 123 排名第 5。",
    latency_ms: 203,
    status: "success",
    created_at: new Date(now - 29 * min).toISOString(),
  },
  {
    id: 3,
    session_id: 1,
    message_id: 2,
    tool_name: "get_video_comments",
    arguments_json: JSON.stringify({ video_id: 123, limit: 50, sort: "hot" }),
    result_json: JSON.stringify({
      total: 4560,
      sampled: 50,
      sentiment: { positive: 0.68, neutral: 0.28, negative: 0.04 },
      top_comments: [
        { id: "c1", content: "讲得太好了，收藏！", likes: 2300, sentiment: "positive" },
        { id: "c2", content: "goroutine 那段可以再深入", likes: 890, sentiment: "neutral" },
      ],
    }),
    result_summary: "获取 50 条评论。正面评价 68%，中性 28%，负面 4%。负面评论集中于对技术观点的争论。",
    latency_ms: 312,
    status: "success",
    created_at: new Date(now - 28 * min).toISOString(),
  },
  {
    id: 4,
    session_id: 1,
    message_id: 4,
    tool_name: "analyze_comment_risk",
    arguments_json: JSON.stringify({ video_id: 123, depth: "full" }),
    result_json: JSON.stringify({
      risk_level: "attention",
      flagged_comments: [
        { content: "Go 写微服务不如 Java", type: "confrontation", score: 0.12 },
        { content: "又一个抄 Spring Boot 的", type: "disparagement", score: 0.08 },
      ],
      rules_triggered: ["confrontation_detection", "high_frequency_repeat"],
    }),
    result_summary: "风险等级: Attention。发现 3 条引战评论，1 条疑似刷屏。触发规则: 对立言论、高频重复。",
    latency_ms: 89,
    status: "success",
    created_at: new Date(now - 23.5 * min).toISOString(),
  },

  // ── Session 2 tool calls (comment risk) ──
  {
    id: 5,
    session_id: 2,
    message_id: 6,
    tool_name: "get_video_detail",
    arguments_json: JSON.stringify({ video_id: 456 }),
    result_json: JSON.stringify({
      video_id: 456,
      title: "周末 vlog | 记录生活",
      play_count: 890000,
      like_count: 45000,
      comment_count: 500,
    }),
    result_summary: "视频播放量 89w，评论 500 条。生活 vlog 类内容。",
    latency_ms: 134,
    status: "success",
    created_at: new Date(now - 1.9 * hour).toISOString(),
  },
  {
    id: 6,
    session_id: 2,
    message_id: 6,
    tool_name: "get_video_comments",
    arguments_json: JSON.stringify({ video_id: 456, limit: 100 }),
    result_json: JSON.stringify({
      total: 500,
      sampled: 100,
      spam_detected: 23,
      attack_detected: 5,
      confrontation_detected: 12,
      ads_detected: 3,
    }),
    result_summary: "采集 100 条评论，检测到刷屏 23 条、人身攻击 5 条、引战 12 条、广告 3 条。",
    latency_ms: 278,
    status: "success",
    created_at: new Date(now - 1.85 * hour).toISOString(),
  },
  {
    id: 7,
    session_id: 2,
    message_id: 6,
    tool_name: "analyze_comment_risk",
    arguments_json: JSON.stringify({ video_id: 456, depth: "full", include_spam: true }),
    result_json: JSON.stringify({
      risk_level: "warning",
      spam_patterns: ["repeat_content", "copy_paste_attack"],
      spam_accounts: ["user_x1", "user_x2", "user_x3"],
      attack_keywords: ["人身攻击词1", "人身攻击词2"],
      rule_violations: [
        "comment_frequency_limit",
        "content_similarity_85",
        "attack_keywords",
        "new_account_abnormal",
      ],
    }),
    result_summary: "风险等级: Warning。两种刷屏模式：重复内容 + 复制粘贴攻击。触发 4 条风控规则。",
    latency_ms: 167,
    status: "success",
    created_at: new Date(now - 1.8 * hour).toISOString(),
  },

  // ── Session 3 tool calls (author profile) ──
  {
    id: 8,
    session_id: 3,
    message_id: 8,
    tool_name: "get_author_profile",
    arguments_json: JSON.stringify({ author_id: 8 }),
    result_json: JSON.stringify({
      author_id: 8,
      name: "GoTech实验室",
      followers: 128000,
      total_videos: 156,
      avg_plays: 83000,
      main_category: "tech",
      join_date: "2022-06-15",
    }),
    result_summary: "作者「GoTech实验室」，粉丝 12.8w，总作品 156 条，主领域科技/编程。",
    latency_ms: 189,
    status: "success",
    created_at: new Date(now - 23.8 * hour).toISOString(),
  },
  {
    id: 9,
    session_id: 3,
    message_id: 8,
    tool_name: "list_author_videos",
    arguments_json: JSON.stringify({ author_id: 8, limit: 20, period: "30d" }),
    result_json: JSON.stringify({
      period: "30d",
      count: 8,
      avg_plays: 112000,
      max_plays: 287000,
      max_video: { title: "Go 并发编程实战", plays: 287000 },
      interaction_rate: 0.087,
      completion_rate: 0.721,
    }),
    result_summary: "近 30 天发布 8 条，平均播放 11.2w（+34%），互动率 8.7%，完播率 72.1%。",
    latency_ms: 245,
    status: "success",
    created_at: new Date(now - 23.7 * hour).toISOString(),
  },

  // ── Session 4 tool calls (tag trend) ──
  {
    id: 10,
    session_id: 4,
    message_id: 10,
    tool_name: "list_tag_videos",
    arguments_json: JSON.stringify({ tag: "Go后端", period: "7d", limit: 50 }),
    result_json: JSON.stringify({
      tag: "Go后端",
      period: "7d",
      video_count: 1247,
      total_plays: 32800000,
      avg_plays: 26300,
      new_creators: 89,
      sub_topics: {
        "微服务架构": 0.32,
        "并发编程": 0.24,
        "性能优化": 0.18,
        "其他": 0.26,
      },
    }),
    result_summary: "7 天内关联视频 1,247 条，总播放 3,280w，新增创作者 89 人。微服务架构占比最高（32%）。",
    latency_ms: 312,
    status: "success",
    created_at: new Date(now - 2.8 * hour).toISOString(),
  },
  {
    id: 11,
    session_id: 4,
    message_id: 10,
    tool_name: "get_hot_videos",
    arguments_json: JSON.stringify({ tag: "Go后端", limit: 20 }),
    result_json: JSON.stringify({
      top_videos: [
        { title: "Go 1.22 泛型全解析", plays: 520000, tag: "泛型" },
        { title: "微服务链路追踪实战", plays: 380000, tag: "微服务" },
        { title: "Go 性能调优：pprof 详解", plays: 290000, tag: "性能优化" },
      ],
      rising_topic: "泛型",
      rising_rate: 0.47,
    }),
    result_summary: "泛型相关内容环比增长 47%，为当前上升最快子话题。",
    latency_ms: 198,
    status: "success",
    created_at: new Date(now - 2.75 * hour).toISOString(),
  },

  // ── Session 5 tool calls (error case) ──
  {
    id: 12,
    session_id: 5,
    message_id: 12,
    tool_name: "get_video_detail",
    arguments_json: JSON.stringify({ video_id: 789 }),
    result_json: JSON.stringify({
      video_id: 789,
      title: "三天学会Go语言",
      play_count: 5200000,
      like_count: 89000,
      comment_count: 89,
      rank_change: "+44 in 6h",
    }),
    result_summary: "视频播放量 520w，评论仅 89 条。6 小时内排名飙升 44 位。",
    latency_ms: 145,
    status: "success",
    created_at: new Date(now - 47.8 * hour).toISOString(),
  },
  {
    id: 13,
    session_id: 5,
    message_id: 12,
    tool_name: "get_author_profile",
    arguments_json: JSON.stringify({ author_id: 42 }),
    result_summary: undefined,
    latency_ms: 30000,
    status: "timeout",
    error_message: "获取作者信息超时（30s），上游服务未响应。建议稍后重试。",
    created_at: new Date(now - 47.6 * hour).toISOString(),
  },
  {
    id: 14,
    session_id: 5,
    message_id: 12,
    tool_name: "get_video_comments",
    arguments_json: JSON.stringify({ video_id: 789, limit: 50 }),
    result_summary: "仅获取到 89 条评论（预期 > 2000）。播放/评论比 = 58,427:1，严重偏离正常范围。",
    latency_ms: 267,
    status: "success",
    created_at: new Date(now - 47.5 * hour).toISOString(),
  },
];
