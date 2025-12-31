# Ops Skills

A collection of Agent Skills for DevOps and infrastructure management operations.

## Skills

### kubekey

KubeKey skill for managing Kubernetes clusters. Provides capabilities to:
- Check and install KubeKey tool
- Create Kubernetes clusters
- Scale clusters (add/remove nodes)
- View cluster configurations

See [skills/kubekey/README.md](skills/kubekey/README.md) for details.

## Installation

### For Claude Code

Copy the skill folder to your skills directory:

```bash
# Personal skills
cp -r skills/kubekey ~/.claude/skills/

# Project-specific skills
cp -r skills/kubekey .claude/skills/
```

### For OpenAI Codex CLI

```bash
cp -r skills/kubekey ~/.codex/skills/
```

## Publishing to SkillsMP

To publish skills to [SkillsMP](https://skillsmp.com/):

1. Ensure your skill has a `marketplace.json` file
2. Push your skill to a GitHub repository
3. SkillsMP will automatically discover and index your skill

## Resources

### Agent Skills

- [Agent Skills Specification](https://agentskills.io/)
- [Agent Skills Specification (GitHub)](https://github.com/anthropics/skills)
- [SkillsMP Marketplace](https://skillsmp.com/)
- [Agent Skills Examples](https://github.com/anthropics/skills/tree/main/examples)

### Skill Development

- [Creating Agent Skills](https://agentskills.io/specification)
- [SKILL.md Format](https://agentskills.io/specification)
- [Publishing Skills](https://skillsmp.com/)

### Related Tools

- [Claude Code](https://claude.ai/code)
- [OpenAI Codex CLI](https://openai.com/)
- [Cursor IDE](https://cursor.sh/)

## License

MIT
