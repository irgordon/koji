// ============================================================
// KOJI Kernel — Capability Table
//
// The capability table is the kernel's central authority structure.
// Every kernel object is accessed only through a capability entry.
//
// Invariants (from ENGINEERING_PRINCIPLES.md):
//   - Rights are monotonically non-increasing (drop only, never add)
//   - Handle generation prevents use-after-free
//   - No capability may exist without a backing kernel object
//   - All access checks are O(1) via table lookup
//
// v1: Fixed-size table, single address space, no per-process scoping.
//     Per-process capability spaces come in v2.
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Kernel Object (tagged union root) ----
// Each concrete object type will be defined in its own module.
// For scaffolding, we use an opaque pointer + type tag.

Kernel_Object :: struct {
	obj_type:  abi.Obj_Type,
	ref_count: u32,
	data:      rawptr,        // concrete object (Process, Thread, Channel, etc.)
}

// ---- Capability Entry ----

Cap_Entry :: struct {
	object:     ^Kernel_Object,   // nil = slot is free
	rights:     abi.Rights,
	generation: u8,               // must match handle's generation bits
	_pad:       [3]u8,
}

// ---- Capability Table ----

// v1: 4096 slots. Enough for early bring-up.
// Production sizing is a policy decision for the higher substrate.
CAP_TABLE_SIZE :: 4096

Cap_Table :: struct {
	entries:   [CAP_TABLE_SIZE]Cap_Entry,
	free_head: u32,                // index of first free slot (simple free list)
}

// Global table (single address space, v1)
g_cap_table: Cap_Table

// ---- Init ----

cap_table_init :: proc "c" () {
	// Zero all entries, chain free list
	for i in 0..<u32(CAP_TABLE_SIZE) {
		g_cap_table.entries[i] = Cap_Entry{
			object     = nil,
			rights     = abi.RIGHT_NONE,
			generation = 0,
			_pad       = {0, 0, 0},
		}
	}
	g_cap_table.free_head = 0
}

// ---- Allocate ----
// Returns HANDLE_INVALID if table is full.

cap_alloc :: proc "c" (obj: ^Kernel_Object, rights: abi.Rights) -> abi.Handle {
	// Linear scan for free slot (v1 simplicity; free-list optimization in v2)
	for i in 0..<u32(CAP_TABLE_SIZE) {
		if g_cap_table.entries[i].object == nil {
			gen := g_cap_table.entries[i].generation
			g_cap_table.entries[i] = Cap_Entry{
				object     = obj,
				rights     = rights,
				generation = gen,
				_pad       = {0, 0, 0},
			}
			return abi.handle_make(i, gen)
		}
	}
	return abi.HANDLE_INVALID
}

// ---- Lookup ----
// Validates handle, checks generation, returns entry or nil.

cap_lookup :: proc "c" (h: abi.Handle) -> ^Cap_Entry {
	if h == abi.HANDLE_INVALID {
		return nil
	}
	idx := abi.handle_index(h)
	gen := abi.handle_gen(h)
	if idx >= CAP_TABLE_SIZE {
		return nil
	}
	entry := &g_cap_table.entries[idx]
	if entry.object == nil {
		return nil
	}
	if entry.generation != gen {
		return nil  // stale handle — use-after-free attempt
	}
	return entry
}

// ---- Rights Check ----

cap_has_rights :: proc "c" (entry: ^Cap_Entry, required: abi.Rights) -> bool {
	if entry == nil {
		return false
	}
	return (entry.rights & required) == required
}

// ---- Close / Free ----
// Invalidates the capability slot and bumps generation.

cap_close :: proc "c" (h: abi.Handle) -> abi.Status {
	entry := cap_lookup(h)
	if entry == nil {
		return abi.ERR_INVALID_HANDLE
	}

	// Decrement object ref count
	if entry.object != nil {
		entry.object.ref_count -= 1
		// Object destruction is the responsibility of the object's own
		// lifecycle manager (not the cap table). When ref_count hits 0,
		// the object subsystem reclaims it.
	}

	entry.object = nil
	entry.rights = abi.RIGHT_NONE
	// Bump generation to invalidate any outstanding handles to this slot
	entry.generation += 1

	return abi.OK
}

// ---- Duplicate ----
// Creates a new capability with equal or fewer rights.

cap_duplicate :: proc "c" (h: abi.Handle, new_rights: abi.Rights) -> (abi.Handle, abi.Status) {
	src := cap_lookup(h)
	if src == nil {
		return abi.HANDLE_INVALID, abi.ERR_INVALID_HANDLE
	}
	if !cap_has_rights(src, abi.RIGHT_DUPLICATE) {
		return abi.HANDLE_INVALID, abi.ERR_ACCESS_DENIED
	}

	// Rights can only be narrowed, never widened
	effective := src.rights & new_rights

	// Bump ref count
	src.object.ref_count += 1

	new_handle := cap_alloc(src.object, effective)
	if new_handle == abi.HANDLE_INVALID {
		src.object.ref_count -= 1
		return abi.HANDLE_INVALID, abi.ERR_NO_MEMORY
	}
	return new_handle, abi.OK
}
