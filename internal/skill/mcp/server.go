package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Server represents an MCP server
type Server struct {
	name         string
	version      string
	capabilities ServerCapabilities
	handlers     map[string]HandlerFunc
	mu           sync.RWMutex
}

// HandlerFunc represents a handler function for MCP methods
type HandlerFunc func(ctx context.Context, params json.RawMessage) (interface{}, error)

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		name:         name,
		version:      version,
		capabilities: ServerCapabilities{},
		handlers:     make(map[string]HandlerFunc),
	}
}

// SetCapabilities sets server capabilities
func (s *Server) SetCapabilities(caps ServerCapabilities) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.capabilities = caps
}

// RegisterHandler registers a handler for an MCP method
func (s *Server) RegisterHandler(method string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[method] = handler
}

// HandleRequest handles a JSON-RPC request
func (s *Server) HandleRequest(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	// Handle initialize specially
	if req.Method == MethodInitialize {
		return s.handleInitialize(ctx, req)
	}

	// Get handler
	s.mu.RLock()
	handler, exists := s.handlers[req.Method]
	s.mu.RUnlock()

	if !exists {
		return NewJSONRPCResponse(req.ID, nil, NewJSONRPCError(
			ErrCodeMethodNotFound,
			fmt.Sprintf("Method not found: %s", req.Method),
			nil,
		))
	}

	// Call handler
	result, err := handler(ctx, req.Params)
	if err != nil {
		return NewJSONRPCResponse(req.ID, nil, NewJSONRPCError(
			ErrCodeInternalError,
			err.Error(),
			nil,
		))
	}

	return NewJSONRPCResponse(req.ID, result, nil)
}

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(ctx context.Context, req *JSONRPCRequest) (*JSONRPCResponse, error) {
	var params InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewJSONRPCResponse(req.ID, nil, NewJSONRPCError(
			ErrCodeInvalidParams,
			"Invalid initialize parameters",
			nil,
		))
	}

	result := InitializeResult{
		ProtocolVersion: "2024-11-05", // MCP protocol version
		Capabilities:    s.capabilities,
		ServerInfo: ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
	}

	return NewJSONRPCResponse(req.ID, result, nil)
}

// Serve serves requests from a reader and writes responses to a writer
func (s *Server) Serve(ctx context.Context, reader io.Reader, writer io.Writer) error {
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(writer)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req JSONRPCRequest
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			// Send parse error response
			resp, _ := NewJSONRPCResponse(nil, nil, NewJSONRPCError(
				ErrCodeParseError,
				"Parse error",
				nil,
			))
			encoder.Encode(resp)
			continue
		}

		// Handle request
		resp, err := s.HandleRequest(ctx, &req)
		if err != nil {
			resp, _ = NewJSONRPCResponse(req.ID, nil, NewJSONRPCError(
				ErrCodeInternalError,
				err.Error(),
				nil,
			))
		}

		// Send response
		if err := encoder.Encode(resp); err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
	}
}



