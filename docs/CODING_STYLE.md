# CODING_STYLE.md

## 1. Purpose

This document defines how code is written in this repository.

The goal is not aesthetic consistency. The goal is code that is readable, reviewable, testable, and safe to operate on customer infrastructure.

The project is a surgical tool. Code should be direct, small, explicit, and boring.

## 2. Core Rules

### 1. One task, one surface.

Each change should touch the smallest reasonable surface area.

Do not mix unrelated work.

Bad:

```text
Add sensor collection, rewrite auth middleware, rename database tables, and change UI routing.
```

Good:

```text
Add procfs memory collector.
```

### 2. Top-down organization.

Put the public entry point first. Put helper functions below it.

A reader should understand the operation before reading the implementation details.

### 3. Avoid deep nesting.

Return early.

Bad:

```go
func handle() error {
    if ready {
        if allowed {
            if valid {
                return run()
            }
        }
    }
    return nil
}
```

Good:

```go
func handle() error {
    if !ready {
        return ErrNotReady
    }

    if !allowed {
        return ErrDenied
    }

    if !valid {
        return ErrInvalid
    }

    return run()
}
```

### 4. Keep functions small.

A function should do one clear operation.

Split parsing, validation, authorization, execution, and persistence when they are distinct concepts.

### 5. Human-readable names.

Use names that describe the thing.

Good:

```go
userID
SensorSample
sessionStore
```

Bad:

```go
u
SS
mgr2
```

Abbreviations are allowed only when project-wide and obvious: `db`, `ctx`, `cfg`, `id`.

### 6. No boolean flag arguments.

Boolean flags hide behavior at call sites.

Bad:

```go
CreateSession(userID, true)
```

Good:

```go
CreatePersistentSession(userID)
CreateEphemeralSession(userID)
```

If the operation has several options, use an options struct.

### 7. No hidden side effects.

Function names and docs must make side effects obvious.

A function that writes to the database, disk, network, process table, socket, or global state must say so through its name, type, or documentation.

Prefer:

```go
sessionStore.Save(ctx, session)
```

over:

```go
save(session)
```

### 8. No global mutable state.

Allowed globals:

- Constants.
- Sentinel errors.
- Embedded filesystems.
- `sync.Once` for explicit initialization.

Forbidden globals:

- Package-level mutable maps.
- Package-level mutable slices.
- Implicit registries mutated by `init()`.
- Shared state hidden behind package functions.

### 9. Prefer explicit types over clever abstractions.

Good:

```go
type UserID string
type Action string
```

Bad:

```go
map[string]any
interface{}
```

Use `any` only at JSON boundaries or when no narrower type is honest.

### 10. Functions should read top to bottom.

Avoid callbacks that mutate outer scope.

Avoid deferred functions that change return values.

Avoid control flow that requires the reader to jump across the file.

### 11. Comments explain why, not what.

Good:

```go
// Some 5.4 kernels report hwmon labels before values. Retry once to avoid a false missing-sensor state.
```

Bad:

```go
// Loop over sensors.
```

### 12. No dead code.

Delete dead code.

Git remembers.

### 13. No speculative abstractions.

Do not add an interface until there are at least two real implementations or a test seam that materially improves correctness.

Do not add factories, registries, plugin systems, or generic frameworks before the concrete need exists.

### 14. No framework magic.

Avoid hidden behavior.

Forbidden patterns:

- Struct tags that trigger non-obvious execution.
- `init()` with side effects except narrow flag registration.
- Reflection-heavy routing or persistence.
- Code generation that reviewers cannot easily inspect.

### 15. No copy-paste logic.

Extract common code at the second use within a package.

Do not force reuse across domain boundaries if it damages readability.

Duplication across unrelated domains can be acceptable when the concepts are different.

## 3. Go Specifics

### Errors

Wrap errors with useful context.

```go
return fmt.Errorf("load config: %w", err)
```

Wrap once at the boundary where context is added.

Do not log and return the same error unless the caller cannot log it.

### Context

`context.Context` is the first parameter for I/O, RPC, database operations, external commands, and long-running work.

```go
func (s *Store) SaveSession(ctx context.Context, session Session) error
```

Do not store contexts in structs.

### Logging

Use structured logging through injected dependencies.

No `log.Println` in libraries.

Command entry points may configure the logger.

### External Commands

Never build shell strings from user input.

Use argument arrays.

```go
cmd := exec.CommandContext(ctx, "systemctl", "restart", serviceName)
```

Validate `serviceName` before use.

### Interfaces

Keep interfaces small.

Define interfaces at the consumer boundary, not the producer boundary, unless a package-level contract is necessary.

### Tests

Use table tests for logic.

Use `package_test` for black-box behavior when practical.

Use `t.Parallel()` where safe.

Test negative paths.

Tests that protect invariants are mandatory.

## 4. TypeScript and React Specifics

### TypeScript

No `any`.

Use `unknown` with a type guard when input is untrusted.

Define API request and response types explicitly.

### React Components

Use props interfaces.

```tsx
interface ButtonProps {
  label: string;
  disabled?: boolean;
}
```

Do not define complex inline prop types in function signatures.

### Hooks

Hooks must be at the top of the component.

No conditional hooks.

Effects must clean up subscriptions, timers, and streams.

### Data Fetching

Do not scatter raw fetch calls through components.

Use a small API client layer.

React Query or SWR may be used if accepted as a deliberate dependency. Otherwise use a small local data-fetching abstraction.

### Component Size

Components should stay under 200 lines.

Start extracting at 150 lines when a clear component boundary exists.

### UI Authority

The UI is not authoritative.

Do not make authorization or policy decisions only in React.

The server must re-check every privileged action.

## 5. Review Rules

A reviewer should reject code that:

- Violates an invariant.
- Adds hidden privileged behavior.
- Adds broad abstractions without need.
- Expands the trusted computing base without justification.
- Uses shell execution where a system integration layer should exist.
- Adds unaudited privileged mutation.
- Makes the UI authoritative.
- Silently ignores errors.
- Silently ignores unsupported capabilities.
