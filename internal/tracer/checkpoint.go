package tracer

import (
	"context"
	"time"

	"github.com/hb-chen/opskills/internal/state"
)

// CheckpointTracer implements ExecutionTracer as a no-op tracer
//
// Design Note: Checkpoint data is automatically saved by langgraphgo's checkpoint store
// during graph execution. This tracer doesn't actively trace events because:
// 1. Checkpoint store already captures state snapshots at each node
// 2. Historical data can be retrieved directly from the checkpoint store
// 3. This tracer serves as a placeholder for future checkpoint-based analysis tools
//
// To read checkpoint history, use the checkpoint store's API directly:
// - Get checkpoint snapshots for a task
// - Replay execution history
// - Analyze state transitions
type CheckpointTracer struct {
	// Future: Could store checkpoint store reference for reading history
}

// NewCheckpointTracer creates a new checkpoint tracer
// This is a no-op tracer since checkpoint data is managed by langgraphgo's checkpoint store
func NewCheckpointTracer() *CheckpointTracer {
	return &CheckpointTracer{}
}

func (c *CheckpointTracer) TraceNodeStart(ctx context.Context, nodeName, taskID string) error {
	// No-op: Checkpoint store automatically saves state at node boundaries
	return nil
}

func (c *CheckpointTracer) TraceNodeEnd(ctx context.Context, nodeName, taskID string, duration time.Duration) error {
	// No-op: Checkpoint store automatically saves state at node boundaries
	return nil
}

func (c *CheckpointTracer) TraceLLMRequest(ctx context.Context, taskID, prompt string) error {
	// No-op: LLM interactions are captured in state messages
	return nil
}

func (c *CheckpointTracer) TraceLLMResponse(ctx context.Context, taskID, response string, duration time.Duration) error {
	// No-op: LLM interactions are captured in state messages
	return nil
}

func (c *CheckpointTracer) TraceStepStart(ctx context.Context, taskID string, step *state.Step) error {
	// No-op: Step execution is captured in state
	return nil
}

func (c *CheckpointTracer) TraceStepEnd(ctx context.Context, taskID string, step *state.Step, result *state.StepResult, duration time.Duration) error {
	// No-op: Step results are captured in state
	return nil
}

func (c *CheckpointTracer) TraceError(ctx context.Context, taskID, nodeName string, err error) error {
	// No-op: Errors are captured in state
	return nil
}

func (c *CheckpointTracer) TraceStateChange(ctx context.Context, taskID string, state *state.State) error {
	// No-op: State changes are automatically checkpointed by langgraphgo
	return nil
}

func (c *CheckpointTracer) Close() error {
	// No-op: Nothing to clean up
	return nil
}

