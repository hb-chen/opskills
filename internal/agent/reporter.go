package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hb-chen/opskills/internal/state"
)

// MarkdownReporter generates Markdown reports from execution state
type MarkdownReporter struct {
	outputDir string
	enabled   bool
}

// NewMarkdownReporter creates a new Markdown reporter
func NewMarkdownReporter(outputDir string, enabled bool) *MarkdownReporter {
	return &MarkdownReporter{
		outputDir: outputDir,
		enabled:   enabled,
	}
}

// GenerateReport generates a Markdown report from the execution state
func (r *MarkdownReporter) GenerateReport(taskID string, finalState *state.State) error {
	if !r.enabled {
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate report content
	report := r.buildReport(taskID, finalState)

	// Generate filename: {timestamp}-{taskID}.md
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.md", timestamp, taskID)
	filepath := filepath.Join(r.outputDir, filename)

	// Write report to file
	if err := os.WriteFile(filepath, []byte(report), 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// buildReport builds the Markdown report content
func (r *MarkdownReporter) buildReport(taskID string, s *state.State) string {
	var b strings.Builder

	// Title
	b.WriteString("# Task Execution Report\n\n")

	// Task Overview
	b.WriteString("## Task Overview\n\n")
	b.WriteString(fmt.Sprintf("- **Task ID**: `%s`\n", taskID))
	b.WriteString(fmt.Sprintf("- **Query**: %s\n", s.Query))
	b.WriteString(fmt.Sprintf("- **Started At**: %s\n", s.StartedAt))
	b.WriteString(fmt.Sprintf("- **Updated At**: %s\n", s.UpdatedAt))
	if s.Error != "" {
		b.WriteString("- **Status**: ❌ Failed\n")
		fmt.Fprintf(&b, "- **Error**: %s\n", s.Error)
	} else if s.FinalResult != nil && s.FinalResult.Success {
		b.WriteString("- **Status**: ✅ Success\n")
	} else {
		b.WriteString("- **Status**: ⏳ In Progress\n")
	}
	b.WriteString("\n")

	// Planning Phase
	b.WriteString("## Planning Phase\n\n")
	if s.PlanError != "" {
		b.WriteString(fmt.Sprintf("❌ **Planning Failed**: %s\n\n", s.PlanError))
	} else if s.Plan != nil {
		b.WriteString("✅ **Plan Generated Successfully**\n\n")
		fmt.Fprintf(&b, "**Total Steps**: %d\n\n", len(s.Plan.Steps))
		
		if len(s.Plan.Steps) > 0 {
			b.WriteString("### Plan Steps\n\n")
			for i, step := range s.Plan.Steps {
				b.WriteString(fmt.Sprintf("#### Step %d: %s\n\n", step.ID, step.Description))
				b.WriteString(fmt.Sprintf("- **Skill**: `%s`\n", step.SkillName))
				b.WriteString(fmt.Sprintf("- **Action**: `%s`\n", step.Action))
				if len(step.Params) > 0 {
					b.WriteString("- **Parameters**:\n")
					for k, v := range step.Params {
						b.WriteString(fmt.Sprintf("  - `%s`: `%v`\n", k, v))
					}
				}
				b.WriteString("\n")
				if i < len(s.Plan.Steps)-1 {
					b.WriteString("---\n\n")
				}
			}
		}
	} else {
		b.WriteString("⏳ **Planning not yet started**\n\n")
	}
	b.WriteString("\n")

	// Execution Phase
	b.WriteString("## Execution Phase\n\n")
	if len(s.Steps) == 0 {
		b.WriteString("⏳ **No steps to execute**\n\n")
	} else {
		b.WriteString(fmt.Sprintf("**Total Steps**: %d\n\n", len(s.Steps)))
		
		// Step execution details
		for i, step := range s.Steps {
			b.WriteString(fmt.Sprintf("### Step %d: %s\n\n", step.ID, step.Description))
			b.WriteString(fmt.Sprintf("- **Skill**: `%s`\n", step.SkillName))
			b.WriteString(fmt.Sprintf("- **Action**: `%s`\n", step.Action))
			b.WriteString(fmt.Sprintf("- **Status**: %s\n", r.formatStatus(step.Status)))
			
			// Find corresponding result
			var stepResult *state.StepResult
			for _, result := range s.Results {
				if result.StepID == step.ID {
					stepResult = result
					break
				}
			}
			
			if stepResult != nil {
				if stepResult.Duration != "" {
					b.WriteString(fmt.Sprintf("- **Duration**: %s\n", stepResult.Duration))
				}
				if stepResult.Success {
					b.WriteString("- **Result**: ✅ Success\n")
				} else {
					b.WriteString("- **Result**: ❌ Failed\n")
				}
				if stepResult.Output != "" {
					b.WriteString(fmt.Sprintf("- **Output**:\n```\n%s\n```\n", stepResult.Output))
				}
				if stepResult.Error != "" {
					b.WriteString(fmt.Sprintf("- **Error**:\n```\n%s\n```\n", stepResult.Error))
				}
			}
			
			b.WriteString("\n")
			if i < len(s.Steps)-1 {
				b.WriteString("---\n\n")
			}
		}
	}
	b.WriteString("\n")

	// Validation Phase
	b.WriteString("## Validation Phase\n\n")
	if s.FinalResult != nil {
		if s.FinalResult.Success {
			b.WriteString("✅ **Validation Passed**\n\n")
		} else {
			b.WriteString("❌ **Validation Failed**\n\n")
		}
		if s.FinalResult.Summary != "" {
			b.WriteString(fmt.Sprintf("**Summary**: %s\n\n", s.FinalResult.Summary))
		}
		if s.FinalResult.Output != "" {
			b.WriteString(fmt.Sprintf("**Output**:\n```\n%s\n```\n\n", s.FinalResult.Output))
		}
		if s.FinalResult.Error != "" {
			b.WriteString(fmt.Sprintf("**Error**:\n```\n%s\n```\n\n", s.FinalResult.Error))
		}
	} else {
		b.WriteString("⏳ **Validation not yet completed**\n\n")
	}
	b.WriteString("\n")

	// Timeline
	b.WriteString("## Timeline\n\n")
	b.WriteString("| Event | Timestamp |\n")
	b.WriteString("|-------|----------|\n")
	if s.StartedAt != "" {
		b.WriteString(fmt.Sprintf("| Task Started | %s |\n", s.StartedAt))
	}
	if s.UpdatedAt != "" && s.UpdatedAt != s.StartedAt {
		b.WriteString(fmt.Sprintf("| Last Updated | %s |\n", s.UpdatedAt))
	}
	b.WriteString("\n")

	// Error Details (if any)
	if s.Error != "" || (s.FinalResult != nil && s.FinalResult.Error != "") {
		b.WriteString("## Error Details\n\n")
		if s.Error != "" {
			b.WriteString(fmt.Sprintf("### Task Error\n\n```\n%s\n```\n\n", s.Error))
		}
		if s.FinalResult != nil && s.FinalResult.Error != "" {
			b.WriteString(fmt.Sprintf("### Final Result Error\n\n```\n%s\n```\n\n", s.FinalResult.Error))
		}
	}

	// Footer
	b.WriteString("---\n\n")
	b.WriteString(fmt.Sprintf("*Report generated at %s*\n", time.Now().Format(time.RFC3339)))

	return b.String()
}

// formatStatus formats the status with emoji
func (r *MarkdownReporter) formatStatus(status string) string {
	switch status {
	case "pending":
		return "⏳ Pending"
	case "running":
		return "🔄 Running"
	case "completed":
		return "✅ Completed"
	case "failed":
		return "❌ Failed"
	default:
		return status
	}
}

