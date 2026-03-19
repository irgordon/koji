#!/usr/bin/env python3
"""
KOJI ABI Generator
==================
Mechanically translates abi/KOJI_ABI_V1.h into language-specific bindings.

Usage:
    python3 tools/abi/gen_abi.py            # regenerate all targets
    python3 tools/abi/gen_abi.py --check    # verify current outputs match (CI mode)

Supported constructs (generator fails hard on anything else):
    - typedef unsigned/signed <width> koji_<type>  (primitive typedefs)
    - typedef koji_<base> koji_<alias>             (type aliases)
    - #define NAME ((type)value)                   (typed constants)
    - typedef struct { ... } name                  (plain structs, fixed-width fields)

Generated targets:
    abi/generated/odin/abi_generated.odin
    abi/generated/go/abi_generated.go
    abi/generated/meta/abi_manifest.json
"""

import argparse
import hashlib
import json
import os
import re
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

# ---------------------------------------------------------------------------
# Generator metadata
# ---------------------------------------------------------------------------

GENERATOR_VERSION = "1.0.0"

# ---------------------------------------------------------------------------
# Repository paths (relative to this script's location: tools/abi/)
# ---------------------------------------------------------------------------

SCRIPT_DIR   = Path(__file__).resolve().parent
REPO_ROOT    = SCRIPT_DIR.parent.parent
ABI_HEADER   = REPO_ROOT / "abi" / "KOJI_ABI_V1.h"
OUT_ODIN     = REPO_ROOT / "abi" / "generated" / "odin" / "abi_generated.odin"
OUT_GO       = REPO_ROOT / "abi" / "generated" / "go"   / "abi_generated.go"
OUT_MANIFEST = REPO_ROOT / "abi" / "generated" / "meta" / "abi_manifest.json"


# ---------------------------------------------------------------------------
# Data model
# ---------------------------------------------------------------------------

@dataclass
class PrimitiveTypedef:
    """e.g. typedef unsigned int koji_u32;"""
    c_name:    str   # koji_u32
    c_type:    str   # "unsigned int"
    width:     int   # 32
    signed:    bool  # False for unsigned


@dataclass
class AliasTypedef:
    """e.g. typedef koji_u32 koji_handle_t;"""
    c_name:    str   # koji_handle_t
    base_name: str   # koji_u32


@dataclass
class Constant:
    """e.g. #define KOJI_ABI_MAGIC ((koji_u32)0x4B4F4A49)"""
    name:        str
    raw_value:   str   # string form, e.g. "0x4B4F4A49" or "1"
    int_value:   int
    cast_type:   str   # e.g. "koji_u32", or "" for untyped
    comment:     str   # inline comment if any


@dataclass
class StructField:
    name:      str
    c_type:    str   # resolved to a primitive or alias
    comment:   str


@dataclass
class Struct:
    c_name:   str              # e.g. "koji_syscall_frame"
    typedef:  str              # e.g. "koji_syscall_frame_t"
    fields:   list[StructField] = field(default_factory=list)


@dataclass
class ABI:
    primitives: list[PrimitiveTypedef]   = field(default_factory=list)
    aliases:    list[AliasTypedef]       = field(default_factory=list)
    constants:  list[Constant]           = field(default_factory=list)
    structs:    list[Struct]             = field(default_factory=list)


# ---------------------------------------------------------------------------
# C header parser
# ---------------------------------------------------------------------------

# Primitive C type → (width, signed)
PRIMITIVE_MAP = {
    "unsigned char":       (8,  False),
    "unsigned short":      (16, False),
    "unsigned int":        (32, False),
    "unsigned long long":  (64, False),
    "signed int":          (32, True),
    "signed long long":    (64, True),
}


def _strip_block_comments(src: str) -> str:
    """Remove /* ... */ block comments, preserving line structure."""
    return re.sub(r"/\*.*?\*/", "", src, flags=re.DOTALL)


def _extract_inline_comment(line: str) -> tuple[str, str]:
    """Split line into (code, inline_comment)."""
    m = re.search(r"/\*(.+?)\*/", line)
    if m:
        return line[:m.start()].rstrip(), m.group(1).strip()
    return line, ""


def parse_header(path: Path) -> ABI:
    raw = path.read_text(encoding="utf-8")
    abi = ABI()

    # ---- strip #ifndef / #define guard and #endif ----
    src = re.sub(r"^\s*#ifndef\s+\S+\s*$", "", raw, flags=re.MULTILINE)
    src = re.sub(r"^\s*#define\s+KOJI_ABI_V1_H\s*$", "", src, flags=re.MULTILINE)
    src = re.sub(r"^\s*#endif.*$", "", src, flags=re.MULTILINE)

    # ---- parse section-by-section ----
    lines = src.splitlines()
    i = 0
    while i < len(lines):
        line = lines[i].strip()

        # ----------------------------------------------------------------
        # Primitive typedef:  typedef unsigned int koji_u32;
        # ----------------------------------------------------------------
        m = re.match(
            r"typedef\s+((?:unsigned|signed)\s+(?:char|short|int|long long))\s+(koji_\w+)\s*;",
            line)
        if m:
            c_type, c_name = m.group(1), m.group(2)
            if c_type not in PRIMITIVE_MAP:
                _fail(f"Unsupported primitive C type: {c_type!r} (line {i+1})")
            width, signed = PRIMITIVE_MAP[c_type]
            abi.primitives.append(PrimitiveTypedef(c_name, c_type, width, signed))
            i += 1
            continue

        # ----------------------------------------------------------------
        # Alias typedef:  typedef koji_u32 koji_handle_t;
        # ----------------------------------------------------------------
        m = re.match(r"typedef\s+(koji_\w+)\s+(koji_\w+)\s*;", line)
        if m:
            base_name, c_name = m.group(1), m.group(2)
            abi.aliases.append(AliasTypedef(c_name, base_name))
            i += 1
            continue

        # ----------------------------------------------------------------
        # Struct: typedef struct koji_name { ... } koji_name_t;
        # ----------------------------------------------------------------
        m = re.match(r"typedef\s+struct\s+(\w+)\s*\{", line)
        if m:
            struct_c_name = m.group(1)
            struct_fields: list[StructField] = []
            i += 1
            while i < len(lines):
                fline = lines[i].strip()
                if fline.startswith("}"):
                    # extract typedef name
                    mt = re.match(r"\}\s*(koji_\w+)\s*;", fline)
                    if not mt:
                        _fail(f"Malformed struct closing at line {i+1}: {fline!r}")
                    typedef_name = mt.group(1)
                    break

                # skip blank / comment-only lines inside struct
                code, comment = _extract_inline_comment(fline)
                code = code.strip()
                if not code or code.startswith("/*") or code.startswith("//"):
                    i += 1
                    continue

                # field:  koji_u64 user_rip;
                mf = re.match(r"(koji_\w+)\s+(\w+)\s*;", code)
                if not mf:
                    _fail(f"Unsupported struct field at line {i+1}: {fline!r}")
                struct_fields.append(StructField(mf.group(2), mf.group(1), comment))
                i += 1
            else:
                _fail(f"Unterminated struct {struct_c_name}")

            abi.structs.append(Struct(struct_c_name, typedef_name, struct_fields))
            i += 1
            continue

        # ----------------------------------------------------------------
        # #define constant  (typed):  #define NAME ((type)value)
        # #define constant (untyped): #define NAME value
        # Skip section headers, guards, includes
        # ----------------------------------------------------------------
        m = re.match(r"#define\s+(\w+)\s+(.*)", line)
        if m:
            name, rest = m.group(1), m.group(2).strip()

            # Strip inline comment
            code_part, comment = _extract_inline_comment(rest)
            code_part = code_part.strip()

            # Reject include guards and bare string tokens
            if name in ("KOJI_ABI_V1_H",):
                i += 1
                continue

            # Pattern: ((koji_type)value)  — greedy + anchored to handle (1 << 0) etc.
            mt = re.match(r"\(\s*\(\s*(koji_\w+)\s*\)\s*(.+)\s*\)\s*$", code_part)
            if mt:
                cast_type, raw_val = mt.group(1), mt.group(2)
                # raw_val may itself be an expression like (1 << 0)
                int_val = _eval_int(raw_val, name)
                abi.constants.append(Constant(name, raw_val, int_val, cast_type, comment))
                i += 1
                continue

            # Pattern: plain integer (no cast)  — e.g. #define KOJI_SYSCALL_COUNT 20
            # Only accept if it looks like a numeric literal
            if re.match(r"^[0-9]", code_part) or re.match(r"^0x", code_part):
                int_val = _eval_int(code_part, name)
                abi.constants.append(Constant(name, code_part, int_val, "", comment))
                i += 1
                continue

            # Section header defines or unrecognised — skip
            i += 1
            continue

        i += 1

    _validate(abi)
    return abi


def _eval_int(expr: str, name: str) -> int:
    """Safely evaluate a restricted C constant integer expression.

    Supported forms only:
        decimal integer literal         e.g.  26, 255, 4096
        hex integer literal             e.g.  0xFF000000, 0x4B4F4A49
        left-shift expression           e.g.  (1 << 7)
        bitwise-OR of the above         e.g.  0x00FF | (1 << 8)

    No eval()/exec() is used.  Any unrecognised pattern is a hard error.
    """
    expr = expr.strip()

    # ---- simple literal (decimal or hex) ----
    try:
        return int(expr, 0)
    except ValueError:
        pass

    # ---- (a << b) ----
    m = re.match(r'^\(\s*(\d+)\s*<<\s*(\d+)\s*\)$', expr)
    if m:
        return int(m.group(1)) << int(m.group(2))

    # ---- term | term | ... where each term is a literal or (a << b) ----
    if "|" in expr:
        result = 0
        for part in expr.split("|"):
            result |= _eval_int(part.strip(), name)
        return result

    _fail(f"Cannot evaluate constant {name!r} = {expr!r}: "
          f"unsupported expression (only literals, shifts, and bitwise-OR allowed)")


def _validate(abi: ABI) -> None:
    """Structural validation: natural alignment for all structs."""
    prim_widths: dict[str, int] = {}
    for p in abi.primitives:
        prim_widths[p.c_name] = p.width // 8  # bytes

    # Resolve alias → width
    def resolve_width(c_type: str) -> Optional[int]:
        if c_type in prim_widths:
            return prim_widths[c_type]
        for a in abi.aliases:
            if a.c_name == c_type:
                return resolve_width(a.base_name)
        return None

    for struct in abi.structs:
        offset = 0
        for f in struct.fields:
            w = resolve_width(f.c_type)
            if w is None:
                _fail(f"Unknown field type {f.c_type!r} in struct {struct.c_name}")
            if offset % w != 0:
                _fail(
                    f"Natural alignment violation in struct {struct.c_name}, "
                    f"field {f.name}: offset {offset} is not a multiple of {w}"
                )
            offset += w


def _fail(msg: str) -> None:
    print(f"[gen_abi] ERROR: {msg}", file=sys.stderr)
    sys.exit(1)


# ---------------------------------------------------------------------------
# Odin emitter
# ---------------------------------------------------------------------------

# koji_u8 → Odin type
_ODIN_TYPES = {
    "koji_u8":  "u8",
    "koji_u16": "u16",
    "koji_u32": "u32",
    "koji_u64": "u64",
    "koji_i32": "i32",
    "koji_i64": "i64",
}


def _odin_type(c_type: str, abi: ABI, aliases: dict[str, str]) -> str:
    """Resolve a koji C type to an Odin type name."""
    if c_type in _ODIN_TYPES:
        return _ODIN_TYPES[c_type]
    # Check if it's an alias we've already mapped
    if c_type in aliases:
        return aliases[c_type]
    # Walk alias chain
    for a in abi.aliases:
        if a.c_name == c_type:
            return _odin_type(a.base_name, abi, aliases)
    _fail(f"Cannot map C type {c_type!r} to Odin")


def emit_odin(abi: ABI, source_path: Path) -> str:
    lines: list[str] = []
    w = lines.append

    w("// AUTO-GENERATED from abi/KOJI_ABI_V1.h")
    w("// Do not edit manually.")
    w("// Regenerate with: python3 tools/abi/gen_abi.py")
    w("package koji_abi")
    w("")

    # ---- Collect alias → Odin name mappings ----
    # Primitive aliases get mapped to their base Odin type but typed distinctly
    # or simply as a type alias, depending on semantic role.
    # We emit every koji_*_t alias as a `distinct` type for type safety.
    alias_odin: dict[str, str] = {}   # c_name → odin_name (CamelCase or base)

    # Map primitive koji types
    for a in abi.aliases:
        # Derive a clean Odin name from the C name
        # koji_handle_t → Handle, koji_obj_type_t → Obj_Type, koji_rights_t → Rights
        odin_name = _c_alias_to_odin(a.c_name)
        alias_odin[a.c_name] = odin_name

    # ---- Emit type aliases ----
    w("// ---- Type Aliases ----")
    w("")
    for a in abi.aliases:
        base = _ODIN_TYPES.get(a.base_name, alias_odin.get(a.base_name, a.base_name))
        odin_name = alias_odin[a.c_name]
        w(f"{odin_name} :: distinct {base}")
    w("")

    # ---- Emit constants, grouped by prefix ----
    w("// ---- Constants ----")
    w("")
    prev_prefix = ""
    for c in abi.constants:
        prefix = _constant_prefix(c.name)
        if prefix != prev_prefix and prev_prefix:
            w("")
        prev_prefix = prefix

        odin_name = _c_const_to_odin(c.name)
        # Determine value representation
        if c.cast_type:
            cast_odin = alias_odin.get(c.cast_type, _ODIN_TYPES.get(c.cast_type, c.cast_type))
            val = f"{cast_odin}({_format_int(c.int_value, c.raw_value)})"
        else:
            val = str(c.int_value)

        comment_part = f"   // {c.comment}" if c.comment else ""
        w(f"{odin_name:<28} :: {val}{comment_part}")

    w("")

    # ---- Emit handle helpers (bit manipulation only, no policy) ----
    w("// ---- Handle Helpers (bit manipulation) ----")
    w("")
    w("handle_index :: #force_inline proc(h: Handle) -> u32 {")
    w("\treturn u32(h) & HANDLE_INDEX_MASK")
    w("}")
    w("")
    w("handle_gen :: #force_inline proc(h: Handle) -> u8 {")
    w("\treturn u8((u32(h) & HANDLE_GEN_MASK) >> HANDLE_GEN_SHIFT)")
    w("}")
    w("")
    w("handle_make :: #force_inline proc(index: u32, gen: u8) -> Handle {")
    w("\treturn Handle((u32(gen) << HANDLE_GEN_SHIFT) | (index & HANDLE_INDEX_MASK))")
    w("}")
    w("")

    # ---- Emit structs ----
    w("// ---- Structs ----")
    w("")
    for s in abi.structs:
        odin_name = _c_struct_to_odin(s.typedef)
        w(f"{odin_name} :: struct #packed {{")
        for f in s.fields:
            odin_t = _odin_type(f.c_type, abi, alias_odin)
            comment_part = f"   // {f.comment}" if f.comment else ""
            name_col = f.name + ":"
            w(f"\t{name_col:<17} {odin_t},{comment_part}")
        w("}")
        w("")

    # ---- Emit layout assertions ----
    w("// ---- Layout Assertions ----")
    w("")
    for s in abi.structs:
        odin_name = _c_struct_to_odin(s.typedef)
        size = _struct_size(s, abi)
        w(f"#assert(size_of({odin_name}) == {size})")
    for a in abi.aliases:
        base = a.base_name
        size = _primitive_size(base, abi)
        if size:
            odin_name = alias_odin[a.c_name]
            w(f"#assert(size_of({odin_name}) == {size})")
    w("")

    return "\n".join(lines)


def _c_alias_to_odin(c_name: str) -> str:
    """koji_handle_t → Handle, koji_obj_type_t → Obj_Type, koji_syscall_t → Syscall_Num"""
    name = c_name
    if name.startswith("koji_"):
        name = name[5:]
    if name.endswith("_t"):
        name = name[:-2]
    # Special renames for clarity
    rename = {
        "syscall": "Syscall_Num",
        "status":  "Status",
    }
    if name in rename:
        return rename[name]
    # CamelCase: obj_type → Obj_Type
    return "_".join(p.capitalize() for p in name.split("_"))


def _c_struct_to_odin(typedef: str) -> str:
    """koji_syscall_frame_t → Syscall_Frame, koji_ipc_header_t → Ipc_Header"""
    name = typedef
    if name.startswith("koji_"):
        name = name[5:]
    if name.endswith("_t"):
        name = name[:-2]
    return "_".join(p.capitalize() for p in name.split("_"))


def _c_const_to_odin(name: str) -> str:
    """KOJI_ABI_MAGIC → ABI_MAGIC, KOJI_SYS_HANDLE_CLOSE → SYS_HANDLE_CLOSE"""
    if name.startswith("KOJI_"):
        return name[5:]
    return name


def _constant_prefix(name: str) -> str:
    """Group constants by their first segment after KOJI_: ABI, HANDLE, OBJ, ERR, SYS, etc."""
    parts = name.split("_")
    if parts[0] == "KOJI" and len(parts) >= 2:
        return parts[1]
    return parts[0]


def _format_int(value: int, raw: str) -> str:
    """Prefer hex for hex-looking originals."""
    if "0x" in raw.lower() or "0X" in raw.lower():
        return hex(value)
    return str(value)


def _primitive_size(c_type: str, abi: ABI) -> Optional[int]:
    for p in abi.primitives:
        if p.c_name == c_type:
            return p.width // 8
    for a in abi.aliases:
        if a.c_name == c_type:
            return _primitive_size(a.base_name, abi)
    return None


def _struct_size(s: Struct, abi: ABI) -> int:
    total = 0
    for f in s.fields:
        sz = _primitive_size(f.c_type, abi)
        if sz is None:
            _fail(f"Cannot determine size of field {f.name} (type {f.c_type})")
        total += sz
    return total


# ---------------------------------------------------------------------------
# Go emitter
# ---------------------------------------------------------------------------

_GO_TYPES = {
    "koji_u8":  "uint8",
    "koji_u16": "uint16",
    "koji_u32": "uint32",
    "koji_u64": "uint64",
    "koji_i32": "int32",
    "koji_i64": "int64",
}


def _go_type(c_type: str, abi: ABI, alias_go: dict[str, str]) -> str:
    if c_type in _GO_TYPES:
        return _GO_TYPES[c_type]
    if c_type in alias_go:
        return alias_go[c_type]
    for a in abi.aliases:
        if a.c_name == c_type:
            return _go_type(a.base_name, abi, alias_go)
    _fail(f"Cannot map C type {c_type!r} to Go")


def _c_alias_to_go(c_name: str) -> str:
    """koji_handle_t → Handle, koji_obj_type_t → ObjType"""
    name = c_name
    if name.startswith("koji_"):
        name = name[5:]
    if name.endswith("_t"):
        name = name[:-2]
    rename = {
        "syscall": "SyscallNum",
        "status":  "Status",
    }
    if name in rename:
        return rename[name]
    # CamelCase without underscores: obj_type → ObjType
    return "".join(p.capitalize() for p in name.split("_"))


def _c_const_to_go(name: str) -> str:
    """KOJI_ABI_MAGIC → ABIMagic (PascalCase)"""
    if name.startswith("KOJI_"):
        name = name[5:]
    # KOJI_ERR_COUNT → ErrCount, KOJI_SYS_HANDLE_CLOSE → SysHandleClose
    return "".join(p.capitalize() for p in name.split("_"))


def _c_struct_to_go(typedef: str) -> str:
    """koji_syscall_frame_t → SyscallFrame"""
    name = typedef
    if name.startswith("koji_"):
        name = name[5:]
    if name.endswith("_t"):
        name = name[:-2]
    return "".join(p.capitalize() for p in name.split("_"))


def emit_go(abi: ABI, source_path: Path) -> str:
    lines: list[str] = []
    w = lines.append

    w("// AUTO-GENERATED from abi/KOJI_ABI_V1.h")
    w("// Do not edit manually.")
    w("// Regenerate with: python3 tools/abi/gen_abi.py")
    w("")
    w("package kojiabi")
    w("")
    w('import "unsafe"')
    w("")

    alias_go: dict[str, str] = {}
    for a in abi.aliases:
        alias_go[a.c_name] = _c_alias_to_go(a.c_name)

    # ---- Type definitions ----
    w("// ---- Type Aliases ----")
    w("")
    for a in abi.aliases:
        base = _GO_TYPES.get(a.base_name, alias_go.get(a.base_name, a.base_name))
        go_name = alias_go[a.c_name]
        w(f"type {go_name} {base}")
    w("")

    # ---- Constants ----
    w("// ---- Constants ----")
    w("")
    w("const (")
    prev_prefix = ""
    for c in abi.constants:
        prefix = _constant_prefix(c.name)
        if prefix != prev_prefix and prev_prefix:
            w("")
        prev_prefix = prefix

        go_name = _c_const_to_go(c.name)
        if c.cast_type:
            cast_go = alias_go.get(c.cast_type, _GO_TYPES.get(c.cast_type, c.cast_type))
            val = f"{cast_go}({_format_int(c.int_value, c.raw_value)})"
        else:
            val = str(c.int_value)

        comment_part = f" // {c.comment}" if c.comment else ""
        w(f"\t{go_name} = {val}{comment_part}")
    w(")")
    w("")

    # ---- Structs ----
    w("// ---- Structs ----")
    w("")
    for s in abi.structs:
        go_name = _c_struct_to_go(s.typedef)
        w(f"type {go_name} struct {{")
        for f in s.fields:
            go_t = _go_type(f.c_type, abi, alias_go)
            comment_part = f" // {f.comment}" if f.comment else ""
            fname = "".join(p.capitalize() for p in f.name.split("_"))
            w(f"\t{fname:<16} {go_t}{comment_part}")
        w("}")
        w("")

    # ---- Size assertions (two-sided: both directions catch mismatch) ----
    w("// ---- Layout Assertions ----")
    w("")
    w("var (")
    for s in abi.structs:
        go_name = _c_struct_to_go(s.typedef)
        size = _struct_size(s, abi)
        w(f"\t_ [{size} - unsafe.Sizeof({go_name}{{}})]struct{{}} // fails if struct > {size} bytes")
        w(f"\t_ [unsafe.Sizeof({go_name}{{}}) - {size}]struct{{}} // fails if struct < {size} bytes")
    w(")")
    w("")

    return "\n".join(lines)


# ---------------------------------------------------------------------------
# Manifest emitter
# ---------------------------------------------------------------------------

def emit_manifest(abi: ABI, source_path: Path, source_hash: str) -> str:
    structs_meta = []
    for s in abi.structs:
        size = _struct_size(s, abi)
        offset = 0
        fields_meta = []
        for f in s.fields:
            fsz = _primitive_size(f.c_type, abi)
            fields_meta.append({
                "name":   f.name,
                "type":   f.c_type,
                "offset": offset,
                "size":   fsz,
            })
            offset += fsz
        structs_meta.append({
            "c_name":  s.c_name,
            "typedef": s.typedef,
            "size":    size,
            "fields":  fields_meta,
        })

    constants_meta = {
        _c_const_to_go(c.name): {
            "c_name": c.name,
            "value":  c.int_value,
            "type":   c.cast_type,
        }
        for c in abi.constants
    }

    manifest = {
        "generator_version": GENERATOR_VERSION,
        "abi_version": {
            "major": next(c.int_value for c in abi.constants if c.name == "KOJI_ABI_VERSION_MAJOR"),
            "minor": next(c.int_value for c in abi.constants if c.name == "KOJI_ABI_VERSION_MINOR"),
            "patch": next(c.int_value for c in abi.constants if c.name == "KOJI_ABI_VERSION_PATCH"),
        },
        "source_file":  "abi/KOJI_ABI_V1.h",
        "source_sha256": source_hash,
        "note": "Field offsets are informational. Trust is placed in generated binding "
                "assertions (Odin #assert, Go unsafe.Sizeof checks), not this manifest.",
        "structs":    structs_meta,
        "constants":  constants_meta,
    }
    return json.dumps(manifest, indent=2) + "\n"


# ---------------------------------------------------------------------------
# File I/O helpers
# ---------------------------------------------------------------------------

def write_or_check(path: Path, content: str, check_mode: bool) -> bool:
    """
    In regenerate mode: write content to path (create dirs as needed).
    In check mode: return True if content matches existing file, False otherwise.
    """
    path.parent.mkdir(parents=True, exist_ok=True)

    if check_mode:
        if not path.exists():
            print(f"[gen_abi] MISSING: {path.relative_to(REPO_ROOT)}", file=sys.stderr)
            return False
        existing = path.read_text(encoding="utf-8")
        if existing != content:
            print(f"[gen_abi] DRIFT DETECTED: {path.relative_to(REPO_ROOT)}", file=sys.stderr)
            return False
        return True
    else:
        path.write_text(content, encoding="utf-8")
        print(f"[gen_abi] wrote {path.relative_to(REPO_ROOT)}")
        return True


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="KOJI ABI generator",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__,
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Verify that current generated files match the canonical header (CI mode).",
    )
    args = parser.parse_args()

    if not ABI_HEADER.exists():
        _fail(f"Canonical header not found: {ABI_HEADER}")

    source_text = ABI_HEADER.read_text(encoding="utf-8")
    source_hash = hashlib.sha256(source_text.encode("utf-8")).hexdigest()

    print(f"[gen_abi] parsing {ABI_HEADER.relative_to(REPO_ROOT)} ...")
    abi = parse_header(ABI_HEADER)
    print(
        f"[gen_abi] found {len(abi.primitives)} primitives, "
        f"{len(abi.aliases)} aliases, "
        f"{len(abi.constants)} constants, "
        f"{len(abi.structs)} structs"
    )

    odin_out     = emit_odin(abi, ABI_HEADER)
    go_out       = emit_go(abi, ABI_HEADER)
    manifest_out = emit_manifest(abi, ABI_HEADER, source_hash)

    ok = True
    ok &= write_or_check(OUT_ODIN,     odin_out,     args.check)
    ok &= write_or_check(OUT_GO,       go_out,       args.check)
    ok &= write_or_check(OUT_MANIFEST, manifest_out, args.check)

    if not ok:
        print(
            "[gen_abi] FAIL: generated files are out of date. "
            "Run: python3 tools/abi/gen_abi.py",
            file=sys.stderr,
        )
        sys.exit(1)

    if args.check:
        print("[gen_abi] OK: all generated files are up to date.")


if __name__ == "__main__":
    main()
