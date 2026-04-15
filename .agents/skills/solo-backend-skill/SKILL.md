---
# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

name: solo-backend-skill
description: Guidelines for Solo Backend Developer
---
# Solo Backend Developer Skill (v1.0)

You are a Senior Backend Engineer for Solo. Your goal is to extend the Go backend of this Wails desktop app cleanly, consistently, and without breaking the existing architecture. Use this skill whenever the user wants to plan o implement anything in go files.

---

## Stack & Entry Points

- **Language**: Go (module `solo`, requires Go 1.25+)
- **Desktop framework**: [Wails v2](https://wails.io/) — the `App` struct in `app.go` is the single boundary between the Go backend and the Svelte frontend.
- **Logging**: `log/slog` (structured JSON, written to `~/.solo/logs/solo.log` via `lumberjack`). Log levels can be toggled at runtime via `App.SetDebugMode(bool)`.
- **Scripting runtime**: `gopher-lua` (Lua 5.1) — sandboxed, no I/O or OS libs, 500 ms execution timeout per script.
- **Key external deps**: `github.com/google/uuid` (ID generation), `github.com/samber/lo` (utilities).

---

## Project Layout

```
app.go                         // Wails App struct + all exported methods (the "controller")
main.go                        // Wails bootstrap, logging setup
internal/
  collection/                  // Collection + Request domain model + CollectionManager
  environment/                 // Environment + ValueType domain model + EnvironmentManager
  configuration/               // Configuration struct + ConfigurationManager (sync.RWMutex)
  host/                        // Host (TLS/cookies) + HostManager (HTTP client pool)
  requester/                   // Service: executes HTTP requests, applies config, runs scripts
  runner/                      // Parallel load-test runner (goroutines + semaphore)
  script/                      // ScriptManager (Lua), EnvAPI, request/response converters
  importer/                    // Importer interface + Postman v2.1, Bruno, Postman Env importers
  git/                         // Manager: wraps system `git` CLI for collection/env sync
  theme/                       // Theme model + predefined presets
  tools/                       // Constants, fs helpers, placeholder regex
```

---

## The `App` Struct (`app.go`)

`App` is the **only** struct bound to Wails. Every public method on it automatically becomes callable from the frontend via `window.go.<MethodName>()`.

### Rules for adding new `App` methods

1. **Every exported method is a frontend RPC** — name it clearly, document it with a single-line Go comment.
2. **Guard all manager fields for nil** before use. The pattern used throughout is:
   ```go
   func (a *App) DoSomething(...) error {
       if a.someManager == nil {
           return fmt.Errorf("someManager not initialized")
       }
       return a.someManager.DoSomething(...)
   }
   ```
3. **Delegate to internal packages** — `App` methods are thin wrappers. Business logic belongs in the corresponding `internal/` package, not in `app.go`.
4. **Use `runtime.EventsEmit`** (not return values) for push notifications to the frontend. Always check `a.ctx != nil` before emitting.
5. **`RequestOptions`** is the canonical DTO used to pass HTTP request parameters from the frontend to `Execute` / `RunParallel`. Do not create parallel DTOs.

### Lifecycle hooks

| Hook | Purpose |
|---|---|
| `startup(ctx)` | Saves Wails context; injects it into `ScriptManager` |
| `beforeClose(ctx) bool` | Emits `app:request-close` event; vetoes close until frontend confirms via `ForceQuit()` |

---

## Domain Models

### Collection (`internal/collection/`)

```go
type Collection struct {
    Id                  string    // uuid
    Name                string    // used as filename: <Name>.json
    Requests            []Request
    CreationTimestamp   time.Time
    LastUpdateTimestamp time.Time
    GitRemote           string    `json:",omitempty"`
    GitPath             string    `json:",omitempty"`
    GitProvider         string    `json:",omitempty"`
}
```

- **Not concurrency-safe** by design — all mutations go through `CollectionManager` which does load-mutate-save.
- `UpdateRequest` uses reflection to copy non-zero fields; it never overwrites `Id`, `CreationTimestamp`, `LastUpdateTimestamp`.
- `Request.BodyType` defaults to `"json"` on unmarshal if empty.
- Placeholders (`{{varName}}`) are extracted from URL, body, headers, and cookies via `Request.GetPlaceholders()`.

### Environment (`internal/environment/`)

```go
type Environment struct {
    Id   string
    Name string                  // used as filename: <Name>.json
    Values map[string]ValueType  // key -> { Value string, Type string }
    CreationTimestamp   time.Time `json:"creation_timestamp"`
    LastUpdateTimestamp time.Time `json:"last_update_timestamp"`
    GitRemote, GitPath, GitProvider string `json:",omitempty"`
}
```

- `ValueType.Type` is a free-form string (e.g. `"text"`, `"secret"`) — the backend does not enforce an enum.
- Note: `Collection` timestamps use camelCase JSON keys; `Environment` timestamps use snake_case JSON keys. **Do not change this** — it would break existing user data.

### Configuration (`internal/configuration/`)

```go
type Configuration struct {
    General      GeneralSettings   // theme, themeMode, dayTheme, nightTheme, checkForUpdates, selectedEnvironment
    Request      RequestSettings   // timeoutSeconds, followRedirects, maxRedirects, validateSSL, defaultUserAgent, proxyUrl
    CustomThemes []theme.Theme
}
```

- `ConfigurationManager` is **the only concurrency-safe manager** (uses `sync.RWMutex`). Always use `Get()` (returns a copy) and `Save(cfg)`.
- `RequestSettingsOverride` mirrors `RequestSettings` with pointer fields: a `nil` field means "use global setting". This cascade logic lives in `requester/service.go`.

### Host (`internal/host/`)

```go
type Host struct {
    Id      string
    Name    string              // hostname (e.g. "api.example.com") — used as map key
    TlsConfig TLSConfig         // mTLS: cert+key+CA paths, InsecureSkipVerify
    Cookies   map[string]string // per-host cookies added to every request
}
```

- `HostManager` maintains an `*http.Transport` pool keyed by `host:port`. Evict by hostname when `UpsertHost` or `DeleteHost` is called.
- TLS config requires both `PublicCertificateFilePath` and `PrivateKeyFilePath` for mTLS; a CA cert is optional (for self-signed server certs).
- Min TLS version: 1.2. Preferred curves: P-256, X25519.

---

## HTTP Execution Pipeline (`internal/requester/`)

```
App.Execute(RequestOptions)
  └─> requester.Service.ExecuteRequest(ExecutionOptions)
        ├─> HostManager.GetClientForUrl()    // get/create pooled http.Transport
        ├─> Apply global config + per-request overrides (timeout, redirects, proxy, SSL)
        ├─> Build http.Request (method, url, headers, cookies)
        ├─> ScriptManager.ExecutePreRequest()  // optional Lua: mutates request
        ├─> client.Do(request)
        └─> ScriptManager.ExecutePostResponse() // optional Lua: can write session vars
```

`ResponseData` (statusCode, headers, body as string, duration in ms) is returned to the frontend.

---

## Scripting (`internal/script/`)

- `ScriptManager` owns a single `lua.LState` protected by `sync.Mutex` — scripts run **sequentially**.
- Each execution gets a 500 ms `context.WithTimeout`. The Lua state is reused across executions.
- **Allowed Lua libraries**: `base`, `table`, `string`, `math` — no I/O, no OS, no coroutine, no package.
- **`env` global** exposed to Lua:

  | API | Behaviour |
  |---|---|
  | `env.get(key)` | Reads from `sessionVars` first, then `currentEnv` |
  | `env.set(key, val)` | Writes to `sessionVars`; emits `session_vars_updated` event |
  | `env.log(msg)` | Logs to slog at Info level with `[LUA]` prefix |
  | `env.varName` | Dynamic read via `__index` metamethod |
  | `env.varName = val` | Dynamic write via `__newindex` metamethod |

- **Pre-request**: `request` global table (`method`, `url`, `body`, `headers`) is writable — mutations are applied back to `http.Request`.
- **Post-response**: `request` and `response` globals (`status`, `body`, `headers`, `time`) are exposed read-only; post-response errors are non-fatal (logged, response still returned).
- `sessionVars` are ephemeral (in-memory only); `currentEnv` is read-only for Lua.

---

## Parallel Runner (`internal/runner/`)

- Goroutine pool controlled by a buffered semaphore channel of size `Concurrency`.
- Each result is sent to `resultChan` and forwarded via the `onResult` callback (used to emit `runner:result` Wails events).
- `RunnerStats` aggregates: total/success/error counts, min/max/avg/p95 latency (ms), req/s, status code distribution.
- `StopOnError = true` cancels the parent context, stopping new dispatches.

---

## Importers (`internal/importer/`)

All importers implement:
```go
type Importer interface {
    Import(path string) (*collection.Collection, error)
}
```

| Importer | Input | Notes |
|---|---|---|
| `PostmanImporter` | `.json` file (Postman v2.1) | Flattens folders recursively; `postmanURL` has a custom unmarshaler handling both string and object URL formats |
| `BrunoImporter` | directory with `bruno.json` + `.bru` files | Parses Bruno DSL line-by-line; folder structure becomes `"folder / request"` names |
| `PostmanEnvironmentImporter` | `.json` Postman env file | — |
| `BrunoEnvironmentImporter` | `.bru` env file or directory | — |

When adding a new importer: implement the interface, add a `New<Format>Importer()` constructor, wire it in `app.go`.

---

## Git Integration (`internal/git/`)

- `Manager` is **stateless** — it wraps the system `git` CLI via `exec.CommandContext`.
- All commands run with `GIT_TERMINAL_PROMPT=0` and `GIT_ASKPASS=echo` to avoid hanging on auth prompts.
- **Sparse-checkout** is used to clone only the relevant path from large repos.
- URL format supports branch selection: `https://github.com/org/repo#branch`.
- **Sync flow** (`SyncGitCollection`): stage local changes → commit if dirty → fetch → rebase on top of remote → push.
- Git repos are stored at `~/.solo/git_storage/<sha1(url)[:8]>` (collections) or `~/.solo/git_storage/env_<sha1(url)[:8]>` (environments).
- `CollectionStatus` captures: branch, rebase-in-progress flag, conflict files, dirty flag, ahead/behind counts, last 5 log entries.

---

## File Storage (`internal/tools/`)

All user data lives under `os.UserConfigDir()/solo/`:

| Path | Content |
|---|---|
| `config.json` | `Configuration` struct (themes, request settings) |
| `collections/<name>.json` | One `Collection` per file |
| `environments/<name>.json` | One `Environment` per file |
| `hosts/host.json` | `[]Host` array |
| `git_storage/<hash>/` | Sparse-checkout git repos |
| `logs/solo.log` | Rotating JSON log (10 MB max, 3 backups, 1 day) |

Helper functions in `tools/fs.go`:

| Function | Use |
|---|---|
| `GetOrCreateConfigDir()` | Returns (and creates) the base config dir |
| `GetMainConfig(subdir)` | Returns path to a named subdirectory |
| `CreateConfigFile(dir, name, data)` | Write with `0600` perms (new file) |
| `UpdateConfigFile(dir, name, data)` | Write with `0666` perms (overwrite) |
| `ReadConfigFile(dir, name)` | Read raw bytes |
| `ReadConfigDirectory(dir)` | Returns `[]os.DirEntry` (empty slice if not exists) |
| `RemoveConfigFile(dir, name)` | Delete a file |

---

## Placeholder System

Placeholders follow the `{{varName}}` syntax (regex: `\{\{(.*?)\}\}`).

- `tools.ExtractPlaceholders(text)` returns unique placeholder names from a string.
- `Request.GetPlaceholders()` unions placeholders from URL, body, headers, and cookies.
- `Environment.GetSelectedValues(keys)` returns only the values matching the given keys.
- `App.ResolveRequestPlaceholders(reqId, collName, envId)` is the end-to-end resolver called by the frontend before executing a request.

---

## Coding Rules

1. **No logic in `app.go`** — all non-trivial logic belongs in `internal/` packages. `app.go` is the RPC façade only.
2. **Nil-guard every manager** before use (see pattern above).
3. **Load-mutate-save** is the only concurrency model for collections and environments (no in-memory caching between calls). `ConfigurationManager` is the exception (uses `sync.RWMutex` with an in-memory copy).
4. **Use `slog` for all logging**, never `fmt.Println`. Prefer structured fields: `slog.Info("msg", "key", value)`.
5. **IDs are always UUIDs** (`github.com/google/uuid`). Use `uuid.NewString()` for new entities; never generate IDs on the frontend.
6. **JSON field names** follow existing conventions — do not silently rename them (it breaks saved user data). Check existing structs before adding fields.
7. **File permissions**: `0600` for new sensitive files (collections, environments), `0666` for updates, `0755` for directories.
8. **Git operations** must use `executeWithTimeout` — never call `git` without a timeout.
9. **Do not add new external dependencies** without a clear reason. The dependency footprint must stay minimal.
10. **Tests** live alongside source files (`_test.go`). Write table-driven tests; use the standard `testing` package.

---

## Wails Events Reference

| Event name | Emitted by | Payload |
|---|---|---|
| `app:request-close` | `App.beforeClose` | — |
| `runner:result` | `App.RunParallel` (via callback) | `runner.RunnerResult` |
| `session_vars_updated` | `script.EnvAPI.Set` / `NewIndex` | `map[string]string` |

When adding new events, document them here and keep the event name in `snake_case` or `kebab-case` (both patterns exist — match the domain).

---

## Adding a New Feature — Checklist

1. Define the domain struct in the appropriate `internal/` package.
2. Add CRUD operations on the manager (load-mutate-save pattern).
3. Add nil-guarded `App` methods in `app.go` that delegate to the manager.
4. If the feature requires a new config file, add a constant in `tools/constants.go` and a helper in `tools/fs.go` if needed.
5. If the feature emits events to the frontend, add them to the events reference above.
6. Write unit tests for the internal package logic.
7. Run `wails build` (or `wails dev`) to regenerate frontend TypeScript bindings.
