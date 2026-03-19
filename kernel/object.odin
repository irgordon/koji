// ============================================================
// KOJI Kernel — Internal Object Header and Lifetime Model
//
// All kernel objects must embed Obj_Header as their FIRST field.
// This guarantees that a ^Obj_Header and a ^ConcreteObject share
// the same address for controlled casts; type and lifetime validity
// still require separate checks.
//
// Lifetime Model:
//   obj_init sets ref_count = 0.  The object is Live but not yet published.
//   Every successful cap_alloc publishes one capability-owned reference.
//   Kernel code may also take temporary references with obj_ref and must
//   release them with obj_deref.
//   cap_close drops one published reference with obj_deref.  Final
//   destruction occurs only when the total ref_count reaches zero.
//
// State Transitions (one-way only):
//   Live → Dying → Dead
//   No backward transitions are permitted.
//
// Invariants:
//   - Only Live objects may be published through capabilities.
//   - ref_count is the total of published capability refs + temporary refs.
//   - ref_count must never underflow (obj_deref guards against this).
//   - destroy_fn executes while state == Dying and must not resurrect,
//     re-own, or re-publish the same object.
//   - All object pools are statically allocated (no heap in v1 kernel).
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Object Lifetime States ----
//
// Transitions are strictly one-way: Live → Dying → Dead.
Obj_State :: enum u32 {
	Live  = 0, // object is alive; capabilities may reference it
	Dying = 1, // ref_count reached zero; destroy_fn executing
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
	ref_count:  u32,            // published capability refs + temporary kernel refs
	destroy_fn: Obj_Destroy_Fn, // called when ref_count → 0; nil is allowed
}

// ---- obj_ref ----
//
// Increments the reference count.
// Returns true on success.
// Returns false unless hdr is non-nil, state == Live, and ref_count < max(u32).
obj_ref :: #force_inline proc "c" (hdr: ^Obj_Header) -> bool {
	if hdr == nil {
		return false
	}
	if hdr.state != .Live {
		return false
	}
	if hdr.ref_count == max(u32) {
		return false
	}
	hdr.ref_count += 1
	return true
}

// ---- obj_deref ----
//
// Decrements the reference count.
// If ref_count reaches zero:
//   1. state → Dying
//   2. destroy_fn is called (if non-nil)
//   3. state → Dead
// Returns true if the object was destroyed, false otherwise.
// Returns false unless hdr is non-nil, state == Live, and ref_count > 0.
obj_deref :: proc "c" (hdr: ^Obj_Header) -> bool {
	if hdr == nil {
		return false
	}
	if hdr.state != .Live {
		return false
	}
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
// Returns true iff hdr is non-nil and the object is in the Live state.
// Used by cap_alloc and cap_lookup to reject dying/dead objects.
obj_is_live :: #force_inline proc "c" (hdr: ^Obj_Header) -> bool {
	if hdr == nil {
		return false
	}
	return hdr.state == .Live
}

// ---- obj_init ----
//
// Initializes an object header in the Live state with ref_count = 0.
// Newly initialized objects are not yet published until cap_alloc succeeds.
obj_init :: #force_inline proc "c" (hdr: ^Obj_Header, t: abi.Obj_Type, destroy: Obj_Destroy_Fn) {
	if hdr == nil {
		return
	}
	hdr.obj_type   = t
	hdr.state      = .Live
	hdr.ref_count  = 0
	hdr.destroy_fn = destroy
}

// ---- Compile-time invariant ----
// Obj_Header must fit 3×u32 plus a pointer and keep pointer alignment.
// On amd64 this is expected to be 24 bytes; keep checks architecture-safe.
#assert(size_of(Obj_Header) >= 12 + size_of(rawptr))
#assert(align_of(Obj_Header) >= align_of(rawptr))
