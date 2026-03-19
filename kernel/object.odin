// ============================================================
// KOJI Kernel — Internal Object Header and Lifetime Model
//
// All kernel objects must embed Obj_Header as their FIRST field.
// This guarantees that a ^Obj_Header and a ^ConcreteObject share
// the same address, enabling safe pointer-level downcasts.
//
// Lifetime Model:
//   ref_count starts at 1 on object creation (set by the object's
//   own init function; cap_alloc does NOT set it).
//   cap_alloc increments ref_count before linking a new slot.
//   cap_close calls obj_deref; when ref_count reaches zero the
//   object transitions Dying → Dead and destroy_fn is invoked.
//
// State Transitions (one-way only):
//   Live → Dying → Dead
//   No backward transitions are permitted.
//
// Invariants:
//   - Only Live objects may be referenced by capabilities.
//   - ref_count must never underflow (obj_deref guards against this).
//   - destroy_fn must not close capabilities that point back to the
//     same object (would re-enter obj_deref with ref_count already 0).
//   - All object pools are statically allocated (no heap in v1 kernel).
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Object Lifetime States ----
//
// Transitions are strictly one-way: Live → Dying → Dead.
Obj_State :: enum u32 {
	Live  = 0, // object is alive; capabilities may reference it
	Dying = 1, // last capability closed; destroy_fn executing
	Dead  = 2, // object fully destroyed; memory may be reclaimed
}

// ---- Destruction Hook ----
//
// Called by obj_deref when ref_count reaches zero.
// Receives a pointer to the object's embedded Obj_Header.
// The concrete type is recovered via transmute (header is at offset 0).
Obj_Destroy_Fn :: #type proc "c" (hdr: ^Obj_Header)

// ---- Object Header ----
//
// Must be the FIRST field of every kernel object struct.
// Never exposed to user mode.
Obj_Header :: struct {
	obj_type:   abi.Obj_Type,   // stable type discriminant (matches ABI)
	state:      Obj_State,      // current lifetime state
	ref_count:  u32,            // number of live capabilities pointing here
	destroy_fn: Obj_Destroy_Fn, // called when ref_count → 0; nil is allowed
}

// ---- obj_ref ----
//
// Increments the reference count.
// Caller must guarantee the object is currently Live.
obj_ref :: #force_inline proc "c" (hdr: ^Obj_Header) {
	hdr.ref_count += 1
}

// ---- obj_deref ----
//
// Decrements the reference count.
// If ref_count reaches zero:
//   1. state → Dying
//   2. destroy_fn is called (if non-nil)
//   3. state → Dead
// Returns true if the object was destroyed, false otherwise.
// Returns false (no-op) if ref_count is already zero — defensive guard.
obj_deref :: proc "c" (hdr: ^Obj_Header) -> bool {
	// Defensive guard: double-free or accounting error.
	if hdr.ref_count == 0 {
		return false
	}
	hdr.ref_count -= 1
	if hdr.ref_count != 0 {
		return false
	}
	// Last reference removed — destroy the object.
	hdr.state = .Dying
	if hdr.destroy_fn != nil {
		hdr.destroy_fn(hdr)
	}
	hdr.state = .Dead
	return true
}

// ---- obj_is_live ----
//
// Returns true iff the object is in the Live state.
// Used by cap_alloc and cap_lookup to reject dying/dead objects.
obj_is_live :: #force_inline proc "c" (hdr: ^Obj_Header) -> bool {
	return hdr.state == .Live
}

// ---- obj_init ----
//
// Initializes an object header in the Live state with ref_count = 1.
// Every object's own init function must call this before calling cap_alloc.
obj_init :: #force_inline proc "c" (hdr: ^Obj_Header, t: abi.Obj_Type, destroy: Obj_Destroy_Fn) {
	hdr.obj_type   = t
	hdr.state      = .Live
	hdr.ref_count  = 1
	hdr.destroy_fn = destroy
}

// ---- Compile-time invariant ----
// Obj_Header must be at least 20 bytes (3×u32 + pointer) and pointer-aligned.
#assert(size_of(Obj_Header) >= 20)
#assert(align_of(Obj_Header) >= 4)
