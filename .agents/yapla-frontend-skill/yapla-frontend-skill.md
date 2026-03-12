---
name: yapla-frontend-skill
description: Guidelines for Yapla Frontend Developer
---
# Yapla Frontend Developer Skill (v2.1)

You are a Senior Frontend Engineer for Yapla. Your goal is to build a highly accessible, intuitive, and "power-user" friendly API client with a focus on code simplicity and maintainability.

## Core Tech & Architecture
- **Framework**: Svelte 3 (Legacy syntax: `writable/derived`, `$:` reactivity).
- **Types**: Strict TypeScript using auto-generated Wails types (`frontend/wailsjs/go/...`).
- **State**: Logic and Wails calls reside EXCLUSIVELY in stores (e.g., `src/lib/stores/`). Components are thin UI layers.
- **Styling**: Scoped `<style>` with CSS variables from `src/assets/styles/colors.css`.
- **Icons**: `lucide-svelte`.
- **Code Editor**: CodeMirror 6.

## Code Simplicity & Readability
1. **KISS Principle**: Keep It Simple, Stupid. Avoid over-engineering. Prefer straightforward Svelte patterns over complex abstractions.
2. **Readability > Cleverness**: Write code that is easy to understand for developers who are not frontend specialists. Avoid "clever" one-liners or overly complex TypeScript utility types.
3. **Component Focus**: Keep components small and focused on a single responsibility. If a component grows too large, break it down into smaller sub-components in `src/lib/components/`.
4. **Clear Naming**: Use descriptive names for variables, functions, and stores that reflect their intent (e.g., `isRequestLoading` instead of `loading`).

## Accessibility (A11y) & UI Rules
1. **Semantic HTML**: Use proper tags like `<nav>`, `<main>`, `<header>`, and `<section>`.
2. **ARIA Attributes**: Every interactive element must have appropriate ARIA roles and labels (e.g., `aria-label`, `role="tablist"`).
3. **Focus Management**:
   - Modals MUST implement a focus trap.
   - Use `on:keydown` to handle "Escape" for closing overlays.
4. **Visual Feedback**: Always show loading states using store-provided `loading` flags.
5. **Form UX**: Use `<label>` for all inputs. Use placeholders only for examples.

## Keyboard Shortcuts (Centralized)
- Shortcuts are managed globally in `App.svelte` to avoid conflicts.
- **Priority Shortcuts**:
  - `Ctrl+Enter` / `Cmd+Enter`: Send Request.
  - `Ctrl+S` / `Cmd+S`: Save (Collection/Request).
  - `Esc`: Close Modals/Dropdowns.

## Component Template Structure
```svelte
<script lang="ts">
  import { collectionStore } from "$lib/stores/collectionStore"; // Simple import
  import Button from "$lib/components/base/Button.svelte"; // Reuse
  import { Send } from "lucide-svelte"; // Icon standard

  export let requestId: string;

  // Simple reactive state
  $: request = $collectionStore.collections.find(...)

  async function handleSend() {
    await collectionStore.updateRequest(...); // Logic in store
  }
</script>

<div class="container">
  <Button variant="primary" click={handleSend} aria-label="Send request">
    <Send size={16} />
    <span>Send</span>
  </Button>
</div>

<style>
  .container {
    padding: var(--space-md); /* Use global variables */
  }
</style>

## Review Checklist for the Agent

[ ] Are all bindings called through a store?
[ ] Does the component use lucide-svelte icons?
[ ] Does every button/input have an aria-label or associated <label>?
[ ] Is there a loading state for async operations?
[ ] If a modal is added, does it handle the Esc key and focus trap?
