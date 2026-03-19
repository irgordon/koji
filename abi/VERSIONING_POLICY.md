# KOJI ABI Versioning Policy

## Version Format

`MAJOR.MINOR.PATCH`

## Rules

- **MAJOR** — breaking change: syscall removed, struct layout changed, handle encoding changed.
- **MINOR** — additive: new syscall, new object type, new rights bit, new error code.
- **PATCH** — documentation or comment-only changes. No binary impact.

## Constraints

- Syscall numbers are never reused after removal.
- Object type discriminants are never reused.
- Error code values are never reused.
- Rights bit positions are never reused.
- Handle encoding layout (gen/index split) changes require MAJOR bump.
- Reserved fields must not be consumed without a MINOR bump.

## Compatibility

- MINOR bumps are backward-compatible: old userspace on new kernel works.
- MAJOR bumps are breaking: requires coordinated rebuild.
- The ABI version is reported via `koji_syscall_abi_info`.
