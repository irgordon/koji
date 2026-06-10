# Koji Documentation

The repository root README is the marketing-forward project introduction.

Start here for engineering governance:

1. `ARCHITECTURE.md`
2. `INVARIANTS.md`
3. `SECURITY.md`
4. `CODING_STYLE.md`
5. `PHASEMAP.md`
6. `AGENTS.md`

For implementation inventory, use:

- `wiki/Developer/Architectural-Inventory.md`
- `wiki/Developer/Backend-Inventory.md`
- `wiki/Developer/Frontend-Inventory.md`
- `wiki/Developer/Phase-History.md`

For release readiness and security review, use:

- `wiki/Operations/Release-Candidate-Checklist.md`
- `wiki/Security/Security-Review.md`
- `wiki/Security/Threat-Model.md`

Core invariant:

```text
The browser is never authoritative.
The web server is never privileged.
The agent is the only privileged execution surface.
```
