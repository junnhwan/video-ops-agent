package skills

func AllowedToolSet(skill DiagnosisSkill) map[string]struct{} {
	return stringSet(skill.AllowedTools)
}

func RequiredEvidence(skill DiagnosisSkill) []string {
	return append([]string(nil), skill.RequiredEvidence...)
}
