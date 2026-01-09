package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hb-chen/opskills/internal/skill"
	"github.com/hb-chen/opskills/internal/skill/direct"
	"github.com/hb-chen/opskills/internal/skill/mcp"
)

// KubeKeyServer implements an MCP server for KubeKey skills
type KubeKeyServer struct {
	server   *mcp.Server
	registry *skill.Registry
	executor *direct.DirectExecutor
	adapter  *skill.MCPAdapter
}

// NewKubeKeyServer creates a new KubeKey MCP server
func NewKubeKeyServer(skillsDir string) (*KubeKeyServer, error) {
	// Load KubeKey skill
	loader := skill.NewLoader(skillsDir)
	kubekeySkill, err := loader.LoadSkill("kubekey")
	if err != nil {
		return nil, fmt.Errorf("failed to load kubekey skill: %w", err)
	}

	// Create registry and register skill
	registry := skill.NewRegistry()
	if err := registry.Register(kubekeySkill); err != nil {
		return nil, fmt.Errorf("failed to register kubekey skill: %w", err)
	}

	// Create executor and adapter
	executor := direct.NewDirectExecutor(30 * 60 * 1000000000) // 30 minutes in nanoseconds
	adapter := skill.NewMCPAdapter(registry)

	// Create MCP server
	server := mcp.NewServer("kubekey-mcp-server", "1.0.0")
	
	// Set capabilities
	server.SetCapabilities(mcp.ServerCapabilities{
		Tools: &mcp.ToolsCapability{},
		Resources: &mcp.ResourcesCapability{
			Subscribe:   false,
			ListChanged: false,
		},
	})

	// Register handlers
	kks := &KubeKeyServer{
		server:   server,
		registry: registry,
		executor: executor,
		adapter:  adapter,
	}

	kks.registerHandlers()

	return kks, nil
}

// registerHandlers registers MCP method handlers
func (kks *KubeKeyServer) registerHandlers() {
	// Register tools/list handler
	kks.server.RegisterHandler(mcp.MethodToolsList, kks.handleToolsList)

	// Register tools/call handler
	kks.server.RegisterHandler(mcp.MethodToolsCall, kks.handleToolsCall)

	// Register resources/list handler
	kks.server.RegisterHandler(mcp.MethodResourcesList, kks.handleResourcesList)

	// Register resources/read handler
	kks.server.RegisterHandler(mcp.MethodResourcesRead, kks.handleResourcesRead)
}

// handleToolsList handles tools/list requests
func (kks *KubeKeyServer) handleToolsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	skills := kks.registry.List()
	tools, err := kks.adapter.SkillsToTools(skills)
	if err != nil {
		return nil, fmt.Errorf("failed to convert skills to tools: %w", err)
	}

	return mcp.ToolsListResult{
		Tools: tools,
	}, nil
}

// handleToolsCall handles tools/call requests
func (kks *KubeKeyServer) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var callParams mcp.ToolCallParams
	if err := json.Unmarshal(params, &callParams); err != nil {
		return nil, fmt.Errorf("invalid tool call parameters: %w", err)
	}

	// Get skill
	skill, err := kks.registry.Get(callParams.Name)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", callParams.Name)
	}

	// Convert tool call to skill params
	skillParams, err := kks.adapter.ToolCallToSkillParams(callParams)
	if err != nil {
		return nil, fmt.Errorf("failed to convert tool call: %w", err)
	}

	// Execute skill
	result, err := kks.executor.Execute(skill, skillParams)
	if err != nil {
		return nil, fmt.Errorf("skill execution failed: %w", err)
	}

	// Convert result to tool call result
	toolResult, err := kks.adapter.SkillResultToToolResult(result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert result: %w", err)
	}

	return toolResult, nil
}

// handleResourcesList handles resources/list requests
func (kks *KubeKeyServer) handleResourcesList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	skills := kks.registry.List()
	resources := skill.SkillsToResources(skills)

	// Add skill-specific resources (configs, scripts)
	for _, s := range skills {
		// Add SKILL.md as resource
		if s.SKILLPath != "" {
			resources = append(resources, mcp.Resource{
				URI:         fmt.Sprintf("skill://%s/skill.md", s.Name),
				Name:        fmt.Sprintf("%s Skill Documentation", s.Name),
				Description: s.Description,
				MimeType:    "text/markdown",
			})
		}

		// Add scripts as resources
		if s.ScriptsPath != "" {
			entries, err := os.ReadDir(s.ScriptsPath)
			if err == nil {
				for _, entry := range entries {
					if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sh" {
						resources = append(resources, mcp.Resource{
							URI:         fmt.Sprintf("skill://%s/script/%s", s.Name, entry.Name()),
							Name:        fmt.Sprintf("%s: %s", s.Name, entry.Name()),
							Description: fmt.Sprintf("Script file for %s", s.Name),
							MimeType:    "text/x-shellscript",
						})
					}
				}
			}
		}
	}

	return mcp.ResourcesListResult{
		Resources: resources,
	}, nil
}

// handleResourcesRead handles resources/read requests
func (kks *KubeKeyServer) handleResourcesRead(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var readParams mcp.ResourceReadParams
	if err := json.Unmarshal(params, &readParams); err != nil {
		return nil, fmt.Errorf("invalid resource read parameters: %w", err)
	}

	// Parse URI
	// Format: skill://<skill-name>/<type>/<path>
	// Examples:
	// - skill://kubekey/skill.md
	// - skill://kubekey/script/create_cluster.sh
	// - skill://kubekey/config/cluster-config.yaml

	uri := readParams.URI
	if !isSkillURI(uri) {
		return nil, fmt.Errorf("invalid skill URI: %s", uri)
	}

	// Extract skill name and path
	skillName, resourcePath := parseSkillURI(uri)

	// Get skill
	skill, err := kks.registry.Get(skillName)
	if err != nil {
		return nil, fmt.Errorf("skill not found: %s", skillName)
	}

	// Read resource based on type
	var content string

	switch {
	case resourcePath == "skill.md":
		// Read SKILL.md
		data, err := os.ReadFile(skill.SKILLPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read SKILL.md: %w", err)
		}
		content = string(data)

	case filepath.Dir(resourcePath) == "script":
		// Read script file
		scriptName := filepath.Base(resourcePath)
		scriptPath := filepath.Join(skill.ScriptsPath, scriptName)
		data, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read script: %w", err)
		}
		content = string(data)

	case filepath.Dir(resourcePath) == "config":
		// Read config file
		configName := filepath.Base(resourcePath)
		configPath := filepath.Join(skill.BasePath, "examples", configName)
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		content = string(data)

	default:
		return nil, fmt.Errorf("unknown resource type: %s", resourcePath)
	}

	return mcp.ResourceReadResult{
		Contents: []mcp.Content{
			{
				Type: "text",
				Text: content,
			},
		},
	}, nil
}

// isSkillURI checks if a URI is a skill URI
func isSkillURI(uri string) bool {
	return len(uri) > 8 && uri[:8] == "skill://"
}

// parseSkillURI parses a skill URI into skill name and resource path
func parseSkillURI(uri string) (skillName, resourcePath string) {
	// Remove "skill://" prefix
	path := uri[8:]
	
	// Find first "/"
	idx := 0
	for i, char := range path {
		if char == '/' {
			idx = i
			break
		}
	}

	if idx == 0 {
		return path, ""
	}

	return path[:idx], path[idx+1:]
}

// GetServer returns the underlying MCP server
func (kks *KubeKeyServer) GetServer() *mcp.Server {
	return kks.server
}



