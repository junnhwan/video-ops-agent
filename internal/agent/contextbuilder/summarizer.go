package contextbuilder

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func CompactToolResult(raw string, policy ContextPolicy) string {
	policy = normalizePolicy(policy)
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return truncateText(redactText(raw), policy.MaxToolResultChars)
	}

	sanitized := sanitizeValue(value, policy)
	encoded, err := json.Marshal(sanitized)
	if err != nil {
		return truncateText(redactText(raw), policy.MaxToolResultChars)
	}
	return truncateText(string(encoded), policy.MaxToolResultChars)
}

func sanitizeValue(value any, policy ContextPolicy) any {
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, inner := range typed {
			if isSensitiveKey(key) {
				result[key] = "[redacted]"
				continue
			}
			if strings.EqualFold(key, "comments") {
				result[key] = compactComments(inner, policy)
				continue
			}
			result[key] = sanitizeValue(inner, policy)
		}
		return result
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, sanitizeValue(item, policy))
		}
		return result
	default:
		return typed
	}
}

func compactComments(value any, policy ContextPolicy) any {
	comments, ok := value.([]any)
	if !ok {
		return sanitizeValue(value, policy)
	}
	limit := policy.MaxCommentsForLLM
	if limit > len(comments) {
		limit = len(comments)
	}
	compacted := make([]any, 0, limit)
	for i := 0; i < limit; i++ {
		item := sanitizeValue(comments[i], policy)
		if commentMap, ok := item.(map[string]any); ok {
			if content, ok := commentMap["content"].(string); ok {
				commentMap["content"] = truncateText(content, policy.MaxCommentCharsEach)
			}
		}
		compacted = append(compacted, item)
	}
	if len(comments) > limit {
		compacted = append(compacted, map[string]any{
			"omitted_comments": len(comments) - limit,
			"reason":           "truncated for llm context",
		})
	}
	return compacted
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(key)
	if normalized == "stack_trace" || normalized == "internal_stack_trace" {
		return true
	}
	for _, marker := range []string{"api_key", "authorization", "password", "secret", "token"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func redactText(value string) string {
	replacer := strings.NewReplacer("sk-", "[redacted]-")
	return replacer.Replace(value)
}

func truncateText(value string, maxChars int) string {
	if maxChars <= 0 {
		return value
	}
	if utf8.RuneCountInString(value) <= maxChars {
		return value
	}
	runes := []rune(value)
	if maxChars > len(runes) {
		maxChars = len(runes)
	}
	return string(runes[:maxChars]) + "...[truncated]"
}
