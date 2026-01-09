package mcp

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// ConnectionType represents the type of MCP connection
type ConnectionType string

const (
	ConnectionTypeStdio ConnectionType = "stdio"
	ConnectionTypeHTTP  ConnectionType = "http"
	ConnectionTypeSSE   ConnectionType = "sse"
)

// Connection represents an MCP connection
type Connection struct {
	Type     ConnectionType
	Command  *exec.Cmd
	Stdin    io.WriteCloser
	Stdout   io.ReadCloser
	Client   *Client
	Server   *Server
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewStdioConnection creates a new stdio-based MCP connection
func NewStdioConnection(command string, args []string) (*Connection, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn := &Connection{
		Type:    ConnectionTypeStdio,
		Command: cmd,
		Stdin:   stdin,
		Stdout:  stdout,
		Client:  NewClient(stdout, stdin),
		ctx:     ctx,
		cancel:  cancel,
	}

	return conn, nil
}

// Start starts the connection
func (c *Connection) Start() error {
	if err := c.Command.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Start client message loop
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.Client.Start(c.ctx)
	}()

	return nil
}

// Stop stops the connection
func (c *Connection) Stop() error {
	c.cancel()

	if c.Stdin != nil {
		c.Stdin.Close()
	}
	if c.Stdout != nil {
		c.Stdout.Close()
	}

	if c.Command != nil && c.Command.Process != nil {
		c.Command.Process.Kill()
		c.Command.Wait()
	}

	c.wg.Wait()
	return nil
}

// Initialize initializes the MCP connection
func (c *Connection) Initialize(clientInfo ClientInfo) (*InitializeResult, error) {
	return c.Client.Initialize(c.ctx, clientInfo)
}



