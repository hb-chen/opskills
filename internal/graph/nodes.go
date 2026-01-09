package graph

import (
	"context"
	"fmt"

	"github.com/hb-chen/opskills/internal/state"
)

// NodeFunc represents a graph node function
type NodeFunc func(ctx context.Context, s *state.State) (*state.State, error)

// PlanningNode handles the planning phase
func PlanningNode(planner Planner) NodeFunc {
	return func(ctx context.Context, s *state.State) (*state.State, error) {
		if s.Plan != nil {
			// Plan already exists, skip
			return s, nil
		}

		plan, err := planner.Plan(ctx, s.Query)
		if err != nil {
			s.PlanError = err.Error()
			s.Error = fmt.Sprintf("planning failed: %v", err)
			return s, err
		}

		s.Plan = plan
		
		// Convert plan steps to execution steps
		s.Steps = make([]*state.Step, len(plan.Steps))
		for i, ps := range plan.Steps {
			s.Steps[i] = &state.Step{
				ID:          ps.ID,
				SkillName:   ps.SkillName,
				Action:      ps.Action,
				Description: ps.Description,
				Params:      ps.Params,
				Status:      "pending",
			}
		}

		return s, nil
	}
}

// ExecutionNode handles the execution phase
func ExecutionNode(executor Executor) NodeFunc {
	return func(ctx context.Context, s *state.State) (*state.State, error) {
		if s.Plan == nil || len(s.Steps) == 0 {
			return s, fmt.Errorf("no plan or steps to execute")
		}

		// Initialize results if needed
		if s.Results == nil {
			s.Results = make([]*state.StepResult, 0)
		}

		// Execute pending steps
		for i := s.CurrentStep; i < len(s.Steps); i++ {
			step := s.Steps[i]
			if step.Status != "pending" {
				continue
			}

			step.Status = "running"
			s.CurrentStep = i

			// Execute the step
			result, err := executor.Execute(ctx, step)
			if err != nil {
				step.Status = "failed"
				s.Results = append(s.Results, &state.StepResult{
					StepID:  step.ID,
					Success: false,
					Error:   err.Error(),
				})
				s.Error = fmt.Sprintf("step %d failed: %v", step.ID, err)
				return s, err
			}

			step.Status = "completed"
			s.Results = append(s.Results, result)
		}

		// Check if all steps are completed
		allCompleted := true
		for _, step := range s.Steps {
			if step.Status != "completed" {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			// Generate final result
			s.FinalResult = &state.FinalResult{
				Success: true,
				Summary: fmt.Sprintf("Completed %d steps successfully", len(s.Steps)),
			}
		}

		return s, nil
	}
}

// Planner interface for planning
type Planner interface {
	Plan(ctx context.Context, query string) (*state.Plan, error)
}

// Executor interface for execution
type Executor interface {
	Execute(ctx context.Context, step *state.Step) (*state.StepResult, error)
}



