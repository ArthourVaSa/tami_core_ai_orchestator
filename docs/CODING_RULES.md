# Coding Rules

These rules apply to every change in this repository. When a rule conflicts with idiomatic Go, idiomatic Go wins, and the deviation is noted here.

Language: Go. TUI framework: Bubble Tea (adapt section 10 if another framework is used).

> Status: the code written before SPEC-001 does not follow these rules yet. SPEC-001 migrates it.
> Every new change must follow them from day one.

---

## 1. Project structure (Clean Architecture)

```
cmd/
  tami/
    main.go                  # entry point: loads config, calls bootstrap, runs the TUI
internal/
  bootstrap/                 # composition root: builds and wires every dependency (DI)
    container.go
    providers.go
    usecases.go
  core/                      # reusable, cross-cutting building blocks; imports nothing internal
    constant/                # app-wide constants shared by several packages
      app.go
      limits.go
      paths.go
      env.go
    config/                  # typed configuration: structs, defaults, loading, validation
    clock/                   # time abstraction
    logger/                  # slog setup and secret-redaction handler
    sanitize/                # terminal-output sanitization
  domain/
    entity/                  # reusable business entities (Account, Provider, Conversation, Message)
    port/                    # interfaces the domain needs (ProviderClient, CredentialStore, SessionRepository)
    errs/                    # domain sentinel errors
  usecase/
    sendmessage/
      usecase.go             # the use case type, its constructor and its methods only
      input.go               # request model used only by this use case
      output.go              # response model used only by this use case
      usecase_test.go
    switchaccount/
      ...
  adapter/
    provider/
      anthropic/             # one package per provider, each implements port.ProviderClient
      openai/
      ollama/
    credential/
      keychain/
    storage/
      sqlite/
  tui/
    navigation/              # screen identifiers and navigation messages (the "router")
    theme/                   # shared styles and colors
    keymap/                  # shared key bindings
    text/                    # user-facing strings
    component/
      chatview/
      accountpicker/
    app/                     # root Bubble Tea model
```

Dependency direction (an arrow means "may import"):

```
cmd → bootstrap → tui, adapter, usecase, domain, core
tui     → usecase, domain, core
adapter → domain, core
usecase → domain, core
domain  → core (constants only)
core    → nothing internal
```

Rules:

- Dependencies point inward only. `domain` never imports `usecase`, `adapter` or `tui`.
- `core` is importable by everyone and therefore imports nothing from this project. This is what prevents import cycles.
- `bootstrap` imports everything and is imported only by `cmd`. Nothing else may import it.
- Everything not meant to be imported by other modules lives under `internal/`.
- An `entity` is a business concept reused across use cases. A model (input/output/DTO) belongs to exactly one use case or one adapter.
- If a model starts being used by a second use case, either promote it to `domain/entity` (if it is a business concept) or to `domain/port` (if it is part of a port contract). Never import one use case package from another.
- Provider-specific request/response structs (wire formats) live inside their adapter package and never leak outside it. Adapters map them to domain entities.

## 2. Single Responsibility: one thing per file

- One exported type per file. The file contains that type, its constructor and its methods. Nothing else.
- File name matches the type in snake_case: `SessionRepository` → `session_repository.go`.
- Unexported helper types used by that type go in their own file inside the same package.
- Allowed exceptions (idiomatic Go):
  - A named type and its constants/enum values (`type Role string` + `RoleUser`, `RoleAssistant`).
  - Functional option types and their `WithX` functions, in an `options.go` file.
  - Sentinel errors of a package, grouped in an `errors.go` file.
  - Package-level shared constants, grouped in a `constants.go` file (see section 4).
- One use case per package. A use case exposes one public method (`Execute`, or a domain verb like `Send`).
- Soft limits: a file over ~250 lines or a function over ~40 lines is a signal to split. Cyclomatic complexity per function ≤ 10 (enforced by linter).

## 3. Package design (Go has no barrel files)

Barrel files (`index.ts`, `export.dart`) do not exist in Go and must not be emulated:

- In Go, a package is a directory. Everything exported in any file of that directory is already available through a single import path. The package itself is the "barrel".
- Do not create re-export files using type aliases or wrapper functions to simulate barrels.
- Control the public surface with capitalization: export only what consumers need; everything else is unexported.
- Package names: short, lowercase, singular, no underscores, describing what it provides (`entity`, `keychain`, `sendmessage`, `constant`). Forbidden names: `util`, `utils`, `common`, `helpers`, `misc`, `base`, `models`.
- Avoid stutter: `sendmessage.UseCase`, not `sendmessage.SendMessageUseCase`; `constant.MaxRetries`, not `constant.ConstMaxRetries`.
- No import cycles. If two packages need each other, the shared contract belongs in `domain/port`.
- Each package has a `doc.go` with a one-line package comment.

## 4. Constants and hardcoded values

No hardcoded strings or numbers inside logic. Every literal with meaning gets a name. Where that name lives depends on who uses it.

Note: Go has no file-private scope. An unexported identifier is visible to the whole package. The levels below are therefore a placement convention, enforced in code review.

| Used by | Where it goes | Visibility |
|---|---|---|
| One file only | `const` block at the top of that file, after imports | unexported (`defaultPageSize`) |
| Several files of the same package | `constants.go` in that package | unexported |
| Several packages, and it belongs to a business concept | the package that owns the concept (`entity.RoleUser`, `ollama.DefaultPort` if only adapters need it) | exported |
| Several packages, app-wide and not owned by any concept | `internal/core/constant`, in a file grouped by topic (`app.go`, `limits.go`, `paths.go`, `env.go`) | exported |

Rules:

- Before adding to `core/constant`, check whether a domain or adapter package owns the concept. `core/constant` is for truly cross-cutting values (app name, config directory name, default timeouts, env var names), not a dumping ground.
- Prefer `const` over `var`. Use `var` only for values Go cannot make constant (slices, maps, regexps, `time.Duration` computed at runtime); such package-level vars must never be mutated.
- Use typed constants for closed sets of values (`type ProviderKind string` with `ProviderAnthropic`, `ProviderOllama`), so the compiler rejects invalid values. Use `iota` for internal enums that are never persisted or serialized.
- Group related constants in one `const ( ... )` block with a short doc comment for the group.
- Durations always carry their unit: `const requestTimeout = 30 * time.Second`, never `30`.
- A literal does not need a name when its meaning is obvious and universal: `0`, `1`, `""`, `len(x) - 1`, HTTP status codes from `net/http`.
- Constants vs configuration:
  - If the value may differ per user, environment, or machine (endpoints, model names, budgets, timeouts the user can tune), it is configuration and belongs in `core/config` with a default, not in `core/constant`.
  - If changing the value requires a code change and a new release, it is a constant.
- Secrets, API keys, tokens, and credentials are never constants or config defaults. They come only from `port.CredentialStore`.
- User-facing text shown in the TUI lives in `tui/text`, not scattered across components and not in `core/constant`. This also prepares the app for translation.

## 5. Configuration

- `core/config` defines typed structs (`Config`, `ProviderConfig`, `LimitsConfig`), their defaults, loading (file + env), and `Validate()`.
- Configuration is loaded once in `cmd/tami/main.go`, validated, and passed to `bootstrap`. If validation fails, the app exits with a clear message.
- No package reads env vars, flags, or config files on its own. Each component receives only the sub-config it needs through its constructor (`sendmessage.New(..., cfg.Limits)`), never the whole `Config`.
- Env var names and config keys are constants in `core/constant/env.go`.
- Config structs never contain secrets, so they are always safe to log or print for debugging.

## 6. Dependency injection

- All dependencies are passed through constructors: `func New(provider port.ProviderClient, store port.SessionRepository, ...) *UseCase`.
- When a constructor needs more than 4 parameters or has optional settings, use functional options (`New(required, WithTimeout(t), WithLogger(l))`) or a small `Deps` struct.
- Wiring happens only in `internal/bootstrap` (the equivalent of a DI container module in other ecosystems). It exposes one entry point, e.g. `bootstrap.Build(ctx, cfg) (*Container, error)`, which returns the fully wired root TUI model and the cleanup functions.
- `bootstrap` is split by layer: `providers.go` (adapters), `usecases.go`, `tui.go`. Manual wiring is the default; a code generator such as `google/wire` is acceptable only if manual wiring becomes hard to follow.
- No runtime service locator: no global container, no `Get[T]()` lookups from inside business code. Dependencies are always explicit in constructor signatures.
- Forbidden: global mutable state, package-level singletons, `init()` with side effects.
- Time, randomness, filesystem, environment, and network are dependencies too. Inject `clock.Clock`, an `fs.FS`, an `*http.Client`, etc., so tests are deterministic.
- `context.Context` is always the first parameter of any function that does I/O or can block. Never store a context in a struct.

## 7. Interfaces and design patterns

Use interfaces at boundaries, not everywhere. Go idiom: "accept interfaces, return concrete types".

- Define an interface where it is consumed (in `domain/port` or in the consuming package), not next to its implementation.
- Keep interfaces small (1–3 methods). Compose larger ones from small ones.
- Do not create an interface with a single implementation unless it is a boundary that needs a test double (I/O, providers, storage, clock).
- Constructors return concrete types (`*UseCase`), not interfaces.
- Add a compile-time check in each adapter: `var _ port.ProviderClient = (*Client)(nil)`.

Patterns to use where they solve a real problem:

| Pattern | Where |
|---|---|
| Ports & Adapters | Every external system (providers, keychain, storage) |
| Adapter | Mapping each provider's API to `port.ProviderClient` |
| Strategy | Selecting provider, model, or tokenizer at runtime |
| Registry / Factory | Building the right provider client from an account config |
| Decorator / Middleware | Cross-cutting behavior wrapped around a port: retries, rate limiting, logging, cost tracking, secret redaction |
| Functional Options | Optional configuration of constructors |
| Repository | Persistence of sessions, accounts, settings |
| Observer (channels / `tea.Msg`) | Streaming tokens and events to the UI |

Avoid patterns that fight Go: getters/setters for every field, deep inheritance-like embedding chains, abstract factories of factories, generic "manager" or "service" god types.

## 8. Readability

- Names describe intent: `activeAccount`, not `acc2`; `ErrAccountNotFound`, not `ErrNF`. Short names are fine only in small scopes (`i`, `r`, `ctx`).
- Early returns over nested `if/else`. Happy path stays left-aligned.
- No magic numbers or strings (see section 4).
- One level of abstraction per function: a function either orchestrates or does low-level work, not both.
- Prefer plain structs and functions over clever generics; use generics only when they remove real duplication.
- Booleans in signatures are a smell; prefer explicit types or options (`Send(ctx, msg, Streaming)` over `Send(ctx, msg, true)`).
- Code must be understandable by a junior developer without asking the author.

## 9. Comments

- Every exported identifier has a one-line doc comment starting with its name (required by Go tooling): `// Execute sends the message to the active provider and stores the reply.`
- Comments explain **why**, never **what**. If a comment is needed to explain what code does, rename or extract instead.
- No commented-out code, no changelog comments, no author names.
- `TODO` comments must reference an issue: `// TODO(#42): support tool calling for Ollama.`
- Security-relevant decisions get a short why-comment (e.g. why a value is sanitized or why a check exists).

## 10. TUI (Bubble Tea)

- One component per package under `tui/component/`, each with its own `Model`, `Init`, `Update`, `View` in separate files (`model.go`, `update.go`, `view.go`, `keys.go`).
- Navigation (the equivalent of routes/router): `tui/navigation` defines a typed `Screen` constant per screen and the navigation messages (`NavigateTo{Screen}`, `Back{}`). The root model in `tui/app` owns the current screen and a back stack, and is the only place that switches screens. Components request navigation by returning a message, never by touching other components.
- Shared styles live in `tui/theme`, shared key bindings in `tui/keymap`, user-facing strings in `tui/text`. Components never hardcode colors, keys, or text.
- `Update` is pure state transition. All I/O happens in `tea.Cmd` functions that call use cases, never adapters directly.
- `View` only renders; no business logic, no I/O.
- The TUI layer depends on use cases through interfaces, so components can be tested with fakes.
- All text coming from models, tools, files, or the network passes through `core/sanitize` before rendering (strip ANSI/OSC escape sequences and dangerous Unicode).

## 11. Errors

- Errors are values: return them, never `panic` outside of `main` or truly impossible states.
- Wrap with context using `%w`: `fmt.Errorf("load account %q: %w", id, err)`. Messages are lowercase, no trailing punctuation.
- Domain errors are sentinel values or typed errors in `domain/errs`; check with `errors.Is` / `errors.As`, never by comparing strings.
- Adapters translate provider/HTTP errors into domain errors (`ErrRateLimited`, `ErrUnauthorized`, `ErrContextTooLong`).
- Handle each error once: either log it or return it, not both.
- Error messages and logs never contain secrets, tokens, full headers, or full prompts.

## 12. Concurrency

- Every goroutine has a clear owner and a way to stop (`context` cancellation). No fire-and-forget goroutines.
- Use `errgroup` for groups of concurrent tasks.
- Channels are owned and closed by the sender only.
- Shared state is protected by a mutex inside the type that owns it; never expose the mutex.
- Tests run with `-race`.

## 13. Testing

- Every use case and adapter has tests. Target ≥ 80% coverage in `domain` and `usecase`.
- Table-driven tests with `t.Run` and descriptive case names.
- Prefer hand-written fakes over mock frameworks; fakes live in a `<package>test` package (e.g. `porttest`) when reused.
- Adapters are tested against `httptest.Server`, never real provider APIs. Real API keys are never used in tests.
- Golden files for TUI `View` output where useful.
- Arrange / Act / Assert structure; one behavior per test.
- Test-only literals may stay inline in the test when they make the case easier to read.

## 14. Logging and security

- Structured logging with `log/slog`, configured in `core/logger` and injected as a dependency. A redaction handler masks anything that looks like a key or token.
- Secrets are only read through `port.CredentialStore`; they never appear in config structs, constants, logs, or serialized data.
- Any input crossing a trust boundary (model output, tool output, files, MCP responses, user-provided URLs) is treated as untrusted and validated at the edge.

## 15. Tooling (enforced in CI)

- `gofmt` / `goimports` on every file.
- `golangci-lint` with at least: `govet`, `staticcheck`, `errcheck`, `revive`, `gocyclo`, `gosec`, `ineffassign`, `unused`, `misspell`, `nolintlint`, `goconst` (repeated string literals), `mnd` (magic numbers), `depguard` (enforces the dependency direction in section 1, e.g. `core` cannot import `domain`, `domain` cannot import `adapter`).
- `govulncheck` on every PR.
- `go test -race ./...` must pass.
- `go mod tidy` leaves no diff.

## 16. Definition of done for any change

- [ ] Follows the folder, dependency-direction, and one-type-per-file rules.
- [ ] No hardcoded meaningful literals; constants placed at the right level (section 4).
- [ ] New tunable values go to `core/config`, not constants.
- [ ] New dependencies are injected and wired in `bootstrap`, not instantiated elsewhere.
- [ ] New external interaction goes through a port.
- [ ] Tests added or updated; linters and race detector pass.
- [ ] No secrets in code, logs, tests, or fixtures.
- [ ] Exported identifiers have one-line doc comments.
