package skills

import (
	"strings"
)

func RenderPrompt(skill DiagnosisSkill) string {
	var builder strings.Builder
	builder.WriteString("Active diagnosis skill: ")
	builder.WriteString(skill.Name)
	builder.WriteString("\nSkill instructions:\n")
	builder.WriteString(strings.TrimSpace(skill.PromptTemplate))
	builder.WriteString("\nRequired output sections:\n")
	for _, section := range skill.OutputSections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		builder.WriteString("- ")
		builder.WriteString(section)
		builder.WriteString("\n")
	}
	if len(skill.RiskNotes) > 0 {
		builder.WriteString("Risk notes:\n")
		for _, note := range skill.RiskNotes {
			note = strings.TrimSpace(note)
			if note == "" {
				continue
			}
			builder.WriteString("- ")
			builder.WriteString(note)
			builder.WriteString("\n")
		}
	}
	return builder.String()
}
