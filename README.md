<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<div align="center">


<html>
    <h2 align="center" style="display: flex; flex-direction: column; align-items: center; padding: 2rem">
      <img src="hack/icon/icon-lighning-spike-round-inverted.svg" width="128"/>
    </h2>
    <img width="256" src="hack/icon/icon-ascii.svg" />
</html>


**The lightweight, fast, open-source API client for modern development.**

[![Build Status](https://github.com/raml-dev/solo/actions/workflows/ci.yaml/badge.svg)](https://github.com/raml-dev/solo/actions)
[![Wails](https://img.shields.io/badge/Wails-v2-red)](https://wails.io)
[![License](https://img.shields.io/badge/license-%20%20AGPL3.0-blue?style=plastic)](LICENSE)

[Features](#-features) • [Installation](#-installation) • [Contributing](#-contributing)



</div>

---

**Solo** is a high-performance API client built with _Wails_, combining the efficiency of a native desktop application with the modern flexibility. Designed to be fast, reliable, and versatile, Solo acts as a single hub for all your API testing and debugging needs.

## ✨ Features

- **Blazing Fast**: Minimal memory footprint and instant startup.
- **Smart Collections**: Effortlessly organize and manage your REST requests.
- **Dynamic Environments**: Seamlessly switch between development, staging, and production.
- **Git Native Sync**: Import and synchronize collections directly via Git.
- **Interoperability**: First-class support for importing **Postman**, **Bruno**, and **OpenAPI**.
- **Scripting with Lua**: Power your workflow with pre-request and post-response Lua scripts.
- **Modern UI**: Clean, intuitive interface with full Dark Mode support.

## 🚀 Installation

For detailed guides and full documentation, please visit our documentation site.

### Download

Download the latest version for your platform from the [Releases](https://github.com/raml-dev/solo/releases) page.

### Build from source

The easiest way to build Solo is using [mise](https://mise.jdx.dev/), which automatically manages all required dependencies including Go, Node.js, and the Wails CLI.

1. Clone the repository:
   ```bash
   git clone https://github.com/raml-dev/solo.git
   cd solo
   ```
2. Install dependencies and build:
   ```bash
   mise install
   wails build
   ```

If you prefer to manage dependencies manually, ensure you have [Go](https://go.dev/), [Node.js](https://nodejs.org/), and the [Wails CLI](https://wails.io/docs/gettingstarted/installation) installed before running `wails build`.

## 🤝 Contributing

Contributions are welcome.

- Open an [Issue](https://github.com/raml-dev/solo/issues) for bugs or ideas
- Submit a [Pull Request](https://github.com/raml-dev/solo/pulls)

Read the [contribution guidelines](CONTRIBUTING.md) before opening an issue or submitting a PR.

---

<div align="center">
Made by the raml-dev
</div>
