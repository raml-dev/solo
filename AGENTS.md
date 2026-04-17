<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->
# AGENTS.md

This file defines directives for AI coding agents working in the Solo repository.
Solo is a macOS desktop application built with [Wails](https://wails.io/) (Go backend + Svelte 5 + Tailwind + Flowbite frontend).

> **Agent compatibility**: These directives are written for **Codex / OpenAI** agents. All rules are mandatory constraints, not suggestions. When a rule conflicts with completing a task efficiently, follow the rule and surface the conflict to the user.

---

## Stack Overview

| Layer | Technology |
|---|---|
| Desktop shell | Wails v2 |
| Backend | Go |
| Frontend | Svelte 5, Tailwind CSS, Flowbite |
| IPC | Wails `bind()` bridge (Go → JS) |

---

## Skill Files

This repo uses two skill files that are the authoritative source of truth for code style, patterns, and MCP usage. **Read the relevant skill file before starting any task.**

| Skill | Path | When to use |
|---|---|---|
| Frontend | `docs/agents/FRONTEND_SKILL.md` | Any work touching `frontend/` |
| Backend | `docs/agents/BACKEND_SKILL.md` | Any work touching `.go` files |

If a task spans both frontend and backend, read both skill files before starting.

Skill files define *how* to write code. This file defines *what is allowed*. When they conflict, raise the conflict with the user before proceeding.

---

## Understanding the Architecture

Solo is a Wails app: the Go backend and the Svelte frontend run as a single desktop process, communicating over an IPC bridge. **Agents cannot launch or interact with the desktop UI.** Do not attempt to run `wails dev` — it spawns a desktop window that is not accessible in an agent context.

### How to verify your work without running the app

| Concern | Command |
|---|---|
| Backend compiles | `go build ./...` |
| Backend tests pass | `go test ./...` |
| Frontend type-checks | `npm run check` (inside `frontend/`) |
| Frontend lints | `npm run lint` (inside `frontend/`) |
| Frontend builds | `npm run build` (inside `frontend/`) |
| Wails bindings regenerated | `wails build` (build only, does not open the app) |

---

## Structural Changes — Ask Before Proceeding

The following actions require explicit approval from the user before being executed. Stop, describe what you intend to do and why, and wait for confirmation.

- **Adding or removing files or directories** anywhere in the repository
- **Changing the Wails `bind()` surface** — any modification to which Go methods or structs are exposed to the frontend
- **Modifying Svelte component props or event contracts** — changing a component's public API (props, emitted events, slot signatures)
- **Changing Go package structure or module layout** — moving packages, renaming them, splitting or merging them, or altering `go.mod`

When in doubt about whether something qualifies, ask.

---

## Hard Rules — Never Violate These

- **Never commit directly to `main`.** All changes must go through a branch and pull request.
- **Never modify generated or auto-synced files.** This includes anything produced by Wails' own codegen (e.g. `frontend/wailsjs/`). Identify these by their headers or ask if unsure.
- **Never add dependencies without confirmation.** This applies to both Go modules and npm packages. Explain what the dependency is, why it is needed, and whether an alternative using existing deps or the standard library is feasible. Wait for explicit confirmation before running `go get` or `npm install`.
- **Never touch app icon or branding assets.** Files under `build/appicon*`, `frontend/public/*icon*`, or any asset explicitly used for the app icon are off-limits.

---

## MCPs

MCPs are only permitted as defined in the skill files. Do not invoke MCP tools or servers that are not documented in the relevant skill file for the task at hand.

---

## What Agents Can Do Freely

- Fixing bugs within existing functions without changing their signatures
- Writing or updating tests
- Improving comments and documentation
- Refactoring internals within a single file without changing that file's public API
- Updating Tailwind classes or minor UI tweaks within an existing component
