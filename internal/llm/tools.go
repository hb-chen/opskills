package llm

import (
	"encoding/json"
	"fmt"

	"github.com/hb-chen/opskills/internal/skill"
	"github.com/tmc/langchaingo/llms"
)

// ConvertSkillsToTools converts skills to LLM tools
func ConvertSkillsToTools(skills []*skill.Skill) []llms.Tool {
	tools := make([]llms.Tool, 0, len(skills))

	for _, s := range skills {
		tool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        s.Name,
				Description: s.Description,
				Parameters:  generateInputSchema(s),
			},
		}
		tools = append(tools, tool)
	}

	return tools
}

// generateInputSchema generates a JSON schema for skill input
func generateInputSchema(s *skill.Skill) map[string]interface{} {
	// Basic schema structure
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The action to perform (e.g., create_cluster, add_nodes)",
			},
			"params": map[string]interface{}{
				"type":        "object",
				"description": "Additional parameters for the action",
			},
		},
		"required": []string{"action"},
	}

	return schema
}

// ParseToolCall parses a tool call and extracts skill execution parameters
func ParseToolCall(toolCall llms.ToolCall) (skillName string, params skill.ExecutionParams, err error) {
	if toolCall.FunctionCall == nil {
		return "", nil, fmt.Errorf("tool call has no function call")
	}

	skillName = toolCall.FunctionCall.Name

	// Parse arguments
	if toolCall.FunctionCall.Arguments != "" {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.FunctionCall.Arguments), &args); err != nil {
			return "", nil, fmt.Errorf("failed to parse tool arguments: %w", err)
		}

		params = make(skill.ExecutionParams)
		for k, v := range args {
			params[k] = v
		}
	} else {
		params = make(skill.ExecutionParams)
	}

	return skillName, params, nil
}

// FormatToolResult formats a skill execution result for LLM consumption
func FormatToolResult(result *skill.ExecutionResult) string {
	if result.Success {
		return fmt.Sprintf("Success: %s", result.Output)
	}
	return fmt.Sprintf("Error: %s", result.Error)
}



