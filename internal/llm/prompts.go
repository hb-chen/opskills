package llm

import (
	"fmt"
	"strings"
)

const (
	// PlanningPromptTemplate is the template for planning prompts
	PlanningPromptTemplate = `You are an intelligent operations agent. Your task is to analyze the user's request and create an execution plan using available skills.

Available Skills:
{{range .Skills}}
- {{.Name}}: {{.Description}}
{{end}}

User Request: {{.Query}}

Please create a step-by-step execution plan. For each step, specify:
1. The skill name to use
2. The action to perform
3. A description of what will be done
4. Any required parameters

Format your response as a JSON object with the following structure:
{
  "steps": [
    {
      "id": 1,
      "skill_name": "skill_name",
      "action": "action_name",
      "description": "description",
      "params": {
        "param1": "value1"
      }
    }
  ]
}

Response:`

	// ExecutionPromptTemplate is the template for execution prompts
	ExecutionPromptTemplate = `You are executing a step in an operations plan.

Step: {{.StepDescription}}
Skill: {{.SkillName}}
Action: {{.Action}}
Parameters: {{.Params}}

Execute this step and provide a summary of the results.`

	// ErrorHandlingPromptTemplate is the template for error handling
	ErrorHandlingPromptTemplate = `An error occurred during execution:

Step: {{.StepDescription}}
Error: {{.Error}}

Please analyze the error and suggest:
1. What went wrong
2. How to fix it
3. Whether to retry or skip this step

Response:`
)

// PlanningPromptData holds data for planning prompt
type PlanningPromptData struct {
	Skills []SkillInfo
	Query  string
}

// SkillInfo holds skill information for prompts
type SkillInfo struct {
	Name        string
	Description string
}

// ExecutionPromptData holds data for execution prompt
type ExecutionPromptData struct {
	StepDescription string
	SkillName        string
	Action          string
	Params          string
}

// ErrorHandlingPromptData holds data for error handling prompt
type ErrorHandlingPromptData struct {
	StepDescription string
	Error           string
}

// FormatPlanningPrompt formats the planning prompt with data
func FormatPlanningPrompt(data PlanningPromptData) string {
	// Simple template replacement (in production, use a proper template engine)
	prompt := PlanningPromptTemplate
	prompt = replaceAll(prompt, "{{.Query}}", data.Query)
	
	// Build skills list
	skillsList := ""
	for _, skill := range data.Skills {
		skillsList += fmt.Sprintf("- %s: %s\n", skill.Name, skill.Description)
	}
	prompt = replaceAll(prompt, "{{range .Skills}}\n- {{.Name}}: {{.Description}}\n{{end}}", skillsList)
	
	return prompt
}

// FormatExecutionPrompt formats the execution prompt with data
func FormatExecutionPrompt(data ExecutionPromptData) string {
	prompt := ExecutionPromptTemplate
	prompt = replaceAll(prompt, "{{.StepDescription}}", data.StepDescription)
	prompt = replaceAll(prompt, "{{.SkillName}}", data.SkillName)
	prompt = replaceAll(prompt, "{{.Action}}", data.Action)
	prompt = replaceAll(prompt, "{{.Params}}", data.Params)
	return prompt
}

// FormatErrorHandlingPrompt formats the error handling prompt with data
func FormatErrorHandlingPrompt(data ErrorHandlingPromptData) string {
	prompt := ErrorHandlingPromptTemplate
	prompt = replaceAll(prompt, "{{.StepDescription}}", data.StepDescription)
	prompt = replaceAll(prompt, "{{.Error}}", data.Error)
	return prompt
}

// replaceAll replaces all occurrences of old with new in s
func replaceAll(s, old, new string) string {
	// Simple implementation - in production use strings.ReplaceAll
	result := s
	for {
		newResult := strings.Replace(result, old, new, -1)
		if newResult == result {
			break
		}
		result = newResult
	}
	return result
}

