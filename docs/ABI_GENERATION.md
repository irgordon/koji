# KOJI ABI Syscall ABI Generation Document

Version: 1.0
Status: Draft Standard
Scope: Defines how the KOJI syscall ABI is authored, versioned, generated,
       validated, and consumed across kernel and userspace.

---

## 1. Purpose

This document defines the generation rules for the KOJI syscall ABI.

The ABI is a binding contract between:
- Ring 0 microkernel
- Ring 3 higher substrate
- Ring 3 compatibility and support layers
- build, validation, and test tooling

The ABI must be mechanically generated, versioned, and validated to prevent
drift between kernel and userspace.

> The ABI is not an implementation detail. It is a security boundary.

---

## 2. Governing Principles

The syscall ABI generation process must preserve:

- single source of truth
- deterministic output
- stable layout
- explicit versioning
- no manual duplication across languages
- no hidden semantic changes
- no privilege-boundary-specific reinterpretation

Any ABI workflow that allows kernel and userspace definitions to diverge is incorrect.

---

## 3. Canonical Source Model

The canonical source for ABI generation is a single authoritative definition
set stored under `abi/`.

Structure:

```
abi/
  KOJI_ABI_V1.h
  VERSIONING_POLICY.md
  generated/
```

For v1, the canonical source is `abi/KOJI_ABI_V1.h`.

This header defines:
- ABI identity
- version fields
- syscall numbers
- handle encoding model
- object taxonomy
- rights bitmask model
- error codes
- syscall frame layout
- IPC header layout
- fixed constants

No generated artifact may become the source of truth.

---

## 4. Generation Goals

The ABI generator must produce language-specific representations for:
- Odin kernel and low-level userspace bindings
- Go userspace bindings
- test fixtures and validation metadata if needed

Generation must preserve:
- field ordering
- field width
- signedness
- constant values
- enum discriminants
- struct layout intent
- reserved fields

Generation must not invent semantics.

---

## 5. Scope of Generated Output

The generator must emit only mechanical translations of ABI definitions.

**Allowed generated categories:**
- constants
- enums
- typedef equivalents
- struct definitions
- comments indicating source provenance
- layout assertions where the target language supports them

**Not allowed:**
- helper logic
- wrappers with policy
- behavioral abstractions
- inferred defaults not present in the source ABI
- language-specific reinterpretation of rights, handles, or status codes

Generated bindings must remain thin.

---

## 6. Required Generated Targets

### 6.1 Odin

Output: `abi/generated/odin/abi_generated.odin`

Must contain:
- ABI constants
- object type constants
- rights constants
- error constants
- syscall number constants
- translated struct definitions
- explicit integer widths

### 6.2 Go

Output: `abi/generated/go/abi_generated.go`

Must contain:
- constants
- typed integer aliases
- struct definitions with exact-width fields
- symmetric compile-time size assertions
- comments indicating generated status

### 6.3 Validation Metadata

Output: `abi/generated/meta/abi_manifest.json`

Includes:
- source file hash
- generator version
- ABI version tuple
- known struct sizes
- known field offsets (with explicit trust qualification)

This metadata must not become normative. It exists to support validation.

---

## 7. Generator Input Rules

The generator treats the canonical ABI header as a strict input contract.

It must parse and preserve:
- `#define` constants via typed integer macros
- fixed-width integer types
- typedef aliases
- plain structs with stable field order

**Unsupported constructs** (generator must fail on):
- enums
- unions
- anonymous structs
- function-like macros
- platform-conditional ABI shapes
- macros with non-deterministic evaluation
- plain integer literals exceeding uint32 range (must use typed macro)

If the source header contains unsupported constructs, generation must fail.

---

## 8. Layout Rules

### 8.1 Struct Field Order

Field order must remain identical to the canonical source.

### 8.2 Width Preservation

The following must remain exact:

| C type     | Width          |
|------------|----------------|
| `uint8_t`  | 8-bit unsigned |
| `uint16_t` | 16-bit unsigned |
| `uint32_t` | 32-bit unsigned |
| `uint64_t` | 64-bit unsigned |
| `int32_t`  | 32-bit signed  |
| `int64_t`  | 64-bit signed  |

No widening or narrowing is allowed.

### 8.3 Reserved Fields

Reserved fields must be preserved exactly.
They are part of the ABI contract and may not be removed for convenience.

### 8.4 Natural Alignment Requirement

All structs in `KOJI_ABI_V1.h` must satisfy natural alignment — every field
must sit at an offset that is a multiple of its own width.

Generator behavior on violation:
- generation must fail, or
- manifest must not emit offsets for the offending struct

The generator validates natural alignment at generation time and fails hard
if a struct would require padding.

### 8.5 Handle and Opaque Types

Opaque ABI types (`koji_handle_t`, `koji_oid_t`) must remain mechanically
faithful and must not be wrapped in behavior-bearing abstractions in generated
output.

---

## 9. Comments and Provenance

Every generated file must begin with:

```
// AUTO-GENERATED from abi/KOJI_ABI_V1.h
// Do not edit manually.
// Regenerate with: python3 tools/abi/gen_abi.py
```

Generated files must not be hand-edited. Manual changes are overwritten on
regeneration and are treated as invalid.

---

## 10. Syscall ABI Content Model

The generation process must preserve the syscall ABI model exactly.

Current ABI categories:

| Section | Content |
|---------|---------|
| 10.1 | Identity and version |
| 10.2 | Global limits and sentinels |
| 10.3 | Return contract (`koji_sysret_t`) |
| 10.4 | Object taxonomy |
| 10.5 | Rights model |
| 10.6 | Handle model (bit layout preserved in comments) |
| 10.7 | Object identity |
| 10.8 | Error model |
| 10.9 | Syscall numbers |
| 10.10 | Syscall frame |
| 10.11 | IPC envelope |

---

## 11. Generation Pipeline

The ABI generation pipeline must be deterministic.

Sequence:
1. read canonical header
2. parse constants, typedefs, and structs
3. validate supported construct set
4. validate natural alignment for all structs
5. generate target-language bindings
6. run layout and value checks
7. write generated outputs
8. emit manifest with explicit offset trust qualification
9. fail if any mismatch is detected

The generator must not silently skip unsupported definitions.

---

## 12. Validation Requirements

ABI generation is incomplete without validation.

### 12.1 Constant Validation

Verify emitted constants match source values exactly.

### 12.2 Struct Shape Validation

For each generated struct, validate:
- field count
- field order
- field widths
- natural alignment (fail generation if violated)

### 12.3 Target Compilation Validation

Generated Odin bindings must parse in Odin.
Generated Go bindings must compile in Go (size assertions enforce layout).

### 12.4 Drift Detection

Generation must fail if:
- source changed but generated output was not refreshed
- generated output changed without source change
- manifest hash mismatch exists

---

## 13. Change Control

Any change to `abi/KOJI_ABI_V1.h` must be treated as a contract change.

Required:
- rationale
- compatibility impact
- version impact (per `VERSIONING_POLICY.md`)
- regeneration of all targets
- validation pass
- review of all consumers

Weak justification is insufficient.

---

## 14. Generator Behavior on Failure

The generator must fail hard when encountering:
- unsupported source construct
- natural alignment violation in any struct
- conflicting names
- impossible type mapping
- target emission inconsistency
- validation mismatch

No partial generation is allowed. If one target cannot be produced correctly,
the entire generation step fails.

---

## 15. Repository Layout

```
abi/
  KOJI_ABI_V1.h
  VERSIONING_POLICY.md
  generated/
    odin/abi_generated.odin
    go/abi_generated.go
    meta/abi_manifest.json
tools/
  abi/
    gen_abi.py
```

Rules:
- canonical source under `abi/`
- generator under `tools/abi/`
- generated outputs under `abi/generated/`
- kernel and userspace consume generated artifacts, not hand-maintained duplicates

---

## 16. Usage

Normal mode (regenerate):

```
python3 tools/abi/gen_abi.py
```

Check mode (verify current, used by CI):

```
python3 tools/abi/gen_abi.py --check
```

---

## 17. Testing Requirements

The ABI generation workflow must include:
- parsing tests
- constant translation tests
- struct translation tests
- natural alignment validation tests
- regression tests for known layouts
- drift detection tests
- target parse/compile smoke tests

Tests must be deterministic and isolated.

---

## 18. Review Standard

An ABI generation change is acceptable only if a reviewer can determine:

1. what source changed
2. what generated output changed
3. whether constant values changed
4. whether struct layouts changed
5. whether compatibility impact is documented

If a reviewer cannot do this quickly and mechanically, the workflow is too opaque.

---

## 19. Generator Scope (v1)

For v1, the generator supports only the constructs present in the current header:
- typed integer macros
- fixed-width typedefs
- plain structs with fixed-width fields

Do not generalize beyond the actual ABI source until required.
Deletion bias applies here. Keep the generator narrow.

---

## Final Principle

The syscall ABI must be authored once, generated mechanically, and validated everywhere.

A syscall boundary that is duplicated manually is already drifting.
