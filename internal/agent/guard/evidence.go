package guard

import "strings"

type EvidenceCheck struct {
	Complete      bool
	RequiredTools []string
	CalledTools   []string
	MissingTools  []string
}

func CheckRequired(requiredTools []string, calledTools []string) EvidenceCheck {
	called := make(map[string]struct{}, len(calledTools))
	for _, tool := range calledTools {
		normalized := strings.TrimSpace(tool)
		if normalized == "" {
			continue
		}
		called[normalized] = struct{}{}
	}

	missing := make([]string, 0)
	required := make([]string, 0, len(requiredTools))
	for _, tool := range requiredTools {
		normalized := strings.TrimSpace(tool)
		if normalized == "" {
			continue
		}
		required = append(required, normalized)
		if _, ok := called[normalized]; !ok {
			missing = append(missing, normalized)
		}
	}

	return EvidenceCheck{
		Complete:      len(missing) == 0,
		RequiredTools: required,
		CalledTools:   calledTools,
		MissingTools:  missing,
	}
}

func RetryInstruction(missingTools []string) string {
	missing := make([]string, 0, len(missingTools))
	for _, tool := range missingTools {
		tool = strings.TrimSpace(tool)
		if tool != "" {
			missing = append(missing, tool)
		}
	}
	if len(missing) == 0 {
		return ""
	}
	return "Evidence is incomplete. Call the missing tools before producing the final report. Missing tools: " + strings.Join(missing, ", ") + "."
}
