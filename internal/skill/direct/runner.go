package direct

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ScriptRunner runs bash scripts
type ScriptRunner struct {
	timeout time.Duration
}

// NewScriptRunner creates a new script runner
func NewScriptRunner(timeout time.Duration) *ScriptRunner {
	return &ScriptRunner{
		timeout: timeout,
	}
}

// Run executes a bash script with the given arguments
func (r *ScriptRunner) Run(scriptPath string, args []string, env map[string]string) (string, string, int, error) {
	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return "", "", -1, fmt.Errorf("script not found: %s", scriptPath)
	}

	// Make script executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return "", "", -1, fmt.Errorf("failed to make script executable: %w", err)
	}

	// Create command
	cmd := exec.Command("bash", append([]string{scriptPath}, args...)...)

	// Set working directory to script's directory
	cmd.Dir = filepath.Dir(scriptPath)

	// Set environment variables
	if env != nil {
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start command
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		return "", "", -1, fmt.Errorf("failed to start script: %w", err)
	}

	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		_ = time.Since(startTime) // duration captured but not used in this simplified version
		exitCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			} else {
				return stdout.String(), stderr.String(), -1, fmt.Errorf("script execution error: %w", err)
			}
		}

		return stdout.String(), stderr.String(), exitCode, nil
	case <-time.After(r.timeout):
		// Kill the process on timeout
		cmd.Process.Kill()
		return "", "", -1, fmt.Errorf("script execution timeout after %v", r.timeout)
	}
}

