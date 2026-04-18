<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

# Contributing to the project

Thank you for your interest in contributing to `Solo`.
This project is open source and every contribution is welcome, but to maintain quality and consistency please follow these guidelines.

---

## Philosophy

The project prioritizes:

- simplicity over premature abstraction
- clarity over "clever code"
- explicit behavior over magic
- small iterative changes

If a solution is difficult to explain, it is probably too complex.

> ### AI disclaimer
>
> We are strong supporters of automated programming tools and we use a wide variety of them. For this reason, we ask for the intelligent use of these tools and we are firmly committed to performing human code reviews. For this reason, we ask you to contribute with clear, atomic code that can be explained by the contributor themselves.

---

## How to contribute

### 1. Always open an issue

Before working on:

- new features
- architectural changes
- important refactors
- bug fixes

If you do not want to necessarily wait for the triage phase, you can start working on it and submit a PR.

---

### 2. Fork and branch

- Fork the repository
- Create a descriptive branch directly from the opened GitHub issue, using the provided template.

### 3. Code and tools

- Write clear and moderately commented code.
- Use [mise](https://mise.jdx.dev/) for dev tools management
- Use [SemVer](https://semver.org/)

## Skills and guidelines

This project includes "skills" (operational guidelines) available in standard paths (.agents, etc):

- use them as a reference
- they are valid even without AI tools

## Commit

Write clear and descriptive commits:

- feat(component): new feature
- fix(component): bug fix
- refactor(component): change without functional impact

Avoid generic commits such as:

- "fix"
- "update"
- "various changes"

Always include the specific component or functionality.

## Pull Request

When opening a PR:

- describe what it does
- explain why
- keep it as small and atomic as possible

PRs that are too large may be rejected.

> **Golden rule:** if you cannot clearly explain the change in the description, the PR is probably too complex.

## Triage

Issues and PRs are evaluated during triage.

A PR may be closed if:

- not aligned with the project philosophy
- too complex or out of scope
- lacking sufficient context

We commit to providing clear feedback when possible.

## Test

When applicable:

- do not break existing tests
- add tests for new features

## Security Issue

Contributors are always reachable at the raml-dev organization email address. If you find a vulnerability please avoid opening an issue about it and contact us promptly via email to prevent exploitation by malicious users.
We will ask for further details if needed.
