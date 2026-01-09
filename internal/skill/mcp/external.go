package mcp

import (
	"context"
	"fmt"
	"sync"
)

// ExternalServerManager manages connections to external MCP servers
type ExternalServerManager struct {
	connections map[string]*Connection
	clients     map[string]*Client
	mu          sync.RWMutex
}

// NewExternalServerManager creates a new external server manager
func NewExternalServerManager() *ExternalServerManager {
	return &ExternalServerManager{
		connections: make(map[string]*Connection),
		clients:     make(map[string]*Client),
	}
}

// Connect connects to an external MCP server
func (m *ExternalServerManager) Connect(serverName string, conn *Connection) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Start connection
	if err := conn.Start(); err != nil {
		return fmt.Errorf("failed to start connection: %w", err)
	}

	// Initialize
	clientInfo := ClientInfo{
		Name:    "opskills-agent",
		Version: "1.0.0",
	}
	if _, err := conn.Initialize(clientInfo); err != nil {
		conn.Stop()
		return fmt.Errorf("failed to initialize connection: %w", err)
	}

	// Store connection and client
	m.connections[serverName] = conn
	m.clients[serverName] = conn.Client

	return nil
}

// Disconnect disconnects from an external MCP server
func (m *ExternalServerManager) Disconnect(serverName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[serverName]
	if !exists {
		return fmt.Errorf("server not connected: %s", serverName)
	}

	if err := conn.Stop(); err != nil {
		return fmt.Errorf("failed to stop connection: %w", err)
	}

	delete(m.connections, serverName)
	delete(m.clients, serverName)

	return nil
}

// GetClient gets a client for an external MCP server
func (m *ExternalServerManager) GetClient(serverName string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[serverName]
	if !exists {
		return nil, fmt.Errorf("server not connected: %s", serverName)
	}

	return client, nil
}

// DiscoverTools discovers tools from an external MCP server
func (m *ExternalServerManager) DiscoverTools(ctx context.Context, serverName string) (*ToolsListResult, error) {
	client, err := m.GetClient(serverName)
	if err != nil {
		return nil, err
	}

	return client.ListTools(ctx)
}

// ListConnectedServers lists all connected servers
func (m *ExternalServerManager) ListConnectedServers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	servers := make([]string, 0, len(m.connections))
	for name := range m.connections {
		servers = append(servers, name)
	}

	return servers
}

// IsConnected checks if a server is connected
func (m *ExternalServerManager) IsConnected(serverName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.connections[serverName]
	return exists
}

// Close closes all connections
func (m *ExternalServerManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, conn := range m.connections {
		if err := conn.Stop(); err != nil {
			lastErr = err
		}
		delete(m.connections, name)
		delete(m.clients, name)
	}

	return lastErr
}



