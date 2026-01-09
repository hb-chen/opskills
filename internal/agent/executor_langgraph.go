package agent

import (
	"context"

	"github.com/hb-chen/opskills/internal/llm"
	"github.com/hb-chen/opskills/internal/skill"
	"github.com/hb-chen/opskills/internal/state"
	// langgraphgo imports - will be uncommented once dependency is added
	// "github.com/smallnest/langgraphgo/prebuilt"
	// "github.com/tmc/langchaingo/llms"
)

// ExecutorAgentLangGraph is an executor agent using langgraphgo prebuilt.CreateAgent
// This will be fully implemented once langgraphgo dependency is added
type ExecutorAgentLangGraph struct {
	llmClient *llm.Client
	router    *skill.Router
	// agent     *prebuilt.Agent // Will be uncommented once langgraphgo is imported
}

// NewExecutorAgentLangGraph creates a new executor agent using langgraphgo
func NewExecutorAgentLangGraph(llmClient *llm.Client, router *skill.Router) (*ExecutorAgentLangGraph, error) {
	// Get execution tools from router
	// tools := router.GetExecutionTools()

	// Get LLM model
	// model := llmClient.GetModel()

	// Create agent using prebuilt.CreateAgent
	// This will be uncommented once langgraphgo is imported:
	/*
		agent, err := prebuilt.CreateAgent(model, tools,
			prebuilt.WithSystemMessage("You are an Ops executor agent. Execute tasks using available skills."),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create executor agent: %w", err)
		}
	*/

	return &ExecutorAgentLangGraph{
		llmClient: llmClient,
		router:    router,
		// agent:     agent,
	}, nil
}

// Execute executes a step using the agent
func (a *ExecutorAgentLangGraph) Execute(ctx context.Context, step *state.Step) (*state.StepResult, error) {
	// This will be implemented using the prebuilt agent once langgraphgo is imported:
	/*
		// Convert step to agent input
		stepDescription := fmt.Sprintf("Execute: %s with params: %v", step.Description, step.Params)
		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, stepDescription),
		}

		result, err := a.agent.Invoke(ctx, messages)
		if err != nil {
			return &state.StepResult{
				StepID:  step.ID,
				Success: false,
				Error:   err.Error(),
			}, err
		}

		// Parse result
		output := fmt.Sprintf("%v", result)
		return &state.StepResult{
			StepID:  step.ID,
			Success: true,
			Output:  output,
		}, nil
	*/

	// Temporary fallback to direct execution via router
	execParams := skill.ExecutionParams{}
	for k, v := range step.Params {
		execParams[k] = v
	}

	result, err := a.router.Execute(step.SkillName, execParams)
	if err != nil {
		return &state.StepResult{
			StepID:  step.ID,
			Success: false,
			Error:   err.Error(),
		}, err
	}

	return &state.StepResult{
		StepID:  step.ID,
		Success: result.Success,
		Output:  result.Output,
		Error:   result.Error,
	}, nil
}

