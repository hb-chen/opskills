package skill

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ExecutionMode represents how a skill should be executed
type ExecutionMode string

const (
	ExecutionModeDirect ExecutionMode = "direct"
	ExecutionModeMCP    ExecutionMode = "mcp"
	ExecutionModeAuto   ExecutionMode = "auto"
)

// SkillConfig represents configuration for a skill
type SkillConfig struct {
	Name          string        `yaml:"name"`
	ExecutionMode ExecutionMode `yaml:"execution_mode"`
	MCPServer     string        `yaml:"mcp_server,omitempty"` // MCP server name if using MCP
}

// Config represents the skills configuration
type Config struct {
	Skills     map[string]SkillConfig     `yaml:"skills"`
	MCPServers map[string]MCPServerConfig `yaml:"mcp_servers,omitempty"`
}

// MCPServerConfig represents configuration for an MCP server
type MCPServerConfig struct {
	Type    string            `yaml:"type"` // stdio, http, sse
	Command string            `yaml:"command,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
	URL     string            `yaml:"url,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
}

// LoadConfig loads skill configuration from a file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetSkillConfig gets configuration for a specific skill
func (c *Config) GetSkillConfig(skillName string) (SkillConfig, bool) {
	config, exists := c.Skills[skillName]
	return config, exists
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Skills:     make(map[string]SkillConfig),
		MCPServers: make(map[string]MCPServerConfig),
	}
}
