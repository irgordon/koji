// ============================================================
// KOJI Kernel — Address Space Object (Phase 4 Skeleton)
//
// Defines the kernel-internal address space (VMAR root) object.
// An address space is the root Virtual Memory Address Region for
// a process.  All per-process memory mappings are descendants of
// this root.
//
// This skeleton covers object representation and basic lifecycle.
// Actual page-table manipulation is deferred to Phase 4/5 (CCR-005).
//
// Attachment Model
// ----------------
//   Detached  — address space exists but no thread is using it
//   Attached  — one or more threads are running in this address space
//   Dying     — last capability closed; teardown in progress
//
//   Threads reference address spaces via handle (addr_space_h in
//   Thread_Object), not via direct pointer, to preserve capability
//   indirection.
//
// Object Pool
// -----------
//   Static pool of ADDR_SPACE_POOL_SIZE entries.
//
// CCR References
// --------------
//   CCR-005: VMAR and VMO syscall contracts
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Address Space State ----

Addr_Space_State :: enum u32 {
	Detached = 0, // created, no threads attached
	Attached = 1, // one or more threads using this address space
	Dying    = 2, // last cap closed; page-table teardown pending
	Dead     = 3, // fully destroyed
}

// ---- Address Space Object ----
//
// Obj_Header MUST be the first field (see kernel/object.odin).
Addr_Space_Object :: struct {
	header:       Obj_Header, // lifetime management; first field — do not reorder
	state:        Addr_Space_State,
	thread_count: u32,        // number of threads currently attached
	// arch-specific page table root will be added here (Phase 4/CCR-005)
	_reserved:    [8]u8,      // placeholder for arch-specific cr3 / page table root
}

addr_space_state_is_valid :: #force_inline proc "c" (s: Addr_Space_State) -> bool {
	return s >= .Detached && s <= .Dead
}

addr_space_invariants_hold :: #force_inline proc "c" (a: ^Addr_Space_Object) -> bool {
	if a == nil {
		return false
	}
	if a.header.obj_type != abi.OBJ_VMAR {
		return false
	}
	if !addr_space_state_is_valid(a.state) {
		return false
	}
	if a.state == .Attached && a.thread_count == 0 {
		return false
	}
	if a.state == .Detached && a.thread_count != 0 {
		return false
	}
	return true
}

// ---- Address Space Pool ----

ADDR_SPACE_POOL_SIZE :: 64 // per-process limit for v1

g_addr_space_pool: [ADDR_SPACE_POOL_SIZE]Addr_Space_Object

// ---- addr_space_alloc ----
//
// Allocates an address space from the static pool.
// Returns a pointer to the allocated Addr_Space_Object, or nil if
// the pool is exhausted.
//
// ref_count is set to 1 via obj_init; caller must call cap_alloc.
addr_space_alloc :: proc "c" () -> ^Addr_Space_Object {
	// A slot is available when ref_count == 0 (fresh or fully destroyed).
	for i in 0 ..< ADDR_SPACE_POOL_SIZE {
		a := &g_addr_space_pool[i]
		if a.header.ref_count == 0 {
			a^ = Addr_Space_Object{}
			obj_init(&a.header, abi.OBJ_VMAR, addr_space_destroy)
			a.state        = .Detached
			a.thread_count = 0
			if !addr_space_invariants_hold(a) {
				return nil
			}
			return a
		}
	}
	return nil
}

// ---- addr_space_destroy ----
//
// Destruction hook called by obj_deref when the last capability to
// an address space is closed.
@(private)
addr_space_destroy :: proc "c" (hdr: ^Obj_Header) {
	if hdr == nil || hdr.obj_type != abi.OBJ_VMAR {
		return
	}
	a := transmute(^Addr_Space_Object)hdr
	// Phase 4/CCR-005: page-table teardown will happen here.
	// For the skeleton, just mark it dead.
	a.thread_count = 0
	a.state = .Dead
}

// ---- Attachment Helpers ----

// addr_space_attach increments thread_count and transitions to Attached.
// Returns ERR_INVALID_ARGS if the address space is Dying or Dead.
addr_space_attach :: proc "c" (a: ^Addr_Space_Object) -> abi.Status {
	if !addr_space_invariants_hold(a) {
		return abi.ERR_INVALID_ARGS
	}
	if a.state == .Dying || a.state == .Dead {
		return abi.ERR_INVALID_ARGS
	}
	a.thread_count += 1
	a.state = .Attached
	return abi.OK
}

// addr_space_detach decrements thread_count.
// Returns ERR_INVALID_ARGS if the address space is not Attached.
addr_space_detach :: proc "c" (a: ^Addr_Space_Object) -> abi.Status {
	if !addr_space_invariants_hold(a) {
		return abi.ERR_INVALID_ARGS
	}
	if a.state != .Attached || a.thread_count == 0 {
		return abi.ERR_INVALID_ARGS
	}
	a.thread_count -= 1
	if a.thread_count == 0 {
		a.state = .Detached
	}
	return abi.OK
}

#assert(offset_of(Addr_Space_Object, header) == 0)
