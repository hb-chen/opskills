package state

import (
	"github.com/tmc/langchaingo/llms"
)

// AgentState represents the agent execution state using langgraphgo State Schema
type AgentState struct {
	// Messages for LLM interaction (using langgraphgo graph tag)
	Messages []llms.MessageContent `graph:"messages" json:"messages,omitempty"`

	// Planning phase
	Plan      *Plan  `graph:"plan" json:"plan,omitempty"`
	PlanError string `graph:"plan_error" json:"plan_error,omitempty"`

	// Execution phase
	Steps       []*Step      `graph:"steps" json:"steps,omitempty"`
	CurrentStep int          `graph:"current_step" json:"current_step,omitempty"`
	Results     []*StepResult `graph:"results" json:"results,omitempty"`
	FinalResult *FinalResult `graph:"final_result" json:"final_result,omitempty"`

	// User query/request
	Query string `graph:"query" json:"query"`

	// Error handling
	Error string `graph:"error" json:"error,omitempty"`

	// Metadata
	TaskID    string `graph:"task_id" json:"task_id"`
	StartedAt string `graph:"started_at" json:"started_at,omitempty"`
	UpdatedAt string `graph:"updated_at" json:"updated_at,omitempty"`

	// Replanning support
	ReplanNeeded bool   `graph:"replan_needed" json:"replan_needed,omitempty"`
	ReplanReason  string `graph:"replan_reason" json:"replan_reason,omitempty"`
	ReplanCount   int    `graph:"replan_count" json:"replan_count,omitempty"`
}

// State represents the agent execution state (legacy, kept for compatibility)
type State struct {
	// User query/request
	Query string `json:"query"`

	// Planning phase
	Plan      *Plan      `json:"plan,omitempty"`
	PlanError string     `json:"plan_error,omitempty"`

	// Execution phase
	Steps        []*Step        `json:"steps,omitempty"`
	CurrentStep  int            `json:"current_step,omitempty"`
	Results      []*StepResult  `json:"results,omitempty"`
	FinalResult  *FinalResult   `json:"final_result,omitempty"`

	// Error handling
	Error string `json:"error,omitempty"`

	// Metadata
	TaskID    string `json:"task_id"`
	StartedAt string `json:"started_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Plan represents an execution plan
type Plan struct {
	Steps []*PlanStep `json:"steps"`
}

// PlanStep represents a single step in the plan
type PlanStep struct {
	ID          int    `json:"id"`
	SkillName   string `json:"skill_name"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Params      map[string]interface{} `json:"params,omitempty"`
}

// Step represents an execution step
type Step struct {
	ID          int    `json:"id"`
	SkillName   string `json:"skill_name"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Params      map[string]interface{} `json:"params,omitempty"`
	Status      string `json:"status"` // pending, running, completed, failed
}

// StepResult represents the result of executing a step
type StepResult struct {
	StepID   int    `json:"step_id"`
	Success  bool   `json:"success"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// FinalResult represents the final execution result
type FinalResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
	Summary string `json:"summary,omitempty"`
}

