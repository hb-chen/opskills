package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/hb-chen/opskills/internal/skill/mcp/servers"
	"github.com/hb-chen/opskills/pkg/logger"
)

func main() {
	skillsDir := flag.String("skills-dir", "./skills", "Skills directory")
	flag.Parse()

	// Initialize logger (using zap)
	// For now, use global logger functions
	logger.Info("Starting KubeKey MCP Server...")

	// Create server
	server, err := servers.NewKubeKeyServer(*skillsDir)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Get MCP server
	mcpServer := server.GetServer()

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down...")
		cancel()
	}()

	// Serve on stdio
	logger.Info("MCP Server ready (stdio)")
	if err := mcpServer.Serve(ctx, os.Stdin, os.Stdout); err != nil {
		if err != context.Canceled {
			logger.Fatalf("Server error: %v", err)
		}
	}

	logger.Info("Server stopped")
}



