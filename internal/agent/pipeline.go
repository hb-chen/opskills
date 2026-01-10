package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/hb-chen/opskills/internal/graph"
	"github.com/hb-chen/opskills/internal/state"
	langgraph "github.com/smallnest/langgraphgo/graph"
)

// Pipeline represents the agent execution pipeline
type Pipeline struct {
	graph           *graph.Graph               // Legacy graph
	checkpointGraph *graph.CheckpointableGraph // Checkpoint-enabled graph
	planner         *PlanningAgent
	executor        *ExecutorAgent
	useCheckpoint   bool
}

// NewPipeline creates a new pipeline with legacy graph
func NewPipeline(planner *PlanningAgent, executor *ExecutorAgent) *Pipeline {
	// Build graph
	g := graph.BuildPlanningExecutionGraph(planner, executor)

	return &Pipeline{
		graph:         g,
		planner:       planner,
		executor:      executor,
		useCheckpoint: false,
	}
}

// NewPipelineWithCheckpoint creates a new pipeline with checkpoint support
func NewPipelineWithCheckpoint(checkpointGraph *graph.CheckpointableGraph, planner *PlanningAgent, executor *ExecutorAgent) *Pipeline {
	return &Pipeline{
		checkpointGraph: checkpointGraph,
		planner:         planner,
		executor:        executor,
		useCheckpoint:   true,
	}
}

// Execute executes a task through the pipeline
func (p *Pipeline) Execute(ctx context.Context, query string, taskID string) (*state.State, error) {
	if p.useCheckpoint && p.checkpointGraph != nil {
		return p.executeWithCheckpoint(ctx, query, taskID)
	}

	// Use legacy graph execution
	// Validate that legacy graph is initialized
	if p.graph == nil {
		return nil, fmt.Errorf("pipeline graph not initialized: both checkpoint and legacy graphs are nil")
	}

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

// executeWithCheckpoint executes using langgraphgo CheckpointableStateGraph with checkpoint support
func (p *Pipeline) executeWithCheckpoint(ctx context.Context, query string, taskID string) (*state.State, error) {
	// Compile checkpointable graph
	runnable, err := p.checkpointGraph.Graph.CompileCheckpointable()
	if err != nil {
		return nil, fmt.Errorf("failed to compile checkpointable graph: %w", err)
	}

	// Convert initial state to map format
	initialStateMap := map[string]any{
		"query":      query,
		"task_id":    taskID,
		"started_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	// Create config with thread_id (taskID) for checkpoint tracking
	// This allows checkpoint store to organize checkpoints by task
	config := &langgraph.Config{
		Configurable: map[string]any{
			"thread_id": taskID,
		},
	}

	// Execute graph with checkpoint support and replanning loop
	// Checkpoint will automatically save state at each node
	maxReplans := 3
	replanCount := 0
	currentState := initialStateMap

	for {
		resultMap, err := runnable.InvokeWithConfig(ctx, currentState, config)
		if err != nil {
			// Convert error state
			finalState := p.mapToState(resultMap, taskID)
			finalState.Error = err.Error()
			finalState.UpdatedAt = time.Now().Format(time.RFC3339)
			return finalState, err
		}

		// Check if replanning is needed
		replanNeeded := false
		if val, ok := resultMap["replan_needed"]; ok {
			if b, ok := val.(bool); ok {
				replanNeeded = b
			}
		}

		// Check if we have a final result (validation passed)
		// If we have a successful final result, we can exit the loop
		if val, ok := resultMap["final_result"]; ok {
			if finalMap, ok := val.(map[string]any); ok {
				if success, ok := finalMap["success"].(bool); ok && success {
					// Validation passed, exit loop
					finalState := p.mapToState(resultMap, taskID)
					finalState.UpdatedAt = time.Now().Format(time.RFC3339)
					return finalState, nil
				}
			}
		}

		// If replanning is needed and we haven't exceeded max replans
		if replanNeeded && replanCount < maxReplans {
			replanCount++
			// Update replan count in state
			resultMap["replan_count"] = replanCount
			// Clear plan to force regeneration
			delete(resultMap, "plan")
			delete(resultMap, "steps")
			delete(resultMap, "results")
			resultMap["current_step"] = 0
			resultMap["replan_needed"] = false // Reset flag
			// Continue loop with updated state
			currentState = resultMap
			continue
		}

		// Convert result map back to State
		finalState := p.mapToState(resultMap, taskID)
		finalState.UpdatedAt = time.Now().Format(time.RFC3339)
		return finalState, nil
	}
}

// mapToState converts a map[string]any to State (simplified conversion)
func (p *Pipeline) mapToState(stateMap map[string]any, taskID string) *state.State {
	s := &state.State{
		TaskID: taskID,
	}

	if query, ok := stateMap["query"].(string); ok {
		s.Query = query
	}
	if startedAt, ok := stateMap["started_at"].(string); ok {
		s.StartedAt = startedAt
	}
	if updatedAt, ok := stateMap["updated_at"].(string); ok {
		s.UpdatedAt = updatedAt
	}
	if err, ok := stateMap["error"].(string); ok {
		s.Error = err
	}

	// Note: Full conversion would require converting plan, steps, results, etc.
	// This is a simplified version - full implementation would use the builder's conversion functions

	return s
}

// GetState returns the current state (for status queries)
func (p *Pipeline) GetState(taskID string) (*state.State, error) {
	// In a real implementation, this would retrieve state from storage
	// For now, return an error indicating state needs to be stored
	return nil, fmt.Errorf("state storage not implemented yet")
}
