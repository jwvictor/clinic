# Clinic

**Your CLI tools, managed — installed, authenticated, agent-ready.**

Clinic manages collections of agent-friendly CLI tools as unified, opinionated stacks. One command turns a bare terminal into a fully agent-capable workspace — tools installed, authenticated, and visible to your AI agent.

```
clinic init --stack saas-founder
```

That installs `gh`, `vercel`, `stripe`, `supabase`, `firebase`, `sentry-cli`, `gws`, `jq`, and `ngrok` — then generates skill files so [OpenClaw](https://github.com/openclaw/openclaw), Claude Code, Gemini CLI, and other AI agents know how to use them.

## Why

Every week another company ships a CLI tool designed for AI agents. Google shipped `gws`. GitHub has `gh`. Stripe, Vercel, DigitalOcean, Fly.io — all have CLIs that AI coding agents invoke.

But setting up 5-10 CLI tools means:
- Installing from different package managers (brew, npm, apt)
- Authenticating each one separately
- Manually writing skill/config files so your agent knows they exist
- Keeping everything updated
- Repeating all of this on every new machine

Clinic handles all of it.

## Quick Start

```bash
curl -fsSL https://getclinic.sh/install | sh
```

Then pick a stack and go:

```bash
clinic init --stack saas-founder
```

Or add tools one at a time:

```bash
clinic add gh
clinic add stripe
```

## What It Does

### Installs CLI tools
Clinic delegates to native package managers — Homebrew for Go/Rust tools, npm for Node tools, official install scripts as fallback. It doesn't host binaries or reinvent package management.

### Authenticates them
`clinic auth <tool>` runs the tool's native auth flow. In headless environments (Docker, SSH, CI), it auto-detects and uses device-code flows — giving you a URL and code to complete auth on any device with a browser.

### Generates agent skill files
After installation, Clinic writes [Agent Skills standard](https://github.com/anthropics/claude-code) `SKILL.md` files to:

- `~/.openclaw/skills/` — [OpenClaw](https://github.com/openclaw/openclaw)
- `~/.claude/skills/` — Claude Code
- `~/.agents/skills/` — Agent Skills open standard (Gemini CLI, Codex CLI, etc.)

These tell your AI agent what tools are available, how to use them, and that they're authenticated. The agent can start using them immediately without any extra configuration.

### Health checks
```
$ clinic doctor

Tool             Version      Auth       Skill    Status
────             ───────      ────       ─────    ──────
gh               2.88.1       ✓ ok       ✓        ✓ ok
stripe           1.25.0       ✗ no       ✓        ⚠ auth needed
jq               1.8.1        n/a        ✓        ✓ ok
```

## Supported Tools & Stacks

Browse all available tools and curated stacks at [getclinic.sh/tools](https://getclinic.sh/tools).

Or from the CLI:

```bash
clinic list --all    # See all available tools
clinic stacks        # See available stacks
```

## Commands

```
clinic init [--stack <name>]    Set up a workspace with a curated stack
clinic add <tool>               Add a single tool
clinic remove <tool>            Remove a tool from your workspace
clinic list [--all]             List installed or available tools
clinic doctor                   Health check all tools
clinic auth <tool>              Authenticate a tool
clinic auth --status            Show auth status for all tools
clinic generate                 Regenerate skill files
clinic update [<tool>]          Update lockfile versions
clinic stacks                   Browse available stacks
clinic shellenv                 Print PATH setup for shell profile
clinic version                  Print version
```

## How It Works

Clinic is a single Go binary. It maintains a lockfile at `~/.clinic/clinic.json` tracking what's installed and how. Tools are installed globally via their native package managers — Clinic orchestrates but doesn't own the installations.

Skill files are the same format everywhere — the [Agent Skills open standard](https://github.com/anthropics/claude-code) with YAML frontmatter. One `SKILL.md` works with OpenClaw, Claude Code, Gemini CLI, Codex CLI, and any other compliant agent.

### Existing tools

If you already have a tool installed, `clinic add` detects it, skips installation, and generates the skill files. It works regardless of how the tool was originally installed.

### Headless auth

In environments without a browser (Docker, SSH, CI), `clinic auth` auto-detects and uses device-code or no-localhost flows. You get a URL and code — open it on your phone or laptop to complete auth.

## Development

```bash
# Build
make build

# Run tests in a Docker container (clean Linux environment)
make test-build     # Build the test container
make test-shell     # Interactive shell in a fresh container
make test-smoke     # Quick smoke test
make test-run CMD="add jq"  # Run a specific command
```

## License

Apache 2.0
