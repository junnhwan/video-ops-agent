package runtime

import "video-ops-agent/internal/agent/llm"

const runtimeReminder = "Runtime rule: if tool evidence is sufficient, produce a final answer; otherwise request only the next necessary tool call."

func appendRuntimeReminder(messages []llm.Message) []llm.Message {
	if len(messages) == 0 {
		return []llm.Message{{Role: llm.RoleSystem, Content: runtimeReminder}}
	}
	messages[0].Content += "\n" + runtimeReminder
	return messages
}
