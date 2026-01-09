package skill

import (
	"fmt"
	"path/filepath"

	"github.com/hb-chen/opskills/internal/skill/mcp"
)

// MCPAdapter adapts skills to MCP tools
type MCPAdapter struct {
	registry *Registry
}

// NewMCPAdapter creates a new MCP adapter
func NewMCPAdapter(registry *Registry) *MCPAdapter {
	return &MCPAdapter{
		registry: registry,
	}
}

// SkillToTool converts a skill to an MCP tool
func (a *MCPAdapter) SkillToTool(s *Skill) (*mcp.Tool, error) {
	schema := a.generateToolSchema(s)

	return &mcp.Tool{
		Name:        s.Name,
		Description: s.Description,
		InputSchema: schema,
	}, nil
}

// SkillsToTools converts multiple skills to MCP tools
func (a *MCPAdapter) SkillsToTools(skills []*Skill) ([]mcp.Tool, error) {
	tools := make([]mcp.Tool, 0, len(skills))

	for _, s := range skills {
		tool, err := a.SkillToTool(s)
		if err != nil {
			return nil, fmt.Errorf("failed to convert skill %s: %w", s.Name, err)
		}
		tools = append(tools, *tool)
	}

	return tools, nil
}

// generateToolSchema generates a JSON schema for a skill tool
func (a *MCPAdapter) generateToolSchema(s *Skill) map[string]interface{} {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The action to perform (e.g., create_cluster, add_nodes, etc.)",
			},
			"params": map[string]interface{}{
				"type":        "object",
				"description": "Additional parameters for the action",
				"properties":  map[string]interface{}{},
			},
		},
		"required": []string{"action"},
	}

	// Try to extract parameter information from skill instructions
	// This is a simplified implementation - in production, you might parse
	// the SKILL.md more intelligently to extract parameter schemas

	return schema
}

// ToolCallToSkillParams converts an MCP tool call to skill execution parameters
func (a *MCPAdapter) ToolCallToSkillParams(toolCall mcp.ToolCallParams) (ExecutionParams, error) {
	params := make(ExecutionParams)

	// Extract action
	if action, ok := toolCall.Arguments["action"].(string); ok {
		params["action"] = action
	}

	// Extract params object
	if paramsObj, ok := toolCall.Arguments["params"].(map[string]interface{}); ok {
		for k, v := range paramsObj {
			params[k] = v
		}
	}

	// Copy all other arguments
	for k, v := range toolCall.Arguments {
		if k != "action" && k != "params" {
			params[k] = v
		}
	}

	return params, nil
}

// SkillResultToToolResult converts a skill execution result to an MCP tool call result
func (a *MCPAdapter) SkillResultToToolResult(result *ExecutionResult) (*mcp.ToolCallResult, error) {
	content := []mcp.Content{}

	if result.Success {
		content = append(content, mcp.Content{
			Type: "text",
			Text: result.Output,
		})
	} else {
		// Include error information
		errorData := map[string]interface{}{
			"error":     result.Error,
			"exit_code": result.ExitCode,
		}
		content = append(content, mcp.Content{
			Type: "text",
			Text: fmt.Sprintf("Error: %s", result.Error),
			Data: errorData,
		})
	}

	return &mcp.ToolCallResult{
		Content: content,
		IsError: !result.Success,
	}, nil
}

// SkillToResource converts a skill to an MCP resource
func SkillToResource(s *Skill) *mcp.Resource {
	return &mcp.Resource{
		URI:         fmt.Sprintf("skill://%s", s.Name),
		Name:        s.Name,
		Description: s.Description,
		MimeType:    "text/markdown",
	}
}

// SkillsToResources converts multiple skills to MCP resources
func SkillsToResources(skills []*Skill) []mcp.Resource {
	resources := make([]mcp.Resource, 0, len(skills))

	for _, s := range skills {
		resources = append(resources, *SkillToResource(s))
	}

	return resources
}

// SkillConfigToResource converts a skill configuration file to an MCP resource
func SkillConfigToResource(skillName, configPath string) *mcp.Resource {
	return &mcp.Resource{
		URI:         fmt.Sprintf("skill://%s/config/%s", skillName, filepath.Base(configPath)),
		Name:        fmt.Sprintf("%s Configuration", skillName),
		Description: fmt.Sprintf("Configuration file for %s skill", skillName),
		MimeType:    "application/yaml",
	}
}

// SkillScriptToResource converts a skill script to an MCP resource
func SkillScriptToResource(skillName, scriptPath string) *mcp.Resource {
	return &mcp.Resource{
		URI:         fmt.Sprintf("skill://%s/script/%s", skillName, filepath.Base(scriptPath)),
		Name:        fmt.Sprintf("%s Script: %s", skillName, filepath.Base(scriptPath)),
		Description: fmt.Sprintf("Script file for %s skill", skillName),
		MimeType:    "text/x-shellscript",
	}
}
