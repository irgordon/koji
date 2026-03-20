// ============================================================
// KOJI Kernel — CNode (Capability Node / Capability Table)
//
// The CNode is the kernel's central authority structure.
// Every kernel object is accessible ONLY through a capability
// stored in a CNode slot.  No pointer escapes to user mode.
//
// Slot Model
// ----------
//   Each slot holds one capability entry or is empty.
//   A capability = (object header pointer, rights mask, generation).
//
//   Slot 0 is a valid usable slot.
//   HANDLE_INVALID (0xFFFFFFFF) is the sole sentinel value.
//
// Handle Encoding (from ABI KOJI_ABI_V1.h §3)
// -------------------------------------------
//   bits [31:24]  generation counter  (u8, 256 generations)
//   bits [23:0]   slot index          (max 16,777,215 entries)
//
// Generation Semantics
// --------------------
//   Generation is stored in Cap_Entry and bumped in cap_close AFTER
//   clearing the slot.  The next cap_alloc at that slot inherits the
//   bumped generation, so any handle carrying the old generation value
//   fails the generation check.
//
//   u8 rollover is a bounded-risk tradeoff in v1.  After 256 reuse cycles
//   at the same slot, an old stale handle could collide with current
//   generation; this is accepted for early bring-up.
//
// Anti-Amplification Rule
// -----------------------
//   effective_rights = src.rights & requested_rights
//   A caller can never gain rights not held by the source capability.
//
// Reference Counting Integration
// ------------------------------
//   obj_init starts with ref_count = 0 (unpublished object).
//   Each successful cap_alloc publishes one capability-owned reference.
//   Internal kernel code may take temporary refs with obj_ref and must
//   release them with obj_deref.
//   cap_close clears the slot first, then calls obj_deref.
//   Final destruction happens when total ref_count reaches zero.
//
// Tests: see tests/cnode_test.odin
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Capability Entry ----

Cap_Entry :: struct {
	object:     ^Obj_Header, // nil = slot is free; non-nil = slot is occupied
	rights:     abi.Rights,
	generation: u8,          // must equal handle's generation bits for the slot to be valid
	_pad:       [3]u8,
}

// ---- Capability Table ----

// v1: 4096 slots.  Production sizing is a policy decision for the substrate.
CAP_TABLE_SIZE :: 4096

Cap_Table :: struct {
	entries: [CAP_TABLE_SIZE]Cap_Entry,
}

// Global table — single address space, v1.
g_cap_table: Cap_Table

// ---- Local freestanding-safe handle helpers ----
//
// ABI handle encoding:
//   bits [31:24] generation (u8)
//   bits [23:0] slot index (stored in lower 24 bits of u32)
handle_make :: #force_inline proc "c" (index: u32, gen: u8) -> abi.Handle {
	return abi.Handle((u32(gen) << 24) | (index & 0x00FF_FFFF))
}

handle_index :: #force_inline proc "c" (h: abi.Handle) -> u32 {
	return u32(h) & 0x00FF_FFFF
}

handle_gen :: #force_inline proc "c" (h: abi.Handle) -> u8 {
	return u8((u32(h) >> 24) & 0xFF)
}

// ---- cap_table_init ----
//
// Zeros every entry and resets all generation counters.
// Called once during kernel boot before any syscall can execute.
cap_table_init :: proc "c" () {
	for i in 0 ..< CAP_TABLE_SIZE {
		g_cap_table.entries[i] = Cap_Entry {
			object     = nil,
			rights     = abi.RIGHT_NONE,
			generation = 0,
			_pad       = {0, 0, 0},
		}
	}
}

// ---- cap_alloc ----
//
// Scans for a free slot, fills it, and returns the encoded handle.
// Returns HANDLE_INVALID if the table is full, obj is nil, or obj is not Live.
// On success, publishes one capability-owned reference via obj_ref.
cap_alloc :: proc "c" (obj: ^Obj_Header, rights: abi.Rights) -> abi.Handle {
	if obj == nil {
		return abi.HANDLE_INVALID
	}
	// Refuse to attach a capability to a non-live object.
	if !obj_is_live(obj) {
		return abi.HANDLE_INVALID
	}
	for i in 0 ..< CAP_TABLE_SIZE {
		if g_cap_table.entries[i].object == nil {
			gen := g_cap_table.entries[i].generation // inherits post-close generation
			if !obj_ref(obj) {
				return abi.HANDLE_INVALID
			}
			g_cap_table.entries[i] = Cap_Entry {
				object     = obj,
				rights     = rights,
				generation = gen,
				_pad       = {0, 0, 0},
			}
			return handle_make(u32(i), gen)
		}
	}
	return abi.HANDLE_INVALID
}

// ---- cap_lookup ----
//
// Validates the handle and returns a pointer to the live entry.
// Execution model assumption: capability table mutation/lookup is serialized
// by the v1 kernel's single-threaded, non-preemptive capability path, so
// returned ^Cap_Entry is for immediate serialized use and must not be retained.
// Returns nil for every invalid state:
//   • h == HANDLE_INVALID                  — explicit sentinel
//   • index >= CAP_TABLE_SIZE              — out-of-range index
//   • entry.object == nil                  — empty / never-allocated slot
//   • entry.generation != handle gen       — stale handle (use-after-close)
//   • !obj_is_live(entry.object)           — object is dying or dead
cap_lookup :: proc "c" (h: abi.Handle) -> ^Cap_Entry {
	// Reject the explicit sentinel value.
	if h == abi.HANDLE_INVALID {
		return nil
	}
	idx := handle_index(h)
	gen := handle_gen(h)
	if idx >= CAP_TABLE_SIZE {
		return nil
	}
	entry := &g_cap_table.entries[idx]
	if entry.object == nil {
		return nil // empty slot
	}
	if entry.generation != gen {
		return nil // stale handle — generation mismatch
	}
	if !obj_is_live(entry.object) {
		return nil // object is dying or already dead
	}
	return entry
}

// ---- cap_has_rights ----
//
// Returns true iff entry holds all bits in required.
// Nil entry always returns false.
cap_has_rights :: #force_inline proc "c" (entry: ^Cap_Entry, required: abi.Rights) -> bool {
	if entry == nil {
		return false
	}
	return (entry.rights & required) == required
}

// ---- cap_close ----
//
// Invalidates the slot (clears object, bumps generation) and
// decrements the object's reference count via obj_deref.
// If the object's ref_count reaches zero, obj_deref invokes its
// destroy_fn and marks it Dead.
//
// The slot is cleared BEFORE calling obj_deref so that if destroy_fn
// recursively closes other handles, this slot is already unreachable.
cap_close :: proc "c" (h: abi.Handle) -> abi.Status {
	entry := cap_lookup(h)
	if entry == nil {
		return abi.ERR_INVALID_HANDLE
	}

	obj := entry.object // save before clearing

	// Clear the slot and bump generation first — the old handle is
	// immediately unreachable from this point forward.
	entry.object     = nil
	entry.rights     = abi.RIGHT_NONE
	entry.generation += 1 // u8 wraps naturally: 0xFF → 0x00

	// Decrement ref count; triggers destruction if this was the last ref.
	obj_deref(obj)

	return abi.OK
}

// ---- cap_duplicate ----
//
// Creates a new capability to the same object with equal or fewer rights.
// Anti-amplification: effective = src.rights & requested_rights.
//
// The caller must hold RIGHT_DUPLICATE on the source handle.
cap_duplicate :: proc "c" (h: abi.Handle, requested_rights: abi.Rights) -> (abi.Handle, abi.Status) {
	src := cap_lookup(h)
	if src == nil {
		return abi.HANDLE_INVALID, abi.ERR_INVALID_HANDLE
	}
	if !cap_has_rights(src, abi.RIGHT_DUPLICATE) {
		return abi.HANDLE_INVALID, abi.ERR_ACCESS_DENIED
	}

	// Anti-amplification: new rights are strictly a subset of source rights.
	effective := src.rights & requested_rights

	// Take a temporary stabilizing ref before publication.
	if !obj_ref(src.object) {
		return abi.HANDLE_INVALID, abi.ERR_NO_MEMORY
	}

	new_handle := cap_alloc(src.object, effective)
	if new_handle == abi.HANDLE_INVALID {
		// Roll back ref count.
		obj_deref(src.object)
		return abi.HANDLE_INVALID, abi.ERR_NO_MEMORY
	}
	// Release temporary ref; published capability ref remains.
	obj_deref(src.object)
	return new_handle, abi.OK
}

// ---- cap_replace ----
//
// Self-attenuation of an existing valid capability: the slot keeps the same
// index but receives a new (bumped) generation and the rights are
// narrowed to (src.rights & new_rights).  The original handle is
// consumed; the returned handle encodes the new generation.
//
// ref_count is unchanged — the capability moves in place, not copied.
// This is equivalent to a rights-narrowing handle refresh.
// No additional management right is required beyond possession of h.
//
// Returns (new_handle, OK) on success.
// Returns (HANDLE_INVALID, ERR_INVALID_HANDLE) if h is invalid.
cap_replace :: proc "c" (h: abi.Handle, new_rights: abi.Rights) -> (abi.Handle, abi.Status) {
	entry := cap_lookup(h)
	if entry == nil {
		return abi.HANDLE_INVALID, abi.ERR_INVALID_HANDLE
	}

	idx := handle_index(h)

	// Narrow rights (anti-amplification).
	effective := entry.rights & new_rights

	// Bump generation to invalidate the old handle.
	entry.generation += 1 // u8 wraps naturally
	entry.rights = effective

	// Return a new handle at the same slot with the updated generation.
	return handle_make(idx, entry.generation), abi.OK
}

// ---- cap_get_type ----
//
// Returns the Obj_Type of the object referenced by handle h,
// or OBJ_NONE if the handle is invalid.
cap_get_type :: proc "c" (h: abi.Handle) -> abi.Obj_Type {
	entry := cap_lookup(h)
	if entry == nil {
		return abi.OBJ_NONE
	}
	return entry.object.obj_type
}
