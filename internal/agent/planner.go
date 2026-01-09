package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hb-chen/opskills/internal/llm"
	"github.com/hb-chen/opskills/internal/skill"
	"github.com/hb-chen/opskills/internal/state"
)

// PlanningAgent generates execution plans using LLM
type PlanningAgent struct {
	llmClient *llm.Client
	registry  *skill.Registry
}

// NewPlanningAgent creates a new planning agent
func NewPlanningAgent(llmClient *llm.Client, registry *skill.Registry) *PlanningAgent {
	return &PlanningAgent{
		llmClient: llmClient,
		registry:  registry,
	}
}

// Plan generates an execution plan from a user query
func (a *PlanningAgent) Plan(ctx context.Context, query string) (*state.Plan, error) {
	// Get available skills
	skills := a.registry.List()
	if len(skills) == 0 {
		return nil, fmt.Errorf("no skills available")
	}

	// Prepare skill information for prompt
	skillInfos := make([]llm.SkillInfo, len(skills))
	for i, s := range skills {
		skillInfos[i] = llm.SkillInfo{
			Name:        s.Name,
			Description: s.Description,
		}
	}

	// Format planning prompt
	promptData := llm.PlanningPromptData{
		Skills: skillInfos,
		Query:  query,
	}
	prompt := llm.FormatPlanningPrompt(promptData)

	// Generate plan using LLM
	response, err := a.llmClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// Parse JSON response
	plan, err := parsePlanResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan response: %w", err)
	}

	return plan, nil
}

// parsePlanResponse parses the LLM response into a Plan
func parsePlanResponse(response string) (*state.Plan, error) {
	// Try to extract JSON from response (LLM might add extra text)
	// Look for JSON object in the response
	startIdx := -1
	endIdx := -1
	braceCount := 0

	for i, char := range response {
		if char == '{' {
			if startIdx == -1 {
				startIdx = i
			}
			braceCount++
		} else if char == '}' {
			braceCount--
			if braceCount == 0 && startIdx != -1 {
				endIdx = i + 1
				break
			}
		}
	}

	if startIdx == -1 || endIdx == -1 {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := response[startIdx:endIdx]

	// Parse JSON
	var planData struct {
		Steps []struct {
			ID          int                    `json:"id"`
			SkillName   string                 `json:"skill_name"`
			Action      string                 `json:"action"`
			Description string                 `json:"description"`
			Params      map[string]interface{} `json:"params"`
		} `json:"steps"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &planData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to Plan
	plan := &state.Plan{
		Steps: make([]*state.PlanStep, len(planData.Steps)),
	}

	for i, step := range planData.Steps {
		plan.Steps[i] = &state.PlanStep{
			ID:          step.ID,
			SkillName:   step.SkillName,
			Action:      step.Action,
			Description: step.Description,
			Params:      step.Params,
		}
	}

	return plan, nil
}



