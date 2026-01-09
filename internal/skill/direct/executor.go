package direct

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hb-chen/opskills/internal/skill"
)

// DirectExecutor executes skills directly by running their scripts
type DirectExecutor struct {
	runner *ScriptRunner
}

// NewDirectExecutor creates a new direct executor
func NewDirectExecutor(timeout time.Duration) *DirectExecutor {
	return &DirectExecutor{
		runner: NewScriptRunner(timeout),
	}
}

// Execute executes a skill with the given parameters
func (e *DirectExecutor) Execute(s *skill.Skill, params skill.ExecutionParams) (*skill.ExecutionResult, error) {
	startTime := time.Now()

	// Determine which script to run
	// For now, we'll look for a script matching the skill name or use a default
	scriptPath, err := e.findScript(s, params)
	if err != nil {
		return &skill.ExecutionResult{
			Success:   false,
			Error:     err.Error(),
			ExitCode:  -1,
			Duration:  time.Since(startTime),
			Timestamp: time.Now(),
		}, err
	}

	// Convert params to script arguments and environment variables
	args, env := e.prepareExecution(s, params)

	// Run the script
	stdout, stderr, exitCode, err := e.runner.Run(scriptPath, args, env)

	duration := time.Since(startTime)
	result := &skill.ExecutionResult{
		Success:   err == nil && exitCode == 0,
		Output:    stdout,
		Error:     stderr,
		ExitCode:  exitCode,
		Duration:  duration,
		Timestamp: time.Now(),
	}

	if err != nil {
		result.Error = fmt.Sprintf("%s: %s", err.Error(), stderr)
		return result, err
	}

	if exitCode != 0 {
		result.Error = stderr
		return result, fmt.Errorf("script exited with code %d: %s", exitCode, stderr)
	}

	return result, nil
}

// findScript finds the appropriate script to execute
func (e *DirectExecutor) findScript(s *skill.Skill, params skill.ExecutionParams) (string, error) {
	// Check if a specific script is requested
	if scriptName, ok := params["script"].(string); ok {
		scriptPath := filepath.Join(s.ScriptsPath, scriptName)
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
		return "", fmt.Errorf("script not found: %s", scriptPath)
	}

	// Check if action is specified (e.g., "create_cluster", "add_nodes")
	if action, ok := params["action"].(string); ok {
		scriptPath := filepath.Join(s.ScriptsPath, fmt.Sprintf("%s.sh", action))
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
	}

	// Default: look for a main script or the first available script
	scriptsDir := s.ScriptsPath
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read scripts directory: %w", err)
	}

	// Look for a main.sh or the first .sh file
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sh" {
			if entry.Name() == "main.sh" {
				return filepath.Join(scriptsDir, entry.Name()), nil
			}
		}
	}

	// Return the first .sh file found
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sh" {
			return filepath.Join(scriptsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("no script found in %s", scriptsDir)
}

// prepareExecution prepares arguments and environment variables for script execution
func (e *DirectExecutor) prepareExecution(s *skill.Skill, params skill.ExecutionParams) ([]string, map[string]string) {
	var args []string
	env := make(map[string]string)

	// Set skill-specific environment variables
	env["SKILL_NAME"] = s.Name
	env["SKILL_BASE_PATH"] = s.BasePath
	env["SKILL_SCRIPTS_PATH"] = s.ScriptsPath

	// Convert params to arguments and environment variables
	for key, value := range params {
		// Skip internal parameters
		if key == "script" || key == "action" {
			continue
		}

		// Convert value to string
		valueStr := fmt.Sprintf("%v", value)

		// Add as environment variable (prefixed with SKILL_PARAM_)
		env[fmt.Sprintf("SKILL_PARAM_%s", key)] = valueStr

		// Also add as argument if it's a simple value
		if len(valueStr) > 0 && valueStr[0] != '-' {
			args = append(args, fmt.Sprintf("--%s", key), valueStr)
		}
	}

	return args, env
}



