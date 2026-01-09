package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

// Client represents an MCP client
type Client struct {
	reader   io.Reader
	writer   io.Writer
	decoder  *json.Decoder
	encoder  *json.Encoder
	requests map[interface{}]chan *JSONRPCResponse
	mu       sync.RWMutex
	nextID   int64
	idMu     sync.Mutex
}

// NewClient creates a new MCP client
func NewClient(reader io.Reader, writer io.Writer) *Client {
	return &Client{
		reader:   reader,
		writer:   writer,
		decoder:  json.NewDecoder(reader),
		encoder:  json.NewEncoder(writer),
		requests: make(map[interface{}]chan *JSONRPCResponse),
		nextID:   1,
	}
}

// Initialize initializes the MCP connection
func (c *Client) Initialize(ctx context.Context, clientInfo ClientInfo) (*InitializeResult, error) {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: ClientCapabilities{},
		ClientInfo:     clientInfo,
	}

	var result InitializeResult
	if err := c.Call(ctx, MethodInitialize, params, &result); err != nil {
		return nil, fmt.Errorf("initialize failed: %w", err)
	}

	// Send initialized notification
	notif, _ := NewJSONRPCRequest(nil, MethodInitialized, nil)
	c.encoder.Encode(notif)

	return &result, nil
}

// Call makes a JSON-RPC call and waits for response
func (c *Client) Call(ctx context.Context, method string, params interface{}, result interface{}) error {
	// Generate request ID
	c.idMu.Lock()
	id := c.nextID
	c.nextID++
	c.idMu.Unlock()

	// Create request
	req, err := NewJSONRPCRequest(id, method, params)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Create response channel
	respChan := make(chan *JSONRPCResponse, 1)
	c.mu.Lock()
	c.requests[id] = respChan
	c.mu.Unlock()

	// Send request
	if err := c.encoder.Encode(req); err != nil {
		c.mu.Lock()
		delete(c.requests, id)
		c.mu.Unlock()
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response
	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.requests, id)
		c.mu.Unlock()
		return ctx.Err()
	case resp := <-respChan:
		c.mu.Lock()
		delete(c.requests, id)
		c.mu.Unlock()

		if resp.Error != nil {
			return fmt.Errorf("RPC error: %s (code: %d)", resp.Error.Message, resp.Error.Code)
		}

		if result != nil && resp.Result != nil {
			if err := json.Unmarshal(resp.Result, result); err != nil {
				return fmt.Errorf("failed to unmarshal result: %w", err)
			}
		}

		return nil
	}
}

// Start starts the client message loop
func (c *Client) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var resp JSONRPCResponse
		if err := c.decoder.Decode(&resp); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to decode response: %w", err)
		}

		// Find waiting request
		c.mu.RLock()
		respChan, exists := c.requests[resp.ID]
		c.mu.RUnlock()

		if exists {
			select {
			case respChan <- &resp:
			default:
			}
		}
	}
}

// ListTools lists available tools
func (c *Client) ListTools(ctx context.Context) (*ToolsListResult, error) {
	var result ToolsListResult
	if err := c.Call(ctx, MethodToolsList, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CallTool calls a tool
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*ToolCallResult, error) {
	params := ToolCallParams{
		Name:      name,
		Arguments: arguments,
	}

	var result ToolCallResult
	if err := c.Call(ctx, MethodToolsCall, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListResources lists available resources
func (c *Client) ListResources(ctx context.Context) (*ResourcesListResult, error) {
	var result ResourcesListResult
	if err := c.Call(ctx, MethodResourcesList, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReadResource reads a resource
func (c *Client) ReadResource(ctx context.Context, uri string) (*ResourceReadResult, error) {
	params := ResourceReadParams{
		URI: uri,
	}

	var result ResourceReadResult
	if err := c.Call(ctx, MethodResourcesRead, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}



