# Koji

Koji is a small server control panel for bare-metal and VPS systems.

The project is designed as a surgical tool:

- Small static Go backend where practical.
- TypeScript and pure React UI.
- Unprivileged web/API daemon.
- Narrow privileged local agent.
- SQLite durable state.
- Native Linux data sources first.
- Strict configuration and capability enforcement.

Before implementation, read:

1. `ARCHITECTURE.md`
2. `INVARIANTS.md`
3. `CODING_STYLE.md`
4. `SECURITY.md`
5. `PHASEMAP.md`
6. `AGENTS.md`

Core rule:

```text
The browser is never authoritative.
The web server is never privileged.
The agent is the only privileged execution surface.
```
