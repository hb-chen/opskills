package config

import (
	"github.com/spf13/viper"
)

var v *viper.Viper

// Init initializes the viper instance
func Init() {
	v = viper.New()
}

// Viper returns the viper instance
func Viper() *viper.Viper {
	return v
}

// Server configuration
type Server struct {
	HTTP HTTPConfig `mapstructure:"http" yaml:"http"`
	GRPC GRPCConfig `mapstructure:"grpc" yaml:"grpc"`
}

type HTTPConfig struct {
	Addr string `mapstructure:"addr" yaml:"addr"`
}

type GRPCConfig struct {
	Addr string `mapstructure:"addr" yaml:"addr"`
}

// Log configuration
type Log struct {
	Level string `mapstructure:"level" yaml:"level"`
	Path  string `mapstructure:"path" yaml:"path"`
	Debug bool   `mapstructure:"debug" yaml:"debug"`
}

// LLM configuration
type LLM struct {
	Provider string `mapstructure:"provider" yaml:"provider"`
	APIKey   string `mapstructure:"api_key" yaml:"api_key"`
	Model    string `mapstructure:"model" yaml:"model"`
	URL      string `mapstructure:"url" yaml:"url"` // Custom LLM service URL
}

// Skills configuration
type Skills struct {
	Dir string `mapstructure:"dir" yaml:"dir"`
}

// Checkpoint configuration - independent from tracing
// Checkpoint is used for conversation memory, state recovery, and rollback
type CheckpointConfig struct {
	Enabled   bool   `mapstructure:"enabled" yaml:"enabled"`
	StoreType string `mapstructure:"store_type" yaml:"store_type"` // file, memory, redis, postgres
	Path      string `mapstructure:"path" yaml:"path"`
}

// Tracing configuration - independent from checkpoint
// Tracing is used for execution observation and reporting
type Tracing struct {
	Enabled  bool             `mapstructure:"enabled" yaml:"enabled"`
	Markdown MarkdownConfig  `mapstructure:"markdown" yaml:"markdown"`
	Log      LogTracingConfig `mapstructure:"log" yaml:"log"`
}

type MarkdownConfig struct {
	Enabled   bool   `mapstructure:"enabled" yaml:"enabled"`
	OutputDir string `mapstructure:"output_dir" yaml:"output_dir"`
}

type LogTracingConfig struct {
	Level string `mapstructure:"level" yaml:"level"` // minimal, standard, detailed
}

// Config represents the application configuration
type Config struct {
	Server Server `mapstructure:"server" yaml:"server"`
	Log    Log    `mapstructure:"log" yaml:"log"`
	LLM    LLM    `mapstructure:"llm" yaml:"llm"`
	Skills Skills `mapstructure:"skills" yaml:"skills"`
	Agent  Agent  `mapstructure:"agent" yaml:"agent"`
}

// Agent configuration
type Agent struct {
	Checkpoint CheckpointConfig `mapstructure:"checkpoint" yaml:"checkpoint"`
	Tracing    Tracing          `mapstructure:"tracing" yaml:"tracing"`
}

// LoadConfig loads configuration from viper
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	if err := Viper().Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.Server.HTTP.Addr == "" {
		cfg.Server.HTTP.Addr = ":8080"
	}
	if cfg.Server.GRPC.Addr == "" {
		cfg.Server.GRPC.Addr = ":8081"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "INFO"
	}
	if cfg.Log.Path == "" {
		cfg.Log.Path = "./log"
	}
	if cfg.Skills.Dir == "" {
		cfg.Skills.Dir = "./skills"
	}
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = "openai"
	}

	// Set default checkpoint config
	// Only set defaults if keys were not explicitly set in config
	if !Viper().IsSet("agent.checkpoint.enabled") {
		cfg.Agent.Checkpoint.Enabled = true
	}
	if cfg.Agent.Checkpoint.StoreType == "" {
		cfg.Agent.Checkpoint.StoreType = "file"
	}
	if cfg.Agent.Checkpoint.Path == "" {
		cfg.Agent.Checkpoint.Path = "./data/checkpoints"
	}

	// Set default tracing config
	// Only set defaults if keys were not explicitly set in config
	if !Viper().IsSet("agent.tracing.enabled") {
		cfg.Agent.Tracing.Enabled = true
	}
	if !Viper().IsSet("agent.tracing.markdown.enabled") {
		cfg.Agent.Tracing.Markdown.Enabled = true
	}
	if cfg.Agent.Tracing.Markdown.OutputDir == "" {
		cfg.Agent.Tracing.Markdown.OutputDir = "./data/reports"
	}
	if cfg.Agent.Tracing.Log.Level == "" {
		cfg.Agent.Tracing.Log.Level = "standard"
	}

	return cfg, nil
}
