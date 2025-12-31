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

- [Agent Skills Specification](https://agentskills.io/)
- [SkillsMP Marketplace](https://skillsmp.com/)
- [Agent Skills on GitHub](https://github.com/anthropics/skills)

## License

MIT
