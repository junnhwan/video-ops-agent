package skills

func BuiltinSkills() []DiagnosisSkill {
	return cloneSkills([]DiagnosisSkill{
		{
			ID:               "hot_rank_attribution",
			Name:             "热榜归因分析",
			Description:      "分析视频进入热榜或热度变化的关键内容、作者和互动因素。",
			Version:          "1.0.0",
			Status:           SkillStatusEnabled,
			Scenario:         "hot_rank_analysis",
			AllowedTools:     []string{"get_hot_videos", "get_video_detail", "get_author_profile"},
			RequiredEvidence: []string{"get_hot_videos", "get_video_detail"},
			PromptTemplate:   "围绕热榜位置、视频内容、互动指标和作者画像进行归因，所有结论必须引用工具证据。",
			OutputSections:   []string{"结论", "关键证据", "影响因素", "运营建议"},
			RiskNotes:        []string{"不要把热度变化归因为未观测到的推荐策略。"},
		},
		{
			ID:               "comment_risk_analysis",
			Name:             "评论风险分析",
			Description:      "识别评论区是否存在敏感词、重复内容、负面反馈和异常互动。",
			Version:          "1.0.0",
			Status:           SkillStatusEnabled,
			Scenario:         "comment_risk_analysis",
			AllowedTools:     []string{"get_video_detail", "analyze_video_comment_risk"},
			RequiredEvidence: []string{"get_video_detail", "analyze_video_comment_risk"},
			PromptTemplate:   "围绕评论风险等级、命中规则和代表证据给出运营判断，避免脱离评论证据扩大结论。",
			OutputSections:   []string{"结论", "命中规则", "代表证据", "运营建议"},
			RiskNotes:        []string{"风险判断必须来自评论分析工具返回的证据。"},
		},
		{
			ID:               "author_support_evaluation",
			Name:             "作者扶持评估",
			Description:      "评估作者内容质量、近期表现和是否适合运营扶持。",
			Version:          "1.0.0",
			Status:           SkillStatusEnabled,
			Scenario:         "author_support_evaluation",
			AllowedTools:     []string{"get_author_profile", "list_author_videos", "get_video_detail"},
			RequiredEvidence: []string{"get_author_profile", "list_author_videos"},
			PromptTemplate:   "结合作者画像、近期视频和互动数据评估扶持价值，并明确证据不足之处。",
			OutputSections:   []string{"结论", "作者表现", "内容证据", "扶持建议"},
			RiskNotes:        []string{"不要承诺平台资源或预测未验证的增长结果。"},
		},
		{
			ID:               "tag_trend_analysis",
			Name:             "标签趋势分析",
			Description:      "分析标签下内容表现、热度趋势和运营机会。",
			Version:          "1.0.0",
			Status:           SkillStatusEnabled,
			Scenario:         "tag_trend_analysis",
			AllowedTools:     []string{"list_tag_videos", "get_hot_videos", "get_video_detail"},
			RequiredEvidence: []string{"list_tag_videos"},
			PromptTemplate:   "基于标签视频列表和可用热榜证据分析趋势，不要虚构站外热点或历史走势。",
			OutputSections:   []string{"结论", "趋势证据", "代表内容", "运营动作"},
			RiskNotes:        []string{"趋势判断仅限当前工具可见样本。"},
		},
		{
			ID:               "content_review_summary",
			Name:             "内容复盘摘要",
			Description:      "汇总单条内容的表现、评论反馈和后续优化建议。",
			Version:          "1.0.0",
			Status:           SkillStatusEnabled,
			Scenario:         "content_review_summary",
			AllowedTools:     []string{"get_video_detail", "get_video_comments", "analyze_video_comment_risk"},
			RequiredEvidence: []string{"get_video_detail"},
			PromptTemplate:   "围绕内容表现、评论反馈和风险点做复盘，输出可执行但不夸大的优化建议。",
			OutputSections:   []string{"结论", "内容表现", "评论反馈", "优化建议"},
			RiskNotes:        []string{"没有评论风险工具证据时，不要断言评论区风险等级。"},
		},
	})
}

func cloneSkills(skills []DiagnosisSkill) []DiagnosisSkill {
	cloned := make([]DiagnosisSkill, 0, len(skills))
	for _, skill := range skills {
		cloned = append(cloned, cloneSkill(skill))
	}
	return cloned
}

func cloneSkill(skill DiagnosisSkill) DiagnosisSkill {
	skill.AllowedTools = append([]string(nil), skill.AllowedTools...)
	skill.RequiredEvidence = append([]string(nil), skill.RequiredEvidence...)
	skill.OutputSections = append([]string(nil), skill.OutputSections...)
	skill.RiskNotes = append([]string(nil), skill.RiskNotes...)
	return skill
}
