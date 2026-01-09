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

// Config represents the application configuration
type Config struct {
	Server Server `mapstructure:"server" yaml:"server"`
	Log    Log    `mapstructure:"log" yaml:"log"`
	LLM    LLM    `mapstructure:"llm" yaml:"llm"`
	Skills Skills `mapstructure:"skills" yaml:"skills"`
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

	return cfg, nil
}

