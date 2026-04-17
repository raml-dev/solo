<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

# AI-Assisted Development Workflow

We suggest using [`pi`](https://github.com/badlogic/pi-mono/tree/main/packages/coding-agent) as its AI coding agent. The project is already set up to wire in two MCP documentation servers:

- [Svelte/SvelteKit](https://svelte.dev/docs/cli/mcp)
- [Flowbite-Svelte](https://flowbite-svelte.com/docs/mcp/overview)

Both MCP servers are configured to run locally. There is also the configuration to wire them through VSCode, if you don't want to use `pi`.

Everything is managed through [mise](https://mise.jdx.dev/), which handles tool versions, tasks, and ensures pi always runs from a predictable root regardless of which subdirectory you're working in.

This document explains why the setup exists, how to get it running, and how to use it day to day.

---

## Why this setup

The project is structured as a [Wails](https://wails.io/) application, which creates a monorepo with a Go backend and a Svelte frontend living under the same root. Because the two codebases are in separate subdirectories, running a coding agent from inside `frontend/` would cause it to lose sight of project-level configuration. The `mise run pi` task pins the working directory to `{{ config_root }}` (the directory containing `mise.toml`) regardless of where you are in the tree when you invoke it.

The `pi` shell alias defined in `mise.toml` means you can type `pi` anywhere in the project and get the same, consistent behaviour.

The MCP servers give the agent live access to documentation to reduce risks of hallucinations:

- **Svelte MCP**: live Svelte 5 and SvelteKit docs, plus a static-analysis linter (`svelte-autofixer`) that catches rune-usage mistakes before they reach your editor.
- **Flowbite-Svelte MCP**: component reference, props tables, and usage examples for every Flowbite-Svelte component.

Both are proxied through [pi-mcp-adapter](https://github.com/nicobailon/pi-mcp-adapter), which exposes them as a single lightweight `mcp` tool (~200 tokens) rather than injecting hundreds of tool definitions into every prompt.

---

## Prerequisites

You need [mise](https://mise.jdx.dev/) installed and activated in your shell. Everything else is managed by mise tasks.

Read their installation instructions - for example, you might prefer to use a package manager (if mise is available there). A quick way to install mise regardless of system packages is:

```bash
# macOS / Linux
curl https://mise.run | sh

# Follow the printed instructions to add mise to your shell profile, then:
mise --version   # should print a recent version
```

---

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

---

## Project structure (AI-relevant files)

```
<project-root>/
├── mise.toml                          # tool versions, tasks, pi alias
├── .pi/
│   └── mcp.json                       # MCP server configuration for pi-mcp-adapter
├── .agents/
│   └── frontend-skill/
│       └── flowbite-svelte-skill.md   # pi skill: Svelte + Flowbite-Svelte workflow rules
└── .mcp/
    └── flowbite-svelte-mcp/           # cloned & built by mise run mcp:flowbite (gitignored)
```

`.pi/mcp.json` configures the two MCP servers for this project. `pi-mcp-adapter` reads project-level config from `.pi/mcp.json` automatically when pi starts from the project root, which the mise alias guarantees.

`.agents/frontend-skill/flowbite-svelte-skill.md` is a pi skill that instructs the agent how and when to call the MCP tools, what patterns to follow for Svelte 5, and how to combine Svelte and Flowbite-Svelte correctly. Pi loads skills from `.agents/` automatically.

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

## Using the agent for frontend work

The `flowbite-svelte-skill` is loaded automatically. It instructs the agent to follow this workflow whenever you ask it to build or modify frontend components:

1. **Consult Svelte docs first**: the agent calls `list-sections` on the Svelte MCP to discover relevant documentation, then `get-documentation` to fetch it.
2. **Find the right Flowbite-Svelte component**: the agent calls `findComponent` with a natural-language query (e.g. `"modal"`, `"data table"`, `"form select"`), then `getComponentDoc` to retrieve the full component reference.
3. **Write the code**: combining Svelte 5 runes with Flowbite-Svelte components.
4. **Validate**: the agent calls `svelte-autofixer` on every `.svelte` file it touches and iterates until there are zero issues.

You do not need to explicitly invoke any of these steps. Describing what you want is sufficient:

```plaintext
Add a confirmation modal to the delete button in UserTable.svelte.
The modal should show the username and ask for confirmation before
calling the deleteUser action.
```

The agent will look up both the Svelte modal patterns and the Flowbite-Svelte `Modal` component docs before writing anything.

---

## MCP tools reference

The agent accesses MCP tools through the `mcp()` proxy. You do not call these directly — the agent does — but knowing what exists helps you write better prompts.

### Svelte MCP

| Tool                | What it does                                                                                                         |
| ------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `list-sections`     | Returns all available documentation sections with titles and use-case descriptions                                   |
| `get-documentation` | Fetches full documentation for one or more sections by path                                                          |
| `svelte-autofixer`  | Analyses Svelte code and returns issues and suggestions; runs in a loop until clean                                  |
| `playground-link`   | Generates a Svelte Playground URL from code (only for exploratory snippets, never for code already written to files) |

### Flowbite-Svelte MCP

| Tool               | What it does                                                              |
| ------------------ | ------------------------------------------------------------------------- |
| `findComponent`    | Searches components by name or category; returns the doc path             |
| `getComponentDoc`  | Fetches full component documentation including props, slots, and examples |
| `getComponentList` | Lists all available components with their categories                      |
| `searchDocs`       | Full-text search across all Flowbite-Svelte documentation                 |

---

## Adding or updating skills

Skills live in `.agents/`. Each skill is a Markdown file the agent loads at session start. To modify the Svelte/Flowbite-Svelte workflow rules, edit `.agents/frontend-skill/flowbite-svelte-skill.md` directly and restart pi. Changes take effect immediately on the next session.

To add a new skill (for example, for the Go backend), create a new directory under `.agents/` and add a Markdown file. Pi discovers all `.md` files under `.agents/` automatically.

## General suggestions

**NEVER** ask AI agents to implement changes right away, especially if you want to make complex changes.

Always ask for an implementation *plan* first. Make the agent write and update the plan to a markdown file, so you can use it on subsequent prompts and start new sessions without re-planning all over again. Be specific - end your prompt with something like "do not implement anything, just write the plan". Reiterate on the plan multiple times, asking the agent to improve specific sections when needed or to add new parts.

Be broad on high level plans. When you have a good overall idea, ask the agent to create sub-plans with deeper details, one at a time, using the same planning and iteration strategies. Also persist those in dedicated markdown files.

When you are satisfied with the planning, ask the agent to start implementing one sub-plan at a time. If your plan involves multiple components, ask it to first show an high level overview of the changes without writing anything on files, so you can check.

Be aware of context exhaustion. `pi` already does automatic compactions, but it's always better to write a specific prompt for them (or start an entirely new session if you already persisted your plans). Models with almost-full contexts can be less effective.
