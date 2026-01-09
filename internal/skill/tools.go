package skill

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/tools"
)

// ConvertSkillToTool converts a Skill to a langgraphgo/langchaingo Tool
func ConvertSkillToTool(s *Skill, router *Router) tools.Tool {
	return &skillTool{
		skill:  s,
		router: router,
	}
}

// skillTool implements the tools.Tool interface
type skillTool struct {
	skill  *Skill
	router *Router
}

// Name returns the tool name
func (t *skillTool) Name() string {
	return t.skill.Name
}

// Description returns the tool description
func (t *skillTool) Description() string {
	return t.skill.Description
}

// Call executes the skill via the router
func (t *skillTool) Call(ctx context.Context, input string) (string, error) {
	// Parse input JSON to get parameters
	var params ExecutionParams
	if input != "" {
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			// If input is not JSON, treat it as a single parameter
			params = ExecutionParams{
				"input": input,
			}
		}
	} else {
		params = make(ExecutionParams)
	}

	// Execute via router
	result, err := t.router.Execute(t.skill.Name, params)
	if err != nil {
		return "", fmt.Errorf("skill execution failed: %w", err)
	}

	if !result.Success {
		return "", fmt.Errorf("skill execution failed: %s", result.Error)
	}

	return result.Output, nil
}

// GetTools converts all skills in the registry to tools
func (r *Router) GetTools() []tools.Tool {
	skills := r.registry.List()
	tools := make([]tools.Tool, 0, len(skills))

	for _, skill := range skills {
		tools = append(tools, ConvertSkillToTool(skill, r))
	}

	return tools
}

// GetPlanningTools returns tools suitable for planning phase
func (r *Router) GetPlanningTools() []tools.Tool {
	// For now, return all tools
	// In the future, this could filter tools based on metadata
	return r.GetTools()
}

// GetExecutionTools returns tools suitable for execution phase
func (r *Router) GetExecutionTools() []tools.Tool {
	// For now, return all tools
	// In the future, this could filter tools based on metadata
	return r.GetTools()
}

