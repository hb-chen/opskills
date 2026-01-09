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

// PlanningAgentLangGraph is a planning agent using langgraphgo prebuilt.CreateReactAgent
// This will be fully implemented once langgraphgo dependency is added
type PlanningAgentLangGraph struct {
	llmClient *llm.Client
	router    *skill.Router
	// agent     *prebuilt.Agent // Will be uncommented once langgraphgo is imported
}

// NewPlanningAgentLangGraph creates a new planning agent using langgraphgo
func NewPlanningAgentLangGraph(llmClient *llm.Client, router *skill.Router) (*PlanningAgentLangGraph, error) {
	// Get planning tools from router
	// tools := router.GetPlanningTools()

	// Get LLM model
	// model := llmClient.GetModel()

	// Create ReAct agent using prebuilt.CreateReactAgent
	// This will be uncommented once langgraphgo is imported:
	/*
		agent, err := prebuilt.CreateReactAgent(model, tools,
			prebuilt.WithSystemMessage("You are a planning agent for Ops tasks. Generate execution plans based on user requirements."),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create ReAct agent: %w", err)
		}
	*/

	return &PlanningAgentLangGraph{
		llmClient: llmClient,
		router:    router,
		// agent:     agent,
	}, nil
}

// Plan generates an execution plan using the ReAct agent
func (a *PlanningAgentLangGraph) Plan(ctx context.Context, query string) (*state.Plan, error) {
	// This will be implemented using the prebuilt agent once langgraphgo is imported:
	/*
		messages := []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, query),
		}

		result, err := a.agent.Invoke(ctx, messages)
		if err != nil {
			return nil, fmt.Errorf("agent invocation failed: %w", err)
		}

		// Parse result to extract plan
		plan, err := parsePlanFromAgentResponse(result)
		return plan, err
	*/

	// Temporary fallback to existing planner
	planner := NewPlanningAgent(a.llmClient, a.router.GetRegistry())
	return planner.Plan(ctx, query)
}

// parsePlanFromAgentResponse parses the agent response into a Plan
// This will be implemented once langgraphgo is integrated
func parsePlanFromAgentResponse(response any) (*state.Plan, error) {
	// TODO: Implement parsing logic
	return nil, nil
}

