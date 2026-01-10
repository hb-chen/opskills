package tracer

import (
	"context"
	"fmt"
	"time"

	"github.com/hb-chen/opskills/internal/state"
	"github.com/hb-chen/opskills/pkg/logger"
)

// ExecutionTracer interface for tracing execution events
type ExecutionTracer interface {
	// TraceNodeStart records when a node starts execution
	TraceNodeStart(ctx context.Context, nodeName, taskID string) error

	// TraceNodeEnd records when a node completes execution
	TraceNodeEnd(ctx context.Context, nodeName, taskID string, duration time.Duration) error

	// TraceLLMRequest records an LLM request
	TraceLLMRequest(ctx context.Context, taskID, prompt string) error

	// TraceLLMResponse records an LLM response
	TraceLLMResponse(ctx context.Context, taskID, response string, duration time.Duration) error

	// TraceStepStart records when a step starts execution
	TraceStepStart(ctx context.Context, taskID string, step *state.Step) error

	// TraceStepEnd records when a step completes execution
	TraceStepEnd(ctx context.Context, taskID string, step *state.Step, result *state.StepResult, duration time.Duration) error

	// TraceError records an error event
	TraceError(ctx context.Context, taskID, nodeName string, err error) error

	// TraceStateChange records a state change (optional, for detailed tracing)
	TraceStateChange(ctx context.Context, taskID string, state *state.State) error

	// Close closes the tracer and flushes any pending data
	Close() error
}

// TraceEventType represents the type of trace event
type TraceEventType string

const (
	TraceEventNodeStart   TraceEventType = "NodeStart"
	TraceEventNodeEnd     TraceEventType = "NodeEnd"
	TraceEventLLMRequest  TraceEventType = "LLMRequest"
	TraceEventLLMResponse TraceEventType = "LLMResponse"
	TraceEventStepStart   TraceEventType = "StepStart"
	TraceEventStepEnd     TraceEventType = "StepEnd"
	TraceEventError       TraceEventType = "Error"
	TraceEventStateChange TraceEventType = "StateChange"
)

// TraceEvent represents a single trace event
type TraceEvent struct {
	Type      TraceEventType
	Timestamp time.Time
	NodeName  string
	TaskID    string
	Data      map[string]interface{}
}

// MultiTracer combines multiple tracers
type MultiTracer struct {
	tracers []ExecutionTracer
}

// NewMultiTracer creates a new multi-tracer that forwards events to all tracers
func NewMultiTracer(tracers ...ExecutionTracer) *MultiTracer {
	return &MultiTracer{tracers: tracers}
}

func (m *MultiTracer) TraceNodeStart(ctx context.Context, nodeName, taskID string) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceNodeStart(ctx, nodeName, taskID); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace node start: tracer=%T, node=%s, task=%s, error=%v",
				tracer, nodeName, taskID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceNodeEnd(ctx context.Context, nodeName, taskID string, duration time.Duration) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceNodeEnd(ctx, nodeName, taskID, duration); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace node end: tracer=%T, node=%s, task=%s, error=%v",
				tracer, nodeName, taskID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceLLMRequest(ctx context.Context, taskID, prompt string) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceLLMRequest(ctx, taskID, prompt); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace LLM request: tracer=%T, task=%s, error=%v",
				tracer, taskID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceLLMResponse(ctx context.Context, taskID, response string, duration time.Duration) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceLLMResponse(ctx, taskID, response, duration); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace LLM response: tracer=%T, task=%s, error=%v",
				tracer, taskID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceStepStart(ctx context.Context, taskID string, step *state.Step) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceStepStart(ctx, taskID, step); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace step start: tracer=%T, task=%s, step=%d, error=%v",
				tracer, taskID, step.ID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceStepEnd(ctx context.Context, taskID string, step *state.Step, result *state.StepResult, duration time.Duration) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceStepEnd(ctx, taskID, step, result, duration); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace step end: tracer=%T, task=%s, step=%d, error=%v",
				tracer, taskID, step.ID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceError(ctx context.Context, taskID, nodeName string, err error) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if traceErr := tracer.TraceError(ctx, taskID, nodeName, err); traceErr != nil {
			logger.Warnf("[MultiTracer] Failed to trace error: tracer=%T, task=%s, node=%s, error=%v",
				tracer, taskID, nodeName, traceErr)
			lastErr = traceErr
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) TraceStateChange(ctx context.Context, taskID string, state *state.State) error {
	var lastErr error
	for _, tracer := range m.tracers {
		if err := tracer.TraceStateChange(ctx, taskID, state); err != nil {
			logger.Warnf("[MultiTracer] Failed to trace state change: tracer=%T, task=%s, error=%v",
				tracer, taskID, err)
			lastErr = err
			// Continue with other tracers (best effort)
		}
	}
	return lastErr
}

func (m *MultiTracer) Close() error {
	var errors []error
	for _, tracer := range m.tracers {
		if err := tracer.Close(); err != nil {
			logger.Warnf("[MultiTracer] Failed to close tracer: tracer=%T, error=%v", tracer, err)
			errors = append(errors, err)
			// Continue closing other tracers
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("failed to close %d tracer(s): %v", len(errors), errors)
	}
	return nil
}
