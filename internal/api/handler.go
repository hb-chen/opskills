package api

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

// Handler handles special HTTP routes that cannot be implemented via gRPC/gateway
// (e.g., SSE streaming, WebSocket)
type Handler struct {
	pipeline *agent.Pipeline
	states   map[string]*state.State // In-memory state storage (should be replaced with proper storage)
}

// NewHandler creates a new special route handler
func NewHandler(pipeline *agent.Pipeline) *Handler {
	return &Handler{
		pipeline: pipeline,
		states:   make(map[string]*state.State),
	}
}

// HandleRun handles SSE task execution requests
// This is a special route that cannot be implemented via gRPC/gateway due to streaming requirements
func (h *Handler) HandleRun(w http.ResponseWriter, r *http.Request) {
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

// HealthCheck handles health check requests
// This is a simple endpoint that doesn't need gRPC
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// sendSSE sends a Server-Sent Event
func (h *Handler) sendSSE(w http.ResponseWriter, flusher http.Flusher, eventType string, data map[string]string) {
	payload := map[string]interface{}{
		"type": eventType,
	}

	// Merge data into payload
	for k, v := range data {
		payload[k] = v
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal SSE payload: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonPayload)
	flusher.Flush()
}

// sendSSEResult sends a result event with complex data
func (h *Handler) sendSSEResult(w http.ResponseWriter, flusher http.Flusher, eventType string, data map[string]interface{}) {
	payload := map[string]interface{}{
		"type": eventType,
	}

	// Merge data into payload
	for k, v := range data {
		payload[k] = v
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf("Failed to marshal SSE result payload: %v", err)
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", jsonPayload)
	flusher.Flush()
}

