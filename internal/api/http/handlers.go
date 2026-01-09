package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hb-chen/opskills/internal/agent"
	"github.com/hb-chen/opskills/internal/state"
	"github.com/hb-chen/opskills/pkg/logger"
)

// Handlers contains HTTP handlers
type Handlers struct {
	pipeline *agent.Pipeline
	logger   logger.Logger
	states   map[string]*state.State // In-memory state storage (should be replaced with proper storage)
}

// NewHandlers creates new HTTP handlers
func NewHandlers(pipeline *agent.Pipeline, log logger.Logger) *Handlers {
	return &Handlers{
		pipeline: pipeline,
		logger:   log,
		states:   make(map[string]*state.State),
	}
}

// SubmitTaskRequest represents a task submission request
type SubmitTaskRequest struct {
	Query string `json:"query"`
}

// SubmitTaskResponse represents a task submission response
type SubmitTaskResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// TaskStatusResponse represents a task status response
type TaskStatusResponse struct {
	TaskID    string        `json:"task_id"`
	Status    string        `json:"status"`
	State     *state.State  `json:"state,omitempty"`
	Error     string        `json:"error,omitempty"`
}

// SubmitTask handles task submission
func (h *Handlers) SubmitTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	// Generate task ID
	taskID := uuid.New().String()

	h.logger.Info("Submitting task %s: %s", taskID, req.Query)

	// Execute task asynchronously
	go func() {
		ctx := r.Context()
		state, err := h.pipeline.Execute(ctx, req.Query, taskID)
		
		// Store state
		h.states[taskID] = state
		
		if err != nil {
			h.logger.Error("Task %s failed: %v", taskID, err)
		} else {
			h.logger.Info("Task %s completed", taskID)
		}
	}()

	// Return immediate response
	resp := SubmitTaskResponse{
		TaskID:  taskID,
		Status:  "pending",
		Message: "Task submitted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

// GetTaskStatus handles task status queries
func (h *Handlers) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	taskID := r.URL.Query().Get("task_id")
	if taskID == "" {
		http.Error(w, "task_id is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Querying status for task %s", taskID)

	// Get state from storage
	state, exists := h.states[taskID]
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
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

	resp := TaskStatusResponse{
		TaskID: taskID,
		Status: status,
		State:  state,
	}

	if state.Error != "" {
		resp.Error = state.Error
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HealthCheck handles health check requests
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// HandleRun handles SSE task execution requests
func (h *Handlers) HandleRun(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Generate task ID
	taskID := uuid.New().String()

	// Send initial status
	h.sendSSE(w, flusher, "update", map[string]string{"step": "正在初始化任务..."})
	h.sendSSE(w, flusher, "log", map[string]string{"message": fmt.Sprintf("任务 ID: %s", taskID)})
	h.sendSSE(w, flusher, "log", map[string]string{"message": fmt.Sprintf("查询: %s", query)})

	// Execute task in goroutine
	resultChan := make(chan *state.State, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errChan)

		ctx := r.Context()
		h.sendSSE(w, flusher, "update", map[string]string{"step": "正在执行任务..."})
		h.sendSSE(w, flusher, "log", map[string]string{"message": "开始执行 Pipeline..."})

		state, err := h.pipeline.Execute(ctx, query, taskID)
		if err != nil {
			errChan <- err
			return
		}

		// Store state
		h.states[taskID] = state

		resultChan <- state
	}()

	// Wait for result or error
	select {
	case res := <-resultChan:
		if res != nil {
			h.sendSSE(w, flusher, "log", map[string]string{"message": "任务执行完成"})
			
			// Send result
			resultData := map[string]interface{}{
				"state": res,
			}
			h.sendSSEResult(w, flusher, "result", resultData)
		}
	case err := <-errChan:
		if err != nil {
			h.sendSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		}
	case <-r.Context().Done():
		h.sendSSE(w, flusher, "error", map[string]string{"message": "请求已取消"})
	}
}

// sendSSE sends a Server-Sent Event
func (h *Handlers) sendSSE(w http.ResponseWriter, flusher http.Flusher, eventType string, data map[string]string) {
	payload := map[string]interface{}{
		"type": eventType,
	}

	// Merge data into payload
	for k, v := range data {
		payload[k] = v
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal SSE payload: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonPayload)
	flusher.Flush()
}

// sendSSEResult sends a result event with complex data
func (h *Handlers) sendSSEResult(w http.ResponseWriter, flusher http.Flusher, eventType string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"type": eventType,
	}

	// Merge data into payload
	for k, v := range data {
		payload[k] = v
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal SSE result payload: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonPayload)
	flusher.Flush()
}



