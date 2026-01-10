package api

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hb-chen/opskills/internal/agent"
	"github.com/hb-chen/opskills/internal/state"
	"github.com/hb-chen/opskills/pkg/logger"
	"github.com/hb-chen/opskills/proto/common"
	"github.com/hb-chen/opskills/proto/ops"
	"google.golang.org/protobuf/types/known/anypb"
)

// Service implements the OpsService gRPC service
type Service struct {
	ops.UnimplementedOpsServiceServer
	pipeline *agent.Pipeline
	states   map[string]*state.State // In-memory state storage (should be replaced with proper storage)
}

// NewService creates a new OpsService implementation
func NewService(pipeline *agent.Pipeline) *Service {
	return &Service{
		pipeline: pipeline,
		states:   make(map[string]*state.State),
	}
}

// SubmitTask submits a new Ops task
func (s *Service) SubmitTask(ctx context.Context, req *ops.SubmitTaskRequest) (*common.Response, error) {
	if req.Query == "" {
		return &common.Response{
			Code:    400,
			Message: "Query is required",
		}, nil
	}

	// Generate task ID
	taskID := uuid.New().String()
	if taskID == "" {
		taskID = uuid.New().String()
	}

	logger.Infof("Submitting task %s: %s", taskID, req.Query)

	// Execute task asynchronously
	go func() {
		state, err := s.pipeline.Execute(ctx, req.Query, taskID)

		// Store state
		s.states[taskID] = state

		if err != nil {
			logger.Errorf("Task %s failed: %v", taskID, err)
		} else {
			logger.Infof("Task %s completed", taskID)
		}
	}()

	// Create response data
	taskData := &ops.Task{
		TaskId:    taskID,
		Query:     req.Query,
		Status:    "pending",
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	anyData, err := anypb.New(taskData)
	if err != nil {
		return &common.Response{
			Code:    500,
			Message: "Failed to marshal task data",
		}, nil
	}

	return &common.Response{
		Code:    202,
		Message: "Task submitted successfully",
		Data:    anyData,
	}, nil
}

// GetTaskStatus gets the status of a task
func (s *Service) GetTaskStatus(ctx context.Context, req *ops.GetTaskStatusRequest) (*common.Response, error) {
	if req.TaskId == "" {
		return &common.Response{
			Code:    400,
			Message: "task_id is required",
		}, nil
	}

	logger.Debugf("Querying status for task %s", req.TaskId)

	// Get state from storage
	state, exists := s.states[req.TaskId]
	if !exists {
		return &common.Response{
			Code:    404,
			Message: "Task not found",
		}, nil
	}

	// Determine status
	status := "running"
	if state.FinalResult != nil {
		if state.FinalResult.Success {
			status = "completed"
		} else {
			status = "failed"
		}
	} else if state.Error != "" {
		status = "failed"
	}

	// Convert state to proto Task
	task := stateToProtoTask(req.TaskId, state, status)

	anyData, err := anypb.New(task)
	if err != nil {
		return &common.Response{
			Code:    500,
			Message: "Failed to marshal task data",
		}, nil
	}

	return &common.Response{
		Code:    200,
		Message: "Success",
		Data:    anyData,
	}, nil
}

// ListTasks lists all tasks
func (s *Service) ListTasks(ctx context.Context, req *ops.ListTasksRequest) (*common.Response, error) {
	// Simple implementation: return all tasks
	// In production, this should support pagination and filtering
	tasks := make([]*ops.Task, 0, len(s.states))

	for taskID, state := range s.states {
		status := "running"
		if state.FinalResult != nil {
			if state.FinalResult.Success {
				status = "completed"
			} else {
				status = "failed"
			}
		} else if state.Error != "" {
			status = "failed"
		}

		// Apply filters
		if req.Status != "" && status != req.Status {
			continue
		}

		task := stateToProtoTask(taskID, state, status)
		tasks = append(tasks, task)
	}

	// Create a wrapper message for the list
	tasksData := map[string]interface{}{
		"tasks": tasks,
		"total": len(tasks),
	}

	jsonData, err := json.Marshal(tasksData)
	if err != nil {
		return &common.Response{
			Code:    500,
			Message: "Failed to marshal tasks data",
		}, nil
	}

	// Use Any with JSON string for complex data
	anyData, err := anypb.New(&common.Empty{})
	if err != nil {
		return &common.Response{
			Code:    500,
			Message: "Failed to create any data",
		}, nil
	}

	// For now, return JSON in message field
	// In production, define a proper proto message for ListTasksResponse
	return &common.Response{
		Code:    200,
		Message: string(jsonData),
		Data:    anyData,
	}, nil
}

// CancelTask cancels a running task
func (s *Service) CancelTask(ctx context.Context, req *ops.CancelTaskRequest) (*common.Response, error) {
	if req.TaskId == "" {
		return &common.Response{
			Code:    400,
			Message: "task_id is required",
		}, nil
	}

	// Check if task exists
	taskState, exists := s.states[req.TaskId]
	if !exists {
		return &common.Response{
			Code:    404,
			Message: "Task not found",
		}, nil
	}

	// Check if task is still running
	if taskState.FinalResult != nil {
		return &common.Response{
			Code:    400,
			Message: "Task is already completed",
		}, nil
	}

	// Mark task as cancelled
	taskState.Error = "Task cancelled by user"
	taskState.FinalResult = &state.FinalResult{
		Success: false,
		Error:   "Task cancelled by user",
		Summary: "Task was cancelled by user request",
	}
	taskState.UpdatedAt = time.Now().Format(time.RFC3339)

	return &common.Response{
		Code:    200,
		Message: "Task cancelled successfully",
	}, nil
}

// stateToProtoTask converts state.State to proto.Task
func stateToProtoTask(taskID string, s *state.State, status string) *ops.Task {
	task := &ops.Task{
		TaskId:    taskID,
		Query:     s.Query,
		Status:    status,
		CreatedAt: s.StartedAt,
		UpdatedAt: s.UpdatedAt,
	}

	if s.Error != "" {
		task.Error = s.Error
	}

	// Convert plan to string
	if s.Plan != nil {
		planJSON, err := json.Marshal(s.Plan)
		if err == nil {
			task.Plan = string(planJSON)
		}
	}

	// Convert results
	if s.Results != nil {
		task.Results = make([]*ops.StepResult, len(s.Results))
		for i, result := range s.Results {
			task.Results[i] = &ops.StepResult{
				StepId:  int32(result.StepID),
				Success: result.Success,
				Output:  result.Output,
				Error:   result.Error,
			}
		}
	}

	return task
}
