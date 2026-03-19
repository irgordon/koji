// ============================================================
// KOJI Kernel — Syscall Dispatch (Phase 2: Ingress Normalization)
//
// koji_syscall_dispatch is the single, deterministic syscall ingress.
// It is called from syscall_entry.asm with a pointer to the
// register-save Syscall_Frame.
//
// Ingress Steps (must happen in this order, no exceptions)
// --------------------------------------------------------
//   1. frame nil check
//   2. frame_validate_ingress — reject malformed frame structure
//      (upper 32 bits of syscall_num must be zero)
//   3. Bounds check syscall_num against SYSCALL_COUNT
//   4. Nil-handler check (number in range but not yet implemented)
//   5. Dispatch to registered handler
//
// No state is mutated before step 5.
// Every exit path returns a defined ABI status code.
//
// Error Code Map
// --------------
//   ERR_INVALID_ARGS    — nil frame or malformed frame (reserved bits set)
//   ERR_INVALID_SYSCALL — syscall number out of range or in-range nil slot
//
// Return Convention
// -----------------
//   Status is returned as the function return value (rax on syscall exit).
//   Optional payloads are written to frame output fields (for example arg2/rdx)
//   only on successful syscall handling.
//
// Handlers live in their respective subsystem files:
//   cap_syscalls.odin   — SYS_HANDLE_CLOSE/DUPLICATE/REPLACE
//   syscall.odin        — SYS_ABI_INFO (and future misc syscalls)
//   (others TBD per phase schedule)
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Syscall Handler Signature ----

Syscall_Handler :: #type proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status

// ---- Dispatch Table ----
// Indexed by syscall number.  nil = not yet implemented.

syscall_table: [abi.SYSCALL_COUNT]Syscall_Handler

syscall_table_init :: proc "c" () {
	// ---- Phase 1/5: Handle-family syscalls ----
	syscall_table[u32(abi.SYS_HANDLE_CLOSE)]     = sys_handle_close
	syscall_table[u32(abi.SYS_HANDLE_DUPLICATE)]  = sys_handle_duplicate
	syscall_table[u32(abi.SYS_HANDLE_REPLACE)]    = sys_handle_replace

	// ---- Phase 5: Miscellaneous / introspection ----
	syscall_table[u32(abi.SYS_ABI_INFO)]          = sys_abi_info

	// All other slots remain nil.
	// Numbers 3-9   (Process/Thread lifecycle) — blocked by CCR-001, CCR-003
	// Numbers 10-13 (IPC channels)             — blocked by CCR-002
	// Numbers 14-16 (Ports)                    — blocked by CCR-002
	// Numbers 17-22 (VMO/VMAR)                 — blocked by CCR-005
	// Numbers 23-24 (IRQ)                      — blocked by CCR-003
	// nil handler → ERR_INVALID_SYSCALL (not yet implemented)
}

// ---- Entry Point (called from NASM) ----
//
// Phase 2: single ingress path with full frame validation before dispatch.
// No handler may be called with an invalid or malformed frame.

@(export, link_name="koji_syscall_dispatch")
koji_syscall_dispatch :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	// Step 1: nil frame pointer rejection.
	if frame == nil {
		return abi.ERR_INVALID_ARGS
	}

	// Step 2: structural frame validation (reserved bits, etc.)
	// Must happen before inspecting syscall_num.
	if frame_validate_ingress(frame) != abi.OK {
		return abi.ERR_INVALID_ARGS
	}

	// Step 3: syscall number bounds check.
	num := u32(frame.syscall_num)
	if num >= abi.SYSCALL_COUNT {
		return abi.ERR_INVALID_SYSCALL
	}

	// Step 4: nil handler → not yet implemented.
	handler := syscall_table[num]
	if handler == nil {
		return abi.ERR_INVALID_SYSCALL
	}

	// Step 5: dispatch.  No state was mutated above this line.
	return handler(frame)
}

// ============================================================
// Miscellaneous Syscall Handlers
// (Capability handlers are in cap_syscalls.odin)
// ============================================================

// ---- SYS_ABI_INFO ----
//
// No input arguments; all args must be zero.
//
// On success:
//   rax (return) = OK
//   rdx (frame.arg2) = packed version: (major << 16) | (minor << 8) | patch
sys_abi_info :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	// All six args must be zero for this no-input syscall.
	if frame_validate_unused_args(frame, 0) != abi.OK {
		return abi.ERR_INVALID_ARGS
	}
	version := u64(abi.ABI_VERSION_MAJOR) << 16 |
	           u64(abi.ABI_VERSION_MINOR) << 8  |
	           u64(abi.ABI_VERSION_PATCH)
	frame.arg2 = version
	return abi.OK
}
