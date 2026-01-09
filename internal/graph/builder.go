package graph

import (
	"context"
	"fmt"

	"github.com/hb-chen/opskills/internal/llm"
	"github.com/hb-chen/opskills/internal/skill"
	"github.com/hb-chen/opskills/internal/state"
	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/store"
	"github.com/smallnest/langgraphgo/store/file"
	"github.com/tmc/langchaingo/llms"
)

// LangGraphNodeFunc represents a graph node function for langgraphgo
type LangGraphNodeFunc func(ctx context.Context, state map[string]any) (map[string]any, error)

// OpsGraphBuilder builds an Ops execution graph using langgraphgo
type OpsGraphBuilder struct {
	skillRouter *skill.Router
	llmClient   *llm.Client
}

// NewOpsGraphBuilder creates a new graph builder
func NewOpsGraphBuilder(skillRouter *skill.Router, llmClient *llm.Client) *OpsGraphBuilder {
	return &OpsGraphBuilder{
		skillRouter: skillRouter,
		llmClient:   llmClient,
	}
}

// Build creates a new StateGraph using langgraphgo
func (b *OpsGraphBuilder) Build() (*graph.StateGraph[map[string]any], error) {
	// Create state graph
	g := graph.NewStateGraph[map[string]any]()

	// Set up schema with reducers
	schema := graph.NewMapSchema()
	schema.RegisterReducer("messages", graph.AddMessages)
	schema.RegisterReducer("results", graph.AppendReducer)
	g.SetSchema(schema)

	// Add nodes
	g.AddNode("planning", "Planning node: generates execution plan", b.createPlanningNode())
	g.AddNode("execution", "Execution node: executes plan steps", b.createExecutionNode())
	g.AddNode("validation", "Validation node: validates execution results", b.createValidationNode())

	// Define edges
	g.AddEdge("planning", "execution")
	g.AddEdge("execution", "validation")
	g.AddEdge("validation", graph.END)
	g.SetEntryPoint("planning")

	return g, nil
}

// BuildWithCheckpointer creates a graph with checkpointer support
func (b *OpsGraphBuilder) BuildWithCheckpointer(storeType string, config map[string]interface{}) (*graph.StateGraph[map[string]any], error) {
	g, err := b.Build()
	if err != nil {
		return nil, err
	}

	// Create checkpointer based on store type
	var checkpointStore store.CheckpointStore
	switch storeType {
	case "file":
		path, _ := config["path"].(string)
		if path == "" {
			path = "./checkpoints"
		}
		checkpointStore, err = file.NewFileCheckpointStore(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create file checkpoint store: %w", err)
		}
	case "redis":
		// TODO: Implement Redis checkpointer when needed
		return nil, fmt.Errorf("Redis checkpointer not yet implemented")
	case "postgres":
		// TODO: Implement Postgres checkpointer when needed
		return nil, fmt.Errorf("Postgres checkpointer not yet implemented")
	default:
		return nil, fmt.Errorf("unknown store type: %s", storeType)
	}

	// Note: The checkpoint store is created but not yet integrated
	// To use it, convert g to CheckpointableStateGraph and use CompileCheckpointable
	_ = checkpointStore // Use checkpointStore when implementing full checkpoint support

	return g, nil
}

// createPlanningNode creates the planning node function
func (b *OpsGraphBuilder) createPlanningNode() LangGraphNodeFunc {
	return func(ctx context.Context, stateMap map[string]any) (map[string]any, error) {
		// Convert map to AgentState for easier manipulation
		agentState := b.mapToAgentState(stateMap)

		// If plan already exists, skip
		if agentState.Plan != nil {
			return stateMap, nil
		}

		// Get query from messages or state
		query := agentState.Query
		if query == "" && len(agentState.Messages) > 0 {
			// Extract query from last human message
			for i := len(agentState.Messages) - 1; i >= 0; i-- {
				msg := agentState.Messages[i]
				if msg.Role == llms.ChatMessageTypeHuman {
					if parts := msg.Parts; len(parts) > 0 {
						for _, part := range parts {
							if textPart, ok := part.(llms.TextContent); ok {
								query = textPart.Text
								break
							}
						}
						if query != "" {
							break
						}
					}
				}
			}
		}

		if query == "" {
			return stateMap, fmt.Errorf("no query provided")
		}

		// Use planning agent to generate plan
		plan, err := b.generatePlan(ctx, query)
		if err != nil {
			agentState.PlanError = err.Error()
			agentState.Error = fmt.Sprintf("planning failed: %v", err)
			return b.agentStateToMap(agentState), err
		}

		agentState.Plan = plan
		agentState.Query = query

		// Convert plan steps to execution steps
		agentState.Steps = make([]*state.Step, len(plan.Steps))
		for i, ps := range plan.Steps {
			agentState.Steps[i] = &state.Step{
				ID:          ps.ID,
				SkillName:   ps.SkillName,
				Action:      ps.Action,
				Description: ps.Description,
				Params:      ps.Params,
				Status:      "pending",
			}
		}

		return b.agentStateToMap(agentState), nil
	}
}

// createExecutionNode creates the execution node function
func (b *OpsGraphBuilder) createExecutionNode() LangGraphNodeFunc {
	return func(ctx context.Context, stateMap map[string]any) (map[string]any, error) {
		// Convert map to AgentState
		agentState := b.mapToAgentState(stateMap)

		if agentState.Plan == nil || len(agentState.Steps) == 0 {
			return stateMap, fmt.Errorf("no plan or steps to execute")
		}

		// Initialize results if needed
		if agentState.Results == nil {
			agentState.Results = make([]*state.StepResult, 0)
		}

		// Execute pending steps
		for i := agentState.CurrentStep; i < len(agentState.Steps); i++ {
			step := agentState.Steps[i]
			if step.Status != "pending" {
				continue
			}

			step.Status = "running"
			agentState.CurrentStep = i

			// Execute the step using skill router
			execParams := make(skill.ExecutionParams)
			for k, v := range step.Params {
				execParams[k] = v
			}

			result, err := b.skillRouter.Execute(step.SkillName, execParams)
			if err != nil {
				step.Status = "failed"
				agentState.Results = append(agentState.Results, &state.StepResult{
					StepID:  step.ID,
					Success: false,
					Error:   err.Error(),
				})
				agentState.Error = fmt.Sprintf("step %d failed: %v", step.ID, err)
				return b.agentStateToMap(agentState), err
			}

			step.Status = "completed"
			output := ""
			errorMsg := ""
			if result != nil {
				output = result.Output
				errorMsg = result.Error
			}
			agentState.Results = append(agentState.Results, &state.StepResult{
				StepID:  step.ID,
				Success: result != nil && result.Success,
				Output:  output,
				Error:   errorMsg,
			})
		}

		return b.agentStateToMap(agentState), nil
	}
}

// createValidationNode creates the validation node function
func (b *OpsGraphBuilder) createValidationNode() LangGraphNodeFunc {
	return func(ctx context.Context, stateMap map[string]any) (map[string]any, error) {
		// Convert map to AgentState
		agentState := b.mapToAgentState(stateMap)

		// Check if all steps are completed
		allCompleted := true
		for _, step := range agentState.Steps {
			if step.Status != "completed" {
				allCompleted = false
				break
			}
		}

		if allCompleted && len(agentState.Steps) > 0 {
			// Generate final result
			agentState.FinalResult = &state.FinalResult{
				Success: true,
				Summary: fmt.Sprintf("Completed %d steps successfully", len(agentState.Steps)),
			}
		}

		return b.agentStateToMap(agentState), nil
	}
}

// generatePlan generates a plan using LLM (temporary, will be replaced with prebuilt agent)
func (b *OpsGraphBuilder) generatePlan(ctx context.Context, query string) (*state.Plan, error) {
	// This is a temporary implementation
	// Will be replaced with prebuilt.CreateReactAgent in the next step
	return &state.Plan{
		Steps: []*state.PlanStep{
			{
				ID:          1,
				SkillName:   "kubekey",
				Action:      "create",
				Description: query,
				Params:      make(map[string]interface{}),
			},
		},
	}, nil
}

// mapToAgentState converts a map[string]any to AgentState
func (b *OpsGraphBuilder) mapToAgentState(stateMap map[string]any) *state.AgentState {
	agentState := &state.AgentState{
		Query:       getString(stateMap, "query"),
		TaskID:      getString(stateMap, "task_id"),
		StartedAt:   getString(stateMap, "started_at"),
		UpdatedAt:   getString(stateMap, "updated_at"),
		Error:       getString(stateMap, "error"),
		PlanError:   getString(stateMap, "plan_error"),
		CurrentStep: getInt(stateMap, "current_step"),
	}

	// Convert plan
	if planVal, ok := stateMap["plan"]; ok {
		if planMap, ok := planVal.(map[string]any); ok {
			agentState.Plan = b.mapToPlan(planMap)
		}
	}

	// Convert steps
	if stepsVal, ok := stateMap["steps"]; ok {
		if stepsSlice, ok := stepsVal.([]any); ok {
			agentState.Steps = b.mapToSteps(stepsSlice)
		}
	}

	// Convert results
	if resultsVal, ok := stateMap["results"]; ok {
		if resultsSlice, ok := resultsVal.([]any); ok {
			agentState.Results = b.mapToStepResults(resultsSlice)
		}
	}

	// Convert final result
	if finalVal, ok := stateMap["final_result"]; ok {
		if finalMap, ok := finalVal.(map[string]any); ok {
			agentState.FinalResult = b.mapToFinalResult(finalMap)
		}
	}

	// Messages are handled by langgraphgo's AddMessages reducer
	// They are kept in the map and managed by the reducer

	return agentState
}

// agentStateToMap converts AgentState to map[string]any
func (b *OpsGraphBuilder) agentStateToMap(agentState *state.AgentState) map[string]any {
	stateMap := make(map[string]any)

	stateMap["query"] = agentState.Query
	stateMap["task_id"] = agentState.TaskID
	stateMap["started_at"] = agentState.StartedAt
	stateMap["updated_at"] = agentState.UpdatedAt
	stateMap["error"] = agentState.Error
	stateMap["plan_error"] = agentState.PlanError
	stateMap["current_step"] = agentState.CurrentStep

	if agentState.Plan != nil {
		stateMap["plan"] = b.planToMap(agentState.Plan)
	}

	if agentState.Steps != nil {
		stateMap["steps"] = b.stepsToMap(agentState.Steps)
	}

	if agentState.Results != nil {
		stateMap["results"] = b.stepResultsToMap(agentState.Results)
	}

	if agentState.FinalResult != nil {
		stateMap["final_result"] = b.finalResultToMap(agentState.FinalResult)
	}

	// Messages are handled separately by langgraphgo

	return stateMap
}

// Helper functions for type conversion
func getString(m map[string]any, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getInt(m map[string]any, key string) int {
	if val, ok := m[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return 0
}

// Conversion helpers (simplified - full implementation would handle all fields)
func (b *OpsGraphBuilder) mapToPlan(planMap map[string]any) *state.Plan {
	plan := &state.Plan{}
	if stepsVal, ok := planMap["steps"]; ok {
		if stepsSlice, ok := stepsVal.([]any); ok {
			plan.Steps = make([]*state.PlanStep, len(stepsSlice))
			for i, stepVal := range stepsSlice {
				if stepMap, ok := stepVal.(map[string]any); ok {
					plan.Steps[i] = &state.PlanStep{
						ID:          getInt(stepMap, "id"),
						SkillName:   getString(stepMap, "skill_name"),
						Action:      getString(stepMap, "action"),
						Description: getString(stepMap, "description"),
						Params:      getMap(stepMap, "params"),
					}
				}
			}
		}
	}
	return plan
}

func (b *OpsGraphBuilder) mapToSteps(stepsSlice []any) []*state.Step {
	steps := make([]*state.Step, len(stepsSlice))
	for i, stepVal := range stepsSlice {
		if stepMap, ok := stepVal.(map[string]any); ok {
			steps[i] = &state.Step{
				ID:          getInt(stepMap, "id"),
				SkillName:   getString(stepMap, "skill_name"),
				Action:      getString(stepMap, "action"),
				Description: getString(stepMap, "description"),
				Params:      getMap(stepMap, "params"),
				Status:      getString(stepMap, "status"),
			}
		}
	}
	return steps
}

func (b *OpsGraphBuilder) mapToStepResults(resultsSlice []any) []*state.StepResult {
	results := make([]*state.StepResult, len(resultsSlice))
	for i, resultVal := range resultsSlice {
		if resultMap, ok := resultVal.(map[string]any); ok {
			results[i] = &state.StepResult{
				StepID:  getInt(resultMap, "step_id"),
				Success: getBool(resultMap, "success"),
				Output:  getString(resultMap, "output"),
				Error:   getString(resultMap, "error"),
			}
		}
	}
	return results
}

func (b *OpsGraphBuilder) mapToFinalResult(finalMap map[string]any) *state.FinalResult {
	return &state.FinalResult{
		Success: getBool(finalMap, "success"),
		Output:  getString(finalMap, "output"),
		Error:   getString(finalMap, "error"),
		Summary: getString(finalMap, "summary"),
	}
}

func getBool(m map[string]any, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func getMap(m map[string]any, key string) map[string]interface{} {
	if val, ok := m[key]; ok {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
		if mapVal, ok := val.(map[string]any); ok {
			result := make(map[string]interface{})
			for k, v := range mapVal {
				result[k] = v
			}
			return result
		}
	}
	return make(map[string]interface{})
}

func (b *OpsGraphBuilder) planToMap(plan *state.Plan) map[string]any {
	result := make(map[string]any)
	if plan.Steps != nil {
		steps := make([]any, len(plan.Steps))
		for i, step := range plan.Steps {
			steps[i] = map[string]any{
				"id":          step.ID,
				"skill_name":  step.SkillName,
				"action":      step.Action,
				"description": step.Description,
				"params":      step.Params,
			}
		}
		result["steps"] = steps
	}
	return result
}

func (b *OpsGraphBuilder) stepsToMap(steps []*state.Step) []any {
	result := make([]any, len(steps))
	for i, step := range steps {
		result[i] = map[string]any{
			"id":          step.ID,
			"skill_name":  step.SkillName,
			"action":      step.Action,
			"description": step.Description,
			"params":      step.Params,
			"status":      step.Status,
		}
	}
	return result
}

func (b *OpsGraphBuilder) stepResultsToMap(results []*state.StepResult) []any {
	result := make([]any, len(results))
	for i, res := range results {
		result[i] = map[string]any{
			"step_id": res.StepID,
			"success": res.Success,
			"output":  res.Output,
			"error":   res.Error,
		}
	}
	return result
}

func (b *OpsGraphBuilder) finalResultToMap(final *state.FinalResult) map[string]any {
	return map[string]any{
		"success": final.Success,
		"output":  final.Output,
		"error":   final.Error,
		"summary": final.Summary,
	}
}

// Legacy Graph type for backward compatibility
// Note: This uses the old NodeFunc signature from nodes.go
type LegacyNodeFunc func(ctx context.Context, s *state.State) (*state.State, error)

type Graph struct {
	nodes map[string]LegacyNodeFunc
	edges map[string][]string
}

// NewGraph creates a new legacy graph (for backward compatibility)
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]LegacyNodeFunc),
		edges: make(map[string][]string),
	}
}

// AddNode adds a node to the legacy graph
func (g *Graph) AddNode(name string, fn LegacyNodeFunc) {
	g.nodes[name] = fn
}

// AddEdge adds an edge between nodes
func (g *Graph) AddEdge(from, to string) {
	g.edges[from] = append(g.edges[from], to)
}

// Execute executes the legacy graph
func (g *Graph) Execute(ctx context.Context, startNode string, initialState *state.State) (*state.State, error) {
	currentState := initialState
	currentNode := startNode
	visited := make(map[string]bool)

	for {
		if visited[currentNode] {
			return currentState, fmt.Errorf("circular dependency detected at node: %s", currentNode)
		}
		visited[currentNode] = true

		nodeFn, exists := g.nodes[currentNode]
		if !exists {
			return currentState, fmt.Errorf("node not found: %s", currentNode)
		}

		var err error
		currentState, err = nodeFn(ctx, currentState)
		if err != nil {
			return currentState, fmt.Errorf("node %s failed: %w", currentNode, err)
		}

		if currentState.FinalResult != nil {
			break
		}

		nextNodes := g.edges[currentNode]
		if len(nextNodes) == 0 {
			break
		}

		currentNode = nextNodes[0]
	}

	return currentState, nil
}

// BuildPlanningExecutionGraph builds a legacy planning -> execution graph
func BuildPlanningExecutionGraph(planner Planner, executor Executor) *Graph {
	g := NewGraph()
	// Convert NodeFunc from nodes.go to LegacyNodeFunc
	planningFn := PlanningNode(planner)
	executionFn := ExecutionNode(executor)

	// Wrap the functions to match LegacyNodeFunc signature
	g.AddNode("planning", func(ctx context.Context, s *state.State) (*state.State, error) {
		return planningFn(ctx, s)
	})
	g.AddNode("execution", func(ctx context.Context, s *state.State) (*state.State, error) {
		return executionFn(ctx, s)
	})
	g.AddEdge("planning", "execution")
	return g
}
