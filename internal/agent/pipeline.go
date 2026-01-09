package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/hb-chen/opskills/internal/graph"
	"github.com/hb-chen/opskills/internal/state"
)

// Pipeline represents the agent execution pipeline
type Pipeline struct {
	graph   *graph.Graph
	planner *PlanningAgent
	executor *ExecutorAgent
}

// NewPipeline creates a new pipeline
func NewPipeline(planner *PlanningAgent, executor *ExecutorAgent) *Pipeline {
	// Build graph
	g := graph.BuildPlanningExecutionGraph(planner, executor)

	return &Pipeline{
		graph:    g,
		planner:  planner,
		executor: executor,
	}
}

// Execute executes a task through the pipeline
func (p *Pipeline) Execute(ctx context.Context, query string, taskID string) (*state.State, error) {
	// Initialize state
	initialState := &state.State{
		Query:     query,
		TaskID:    taskID,
		StartedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// Execute graph starting from planning node
	finalState, err := p.graph.Execute(ctx, "planning", initialState)
	if err != nil {
		finalState.Error = err.Error()
		finalState.UpdatedAt = time.Now().Format(time.RFC3339)
		return finalState, err
	}

	finalState.UpdatedAt = time.Now().Format(time.RFC3339)
	return finalState, nil
}

// GetState returns the current state (for status queries)
func (p *Pipeline) GetState(taskID string) (*state.State, error) {
	// In a real implementation, this would retrieve state from storage
	// For now, return an error indicating state needs to be stored
	return nil, fmt.Errorf("state storage not implemented yet")
}



