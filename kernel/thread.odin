// ============================================================
// KOJI Kernel — Thread Object (Phase 4 Skeleton)
//
// Defines the kernel-internal thread object.  This file covers
// object representation, state model, and basic lifecycle.
// It does NOT expose a stable ABI syscall yet — that is blocked
// by CCR-001 (thread creation contract) and CCR-003 (scheduler).
//
// Thread State Model
// ------------------
//   Stopped  → Runnable  (when made eligible for scheduling)
//   Runnable → Blocked   (when waiting on IPC/port/sleep)
//   Blocked  → Runnable  (when wait condition is satisfied)
//   any live → Dying     (thread_exit or process exit)
//   Dying    → Dead      (after thread context teardown)
//
//   State transitions are enforced by the state helper functions.
//   No transition may skip states.
//
// Arch Frame Storage
// ------------------
//   Thread_Arch_Frame holds the minimal register context needed for
//   future context switching.  Fields are named to match the Syscall_Frame
//   but are distinct — this is the per-thread saved context, not the
//   syscall-entry frame.  The full context switch is deferred to the
//   scheduler phase (Phase 4 / CCR-003).
//
// Object Pool
// -----------
//   Static pool of THREAD_POOL_SIZE entries.  Allocation is O(n) for
//   v1; a free-list optimisation can be added without ABI impact.
//
// CCR References
// --------------
//   CCR-001: thread creation and lifecycle syscall contract
//   CCR-003: scheduler integration
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Thread State ----

Thread_State :: enum u32 {
	Stopped  = 0, // created but not yet scheduled
	Runnable = 1, // eligible for the run queue
	Blocked  = 2, // waiting on IPC/port/sleep
	Dying    = 3, // exit requested, teardown in progress
	Dead     = 4, // fully destroyed
}

// ---- Architecture-Neutral Register Frame ----
//
// Minimal context storage for future context switching.
// Not part of the ABI; internal to the kernel.
Thread_Arch_Frame :: struct #packed {
	rip:    u64, // instruction pointer at last context save
	rsp:    u64, // user stack pointer
	rflags: u64, // saved rflags
	rax:    u64, // general purpose (return value / syscall result)
	rbx:    u64,
	rcx:    u64,
	rdx:    u64,
	rsi:    u64,
	rdi:    u64,
	rbp:    u64,
	r8:     u64,
	r9:     u64,
	r10:    u64,
	r11:    u64,
	r12:    u64,
	r13:    u64,
	r14:    u64,
	r15:    u64,
}

// ---- Thread Object ----
//
// Obj_Header MUST be the first field (see kernel/object.odin).
Thread_Object :: struct {
	header:       Obj_Header,      // lifetime management; first field — do not reorder
	state:        Thread_State,
	arch_frame:   Thread_Arch_Frame,
	addr_space_h: abi.Handle,      // handle to the owning address space (HANDLE_INVALID = none)
	flags:        u32,             // reserved; must be zero in v1
	_pad:         [4]u8,
}

// ---- Thread Pool ----

THREAD_POOL_SIZE :: 256 // matches KOJI_MAX_THREADS_PER_PROC

g_thread_pool: [THREAD_POOL_SIZE]Thread_Object

// ---- thread_alloc ----
//
// Allocates a thread object from the static pool and initialises its
// Obj_Header with ref_count = 1 and the thread destroy hook.
//
// Returns a pointer to the allocated Thread_Object, or nil if the pool
// is exhausted.
//
// The caller must subsequently call cap_alloc to bind a capability.
thread_alloc :: proc "c" () -> ^Thread_Object {
	// A slot is available when ref_count == 0 (either fresh / never used,
	// or fully destroyed after obj_deref).  A Live object always has
	// ref_count ≥ 1, so this test is safe and sufficient in v1.
	for i in 0 ..< THREAD_POOL_SIZE {
		t := &g_thread_pool[i]
		if t.header.ref_count == 0 {
			// Zero the slot before reuse to clear any residual state.
			t^ = Thread_Object{}
			obj_init(&t.header, abi.OBJ_THREAD, thread_destroy)
			t.state        = .Stopped
			t.addr_space_h = abi.HANDLE_INVALID
			return t
		}
	}
	return nil
}

// ---- thread_destroy ----
//
// Destruction hook called by obj_deref when the last capability to a
// thread is closed.  Cleans up thread-specific resources.
//
// Note: the thread must already be in Dying state before this is called
// (enforced by obj_deref via obj_deref → Dying → destroy_fn → Dead).
@(private)
thread_destroy :: proc "c" (hdr: ^Obj_Header) {
	t := transmute(^Thread_Object)hdr
	// Close the address-space handle if one was attached.
	// (cap_close handles double-close safely via cap_lookup nil check)
	if t.addr_space_h != abi.HANDLE_INVALID {
		cap_close(t.addr_space_h)
		t.addr_space_h = abi.HANDLE_INVALID
	}
	t.state = .Dead
	// Architecture frame and other fields are zeroed on the next thread_alloc.
}

// ---- State Transition Helpers ----
//
// These enforce valid one-way transitions.
// Returns ERR_INVALID_ARGS if the transition is not permitted.

thread_set_runnable :: proc "c" (t: ^Thread_Object) -> abi.Status {
	if t.state != .Stopped && t.state != .Blocked {
		return abi.ERR_INVALID_ARGS
	}
	t.state = .Runnable
	return abi.OK
}

thread_set_blocked :: proc "c" (t: ^Thread_Object) -> abi.Status {
	if t.state != .Runnable {
		return abi.ERR_INVALID_ARGS
	}
	t.state = .Blocked
	return abi.OK
}

thread_set_dying :: proc "c" (t: ^Thread_Object) -> abi.Status {
	if t.state == .Dying || t.state == .Dead {
		return abi.ERR_INVALID_ARGS
	}
	t.state = .Dying
	return abi.OK
}

// ---- Compile-time invariants ----
// 18 × u64 fields in a #packed struct = 144 bytes exactly.
#assert(size_of(Thread_Arch_Frame) == 18 * 8)
