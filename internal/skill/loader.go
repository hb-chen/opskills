package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hb-chen/opskills/pkg/logger"
)

// Loader loads skills from a directory
type Loader struct {
	skillsDir string
}

// NewLoader creates a new skill loader
func NewLoader(skillsDir string) *Loader {
	return &Loader{
		skillsDir: skillsDir,
	}
}

// LoadAll loads all skills from the skills directory
func (l *Loader) LoadAll() ([]*Skill, error) {
	var skills []*Skill

	// Check if skills directory exists
	if _, err := os.Stat(l.skillsDir); os.IsNotExist(err) {
		return skills, fmt.Errorf("skills directory does not exist: %s", l.skillsDir)
	}

	// Walk through the skills directory
	err := filepath.Walk(l.skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for SKILL.md files
		if info.Name() == "SKILL.md" {
			skill, err := ParseSKILL(path)
			if err != nil {
				// Log error but continue loading other skills
				logger.Warnf("Failed to load skill from %s: %v", path, err)
				return nil
			}
			skills = append(skills, skill)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk skills directory: %w", err)
	}

	return skills, nil
}

// LoadSkill loads a specific skill by name
func (l *Loader) LoadSkill(name string) (*Skill, error) {
	skillPath := filepath.Join(l.skillsDir, name, "SKILL.md")

	// Check if file exists
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("skill not found: %s", name)
	}

	return ParseSKILL(skillPath)
}
