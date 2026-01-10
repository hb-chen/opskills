package tracer

import (
	"context"
	"fmt"
	"time"

	"github.com/hb-chen/opskills/internal/state"
	"github.com/hb-chen/opskills/pkg/logger"
)

// LogTracer implements ExecutionTracer using structured logging
type LogTracer struct {
	level string // minimal, standard, detailed
}

// NewLogTracer creates a new log tracer
func NewLogTracer(level string) *LogTracer {
	return &LogTracer{level: level}
}

func (l *LogTracer) TraceNodeStart(ctx context.Context, nodeName, taskID string) error {
	if l.level == "minimal" {
		return nil
	}
	logger.Infof("[Tracer] Node started: task=%s, node=%s", taskID, nodeName)
	return nil
}

func (l *LogTracer) TraceNodeEnd(ctx context.Context, nodeName, taskID string, duration time.Duration) error {
	if l.level == "minimal" {
		return nil
	}
	logger.Infof("[Tracer] Node completed: task=%s, node=%s, duration=%v", taskID, nodeName, duration)
	return nil
}

func (l *LogTracer) TraceLLMRequest(ctx context.Context, taskID, prompt string) error {
	if l.level != "detailed" {
		return nil
	}
	// Truncate long prompts for readability
	truncatedPrompt := prompt
	if len(prompt) > 200 {
		truncatedPrompt = prompt[:200] + "..."
	}
	logger.Debugf("[Tracer] LLM request: task=%s, prompt=%s", taskID, truncatedPrompt)
	return nil
}

func (l *LogTracer) TraceLLMResponse(ctx context.Context, taskID, response string, duration time.Duration) error {
	if l.level != "detailed" {
		return nil
	}
	// Truncate long responses for readability
	truncatedResponse := response
	if len(response) > 200 {
		truncatedResponse = response[:200] + "..."
	}
	logger.Debugf("[Tracer] LLM response: task=%s, duration=%v, response=%s", taskID, duration, truncatedResponse)
	return nil
}

func (l *LogTracer) TraceStepStart(ctx context.Context, taskID string, step *state.Step) error {
	if l.level == "minimal" {
		return nil
	}
	logger.Infof("[Tracer] Step started: task=%s, step=%d, skill=%s, action=%s",
		taskID, step.ID, step.SkillName, step.Action)
	return nil
}

func (l *LogTracer) TraceStepEnd(ctx context.Context, taskID string, step *state.Step, result *state.StepResult, duration time.Duration) error {
	if l.level == "minimal" {
		return nil
	}
	status := "success"
	if !result.Success {
		status = "failed"
	}
	logger.Infof("[Tracer] Step completed: task=%s, step=%d, status=%s, duration=%v",
		taskID, step.ID, status, duration)
	if result.Error != "" && l.level == "detailed" {
		logger.Debugf("[Tracer] Step error: task=%s, step=%d, error=%s", taskID, step.ID, result.Error)
	}
	return nil
}

func (l *LogTracer) TraceError(ctx context.Context, taskID, nodeName string, err error) error {
	// Always log errors regardless of level
	logger.Errorf("[Tracer] Error occurred: task=%s, node=%s, error=%v", taskID, nodeName, err)
	return nil
}

func (l *LogTracer) TraceStateChange(ctx context.Context, taskID string, state *state.State) error {
	if l.level != "detailed" {
		return nil
	}
	// Log state summary
	summary := fmt.Sprintf("query=%s, steps=%d, current_step=%d",
		state.Query, len(state.Steps), state.CurrentStep)
	if state.FinalResult != nil {
		summary += fmt.Sprintf(", final_result=%v", state.FinalResult.Success)
	}
	logger.Debugf("[Tracer] State changed: task=%s, %s", taskID, summary)
	return nil
}

func (l *LogTracer) Close() error {
	return nil
}

