package skill

import (
	"context"
	"fmt"

	"github.com/hb-chen/opskills/internal/skill/mcp"
)

// Router routes skill execution to the appropriate executor
type Router struct {
	directExecutor Executor // Use interface instead of concrete type
	mcpClients      map[string]*mcp.Client
	config          *Config
	registry        *Registry
}

// NewRouter creates a new skill router
func NewRouter(directExecutor Executor, config *Config, registry *Registry) *Router {
	return &Router{
		directExecutor: directExecutor,
		mcpClients:      make(map[string]*mcp.Client),
		config:          config,
		registry:        registry,
	}
}

// Execute executes a skill using the appropriate method
func (r *Router) Execute(skillName string, params ExecutionParams) (*ExecutionResult, error) {
	// Get skill
	skill, err := r.registry.Get(skillName)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}

	// Determine execution mode
	mode := r.determineExecutionMode(skillName)

	switch mode {
	case ExecutionModeDirect:
		return r.executeDirect(skill, params)
	case ExecutionModeMCP:
		return r.executeMCP(skillName, params)
	case ExecutionModeAuto:
		// Auto mode: try direct first, fallback to MCP if needed
		result, err := r.executeDirect(skill, params)
		if err == nil {
			return result, nil
		}
		// Fallback to MCP
		return r.executeMCP(skillName, params)
	default:
		return nil, fmt.Errorf("unknown execution mode: %s", mode)
	}
}

// determineExecutionMode determines the execution mode for a skill
func (r *Router) determineExecutionMode(skillName string) ExecutionMode {
	// Check config first
	if r.config != nil {
		if skillConfig, exists := r.config.GetSkillConfig(skillName); exists {
			if skillConfig.ExecutionMode != ExecutionModeAuto {
				return skillConfig.ExecutionMode
			}
		}
	}

	// Default to direct for now
	// In the future, this could check skill metadata or other factors
	return ExecutionModeDirect
}

// executeDirect executes a skill directly
func (r *Router) executeDirect(skill *Skill, params ExecutionParams) (*ExecutionResult, error) {
	if r.directExecutor == nil {
		return nil, fmt.Errorf("direct executor not available")
	}
	return r.directExecutor.Execute(skill, params)
}

// executeMCP executes a skill via MCP
func (r *Router) executeMCP(skillName string, params ExecutionParams) (*ExecutionResult, error) {
	// Get MCP server config for this skill
	if r.config == nil {
		return nil, fmt.Errorf("MCP execution requires configuration")
	}

	skillConfig, exists := r.config.GetSkillConfig(skillName)
	if !exists || skillConfig.MCPServer == "" {
		return nil, fmt.Errorf("MCP server not configured for skill: %s", skillName)
	}

	// Get or create MCP client
	client, err := r.getMCPClient(skillConfig.MCPServer)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP client: %w", err)
	}

	// Convert params to tool call arguments
	arguments := make(map[string]interface{})
	for k, v := range params {
		arguments[k] = v
	}

	// Call tool via MCP
	ctx := context.Background()
	result, err := client.CallTool(ctx, skillName, arguments)
	if err != nil {
		return nil, fmt.Errorf("MCP tool call failed: %w", err)
	}

	// Convert MCP result to ExecutionResult
	return r.convertMCPResult(result), nil
}

// getMCPClient gets or creates an MCP client for a server
func (r *Router) getMCPClient(serverName string) (*mcp.Client, error) {
	// Check if client already exists
	if client, exists := r.mcpClients[serverName]; exists {
		return client, nil
	}

	// Get server config
	if r.config == nil {
		return nil, fmt.Errorf("configuration not available")
	}

	serverConfig, exists := r.config.MCPServers[serverName]
	if !exists {
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	// Create connection based on type
	// For now, only stdio is implemented
	if serverConfig.Type != "stdio" {
		return nil, fmt.Errorf("unsupported MCP server type: %s", serverConfig.Type)
	}

	// Create stdio connection
	conn, err := mcp.NewStdioConnection(serverConfig.Command, serverConfig.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP connection: %w", err)
	}

	// Start connection
	if err := conn.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP connection: %w", err)
	}

	// Initialize
	clientInfo := mcp.ClientInfo{
		Name:    "opskills-agent",
		Version: "1.0.0",
	}
	if _, err := conn.Initialize(clientInfo); err != nil {
		conn.Stop()
		return nil, fmt.Errorf("failed to initialize MCP connection: %w", err)
	}

	// Store client
	r.mcpClients[serverName] = conn.Client

	return conn.Client, nil
}

// convertMCPResult converts an MCP tool call result to ExecutionResult
func (r *Router) convertMCPResult(result *mcp.ToolCallResult) *ExecutionResult {
	execResult := &ExecutionResult{
		Success: !result.IsError,
	}

	// Extract content
	for _, content := range result.Content {
		if content.Type == "text" {
			if execResult.Success {
				execResult.Output += content.Text + "\n"
			} else {
				execResult.Error += content.Text + "\n"
			}
		}
	}

	return execResult
}

// GetRegistry returns the skill registry
func (r *Router) GetRegistry() *Registry {
	return r.registry
}

