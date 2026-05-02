<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

# AI-Assisted Development

We are strong supporters of AI-assisted development, but we want to use it the right way.

When contributing to this project, you can use AI coding agents to help you with the development. The important thing to remember is that AI coding agents are not always reliable and they can make mistakes: you should review the changes they make and make sure they are correct. More importantly, you should **always** understand the changes they make and **be able to explain them**.

## Our workflow

The following is a description of how _we_ use AI coding agents to develop this project. It is not a requirement for contributing, but it is a good starting point. You can use whatever coding agent, skill, workflow, or prompt you want.

Our setup uses the [`pi`](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent) AI coding agent. The project is already set up to install it and wire two MCP documentation servers:

- [Svelte/SvelteKit](https://svelte.dev/docs/cli/mcp)
- [Flowbite-Svelte](https://flowbite-svelte.com/docs/mcp/overview)

Both MCP servers are configured to run locally. There is also the configuration to wire them through VSCode, if you don't want to use `pi`.

Everything is managed through [mise](https://mise.jdx.dev/), which handles tool versions, tasks, and ensures `pi` always runs from a predictable root regardless of which subdirectory you're launching it from.

This document explains why the setup exists, how to get it running, and how to use it day to day.

## General Workflow

- Read the available skills:

  - [Frontend](../.agents/skills/solo-frontend-skill/SKILL.md)
  - [Backend](../.agents/skills/solo-backend-skill/SKILL.md)

  They provide useful guidelines even if you don't want to use an AI agent.

- We **NEVER** ask AI agents to implement changes right away, especially if for complex changes.

  Always ask for a _plan_ first. Make the agent write and update the plan to a markdown file, so you can use it on subsequent prompts and start new sessions without re-planning all over again. Be specific - if the agent has a plan mode use it, otherwise end the prompt with something like "do not implement anything, just write the plan".

  It's recommended to persist plans in the `agents` folder. Contents inside that folder are gitignored.

- We always apply _our_ critical thinking. We ask for specific features and reiterate on the plan multiple times, asking the agent to improve specific sections when needed or to add new parts. We deeply analyze the plan, ask the agent to explain it if there is something we don't fully understand, and make it change the approach if the one it proposes does not follow our intention.

- Create and iterate on high level plans first, then ask the agent to create sub-plans with deeper details and specific implementation tasks, one at a time, using the same planning and iteration strategies. We also persist those in dedicated markdown files.

- Once the planning is done, it's time to ask the agent to start implementing one sub-plan at a time. If the plan involves multiple components, we usually ask the agent to first show an high level overview of the changes without writing anything on files. We ensure that the agent stops at each task to review changes incrementally.

- Keep track of context utilization. Your coding agent of choice probably already does automatic compactions, but it's always better to write a specific prompt for them (or start an entirely new session if you already persisted your plans). Models with almost-full contexts can be less effective.

## Why this setup

The project is structured as a [Wails](https://wails.io/) application, which creates a monorepo with a Go backend and a Svelte frontend living under the same root. Because the two codebases are in separate subdirectories, running a coding agent from inside `frontend/` would cause it to lose sight of project-level configuration. The `mise run pi` task pins the working directory to `{{ config_root }}` (the directory containing `mise.toml`) regardless of where you are in the tree when you invoke it.

The `pi` shell alias defined in `mise.toml` means you can type `pi` anywhere in the project and get the same, consistent behaviour.

The MCP servers give the agent live access to documentation to reduce risks of hallucinations:

- **Svelte MCP**: live Svelte 5 and SvelteKit docs, plus a static-analysis linter (`svelte-autofixer`) that catches rune-usage mistakes before they reach your editor.
- **Flowbite-Svelte MCP**: component reference, props tables, and usage examples for every Flowbite-Svelte component.

Both are proxied through [pi-mcp-adapter](https://github.com/nicobailon/pi-mcp-adapter), which exposes them as a single lightweight `mcp` tool (~200 tokens) rather than injecting hundreds of tool definitions into every prompt. We also provide the setup for VSCode. You can always configure your agents the way you want, but please do not commit specific agents configuration folders.

## Prerequisites

You need [mise](https://mise.jdx.dev/) installed and activated in your shell. Everything else is managed by mise tasks.

Read their installation instructions - for example, you might prefer to use a package manager (if mise is available there). A quick way to install mise regardless of system packages is:

```bash
# macOS / Linux
curl https://mise.run | sh

# Follow the printed instructions to add mise to your shell profile, then:
mise --version   # should print a recent version
```

## One-time setup

From the project root, run:

```bash
mise install
```

This installs Go, Node, pnpm, wails and pi-coding-agent at the pinned versions.

That's it. There is also a postinstall hook that detects if both `pi-mcp-adapter` and the Flowbite MCP server are installed, and installs them if they're not.

### What the tasks do

`mise run mcp:check` runs once automatically after `mise install` and is just a wrapper for two subtasks:

1. `mcp:flowbite-check` looks for the `.mcp/flowbite-svelte-mcp/` folder inside the project root. If not present, installs the Flowbite MCP server
2. `mcp:adapter-check` checks if the MCP extension for pi is installed. If not, runs `pi install npm:pi-mcp-adapter` to wire in MCP support.

## Project structure (AI-relevant files)

```plaintext
<project-root>/
├── mise.toml                          # tool versions, tasks, pi alias
├── .pi/
│   └── mcp.json                       # MCP server configuration for pi-mcp-adapter
├── .agents/skills
│   ├── solo-backend-skill/
│   │   └── SKILL.md   # skill to go development
│   └── solo-frontend-skill/
│       └── SKILL.md   # skill for UI development
└── .mcp/
    └── flowbite-svelte-mcp/           # cloned & built by mise run mcp:flowbite (gitignored)
```

`.pi/mcp.json` configures the two MCP servers for this project. `pi-mcp-adapter` reads project-level config from `.pi/mcp.json` automatically when pi starts from the project root, which the mise alias guarantees.

`.agents/solo-frontend-skill/SKILL.md` is a skill that instructs the agent how and when to call the MCP tools, what patterns to follow for Svelte 5, and how to combine Svelte and Flowbite-Svelte correctly. Most coding agents load skills from `.agents/` automatically.

---

## Running the agent

```bash
# From anywhere in the project tree
pi
# or equivalently:
mise run pi
```

Both expand to the same thing: `pi` launched with `{{ config_root }}` as the working directory.

The agent starts with both MCP servers available but not yet connected: they connect on demand the first time a tool is called. You will see this reflected in the pi session output.

---

## Adding or updating skills

Skills live in `.agents/skills` and follow the [Agent Skills specification](https://agentskills.io/home). The agent loads skills at session start.

To add a new skill, create a new directory under `.agents/` named as the skill you want to add, and add a `SKILL.md` Markdown file and any optional references or scripts. Agents discovers all `SKILL.md` files under `.agents/` automatically.
