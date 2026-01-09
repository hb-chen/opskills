package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SkillMetadata represents the YAML frontmatter in SKILL.md
type SkillMetadata struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	License       string `yaml:"license"`
	Compatibility string `yaml:"compatibility"`
}

// ParseSKILL parses a SKILL.md file and extracts metadata and content
func ParseSKILL(skillPath string) (*Skill, error) {
	// Read the file
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SKILL.md: %w", err)
	}

	content := string(data)

	// Extract frontmatter (between --- markers)
	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var metadata SkillMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter YAML: %w", err)
	}

	// Get base directory
	basePath := filepath.Dir(skillPath)
	scriptsPath := filepath.Join(basePath, "scripts")

	// Create Skill object
	skill := &Skill{
		Name:          metadata.Name,
		Description:   metadata.Description,
		License:       metadata.License,
		Compatibility: metadata.Compatibility,
		BasePath:      basePath,
		ScriptsPath:   scriptsPath,
		SKILLPath:     skillPath,
		Instructions:  strings.TrimSpace(body),
		LoadedAt:      time.Now(),
	}

	return skill, nil
}

// extractFrontmatter extracts YAML frontmatter from markdown content
// Returns frontmatter, body, and error
func extractFrontmatter(content string) (string, string, error) {
	// Check if content starts with ---
	if !strings.HasPrefix(content, "---") {
		return "", content, fmt.Errorf("SKILL.md must start with YAML frontmatter (---)")
	}

	// Find the second ---
	lines := strings.Split(content, "\n")
	var frontmatterLines []string
	var bodyStart int

	// First line should be ---
	if len(lines) == 0 || lines[0] != "---" {
		return "", content, fmt.Errorf("invalid frontmatter format: first line must be ---")
	}

	frontmatterLines = append(frontmatterLines, lines[0])
	bodyStart = 1

	// Find closing ---
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			bodyStart = i + 1
			break
		}
		frontmatterLines = append(frontmatterLines, lines[i])
	}

	if bodyStart == 1 {
		return "", content, fmt.Errorf("invalid frontmatter format: closing --- not found")
	}

	frontmatter := strings.Join(frontmatterLines[1:], "\n") // Skip first ---
	body := strings.Join(lines[bodyStart:], "\n")

	return frontmatter, body, nil
}

