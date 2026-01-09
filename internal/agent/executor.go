package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/hb-chen/opskills/internal/skill"
	"github.com/hb-chen/opskills/internal/skill/direct"
	"github.com/hb-chen/opskills/internal/state"
)

// ExecutorAgent executes steps using skills
// It implements the graph.Executor interface
type ExecutorAgent struct {
	executor *direct.DirectExecutor
	registry *skill.Registry
}

// NewExecutorAgent creates a new executor agent
func NewExecutorAgent(executor *direct.DirectExecutor, registry *skill.Registry) *ExecutorAgent {
	return &ExecutorAgent{
		executor: executor,
		registry: registry,
	}
}

// Execute executes a single step
func (a *ExecutorAgent) Execute(ctx context.Context, step *state.Step) (*state.StepResult, error) {
	startTime := time.Now()

	// Get the skill
	skill, err := a.registry.Get(step.SkillName)
	if err != nil {
		return &state.StepResult{
			StepID:  step.ID,
			Success: false,
			Error:   fmt.Sprintf("skill not found: %s", step.SkillName),
		}, err
	}

	// Prepare execution parameters
	execParams := make(map[string]interface{})
	if step.Params != nil {
		// Convert map[string]interface{} to ExecutionParams
		for k, v := range step.Params {
			execParams[k] = v
		}
	}

	// Add action to params
	execParams["action"] = step.Action

	// Execute the skill
	// ExecutionParams is a type alias for map[string]interface{}, so we can pass execParams directly
	result, err := a.executor.Execute(skill, execParams)
	if err != nil {
		duration := time.Since(startTime)
		return &state.StepResult{
			StepID:   step.ID,
			Success:  false,
			Error:    result.Error,
			Duration: duration.String(),
		}, err
	}

	duration := time.Since(startTime)
	return &state.StepResult{
		StepID:   step.ID,
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
		Duration: duration.String(),
	}, nil
}

// ExecutePlan executes all steps in a plan
func (a *ExecutorAgent) ExecutePlan(ctx context.Context, plan *state.Plan) ([]*state.StepResult, error) {
	results := make([]*state.StepResult, 0, len(plan.Steps))

	for _, planStep := range plan.Steps {
		// Convert PlanStep to Step
		step := &state.Step{
			ID:          planStep.ID,
			SkillName:   planStep.SkillName,
			Action:      planStep.Action,
			Description: planStep.Description,
			Params:      planStep.Params,
			Status:      "pending",
		}

		// Execute step
		result, err := a.Execute(ctx, step)
		if err != nil {
			results = append(results, result)
			return results, fmt.Errorf("step %d failed: %w", step.ID, err)
		}

		results = append(results, result)

		// Check if we should continue
		if !result.Success {
			return results, fmt.Errorf("step %d failed", step.ID)
		}
	}

	return results, nil
}
