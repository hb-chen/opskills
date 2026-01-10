package skill

import "time"

// Skill represents a skill definition
type Skill struct {
	// Metadata from SKILL.md frontmatter
	Name          string
	Description   string
	License       string
	Compatibility string

	// Path information
	BasePath    string // Path to skill directory (e.g., skills/kubekey)
	ScriptsPath string // Path to scripts directory (e.g., skills/kubekey/scripts)
	SKILLPath   string // Path to SKILL.md file

	// Content
	Instructions string // Full content of SKILL.md (after frontmatter)

	// Metadata
	LoadedAt time.Time
}

// ExecutionResult represents the result of executing a skill
type ExecutionResult struct {
	Success   bool
	Output    string
	Error     string
	ExitCode  int
	Duration  time.Duration
	Timestamp time.Time
}

// ExecutionParams represents parameters for skill execution
type ExecutionParams map[string]interface{}

// Executor defines the interface for executing skills
type Executor interface {
	Execute(skill *Skill, params ExecutionParams) (*ExecutionResult, error)
}
