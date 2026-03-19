// ============================================================
// KOJI Kernel — Capability Syscall Handlers
//
// Handlers for the three handle-management syscalls:
//   SYS_HANDLE_CLOSE     (0) — close and invalidate a handle
//   SYS_HANDLE_DUPLICATE (1) — copy with equal-or-fewer rights
//   SYS_HANDLE_REPLACE   (2) — rights-narrowing in-place refresh
//
// Calling Convention (x86_64 SYSCALL; see KOJI_ABI_V1.h §8)
// -----------------------------------------------------------
//   rdi → frame.arg0    rsi → frame.arg1    rdx → frame.arg2 (out)
//
// Handler Contract
// ----------------
//   • Validate all inputs BEFORE mutating any state.
//   • Return only ABI-defined koji_status_t values.
//   • Write secondary outputs into frame fields ONLY on OK.
//   • Never access frame.arg* that are reserved for this syscall.
//
// Error Codes Used
// ----------------
//   ERR_INVALID_HANDLE   — handle is the sentinel, index out of range,
//                          slot empty, generation mismatch, or object dead
//   ERR_ACCESS_DENIED    — caller lacks required right (DUPLICATE)
//   ERR_INVALID_ARGS     — reserved arg field is non-zero
//   ERR_NO_MEMORY        — cap table is full
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- SYS_HANDLE_CLOSE ----
//
// arg0 = handle to close
// args 1-5 must be zero
//
// Returns OK on success.
// Returns ERR_INVALID_HANDLE if the handle is invalid.
// Returns ERR_INVALID_ARGS if any reserved arg is non-zero.
sys_handle_close :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	// Reserved-arg guard: only arg0 is used.
	if frame_validate_unused_args(frame, 1) != abi.OK {
		return abi.ERR_INVALID_ARGS
	}
	if !frame_validate_handle_arg(frame.arg0) {
		return abi.ERR_INVALID_HANDLE
	}
	h := abi.Handle(u32(frame.arg0))
	return cap_close(h)
}

// ---- SYS_HANDLE_DUPLICATE ----
//
// arg0 = source handle
// arg1 = new rights mask (u32 in lower 32 bits)
// args 2-5 must be zero
//
// On success:
//   rax (return) = OK
//   rdx (frame.arg2) = new handle
sys_handle_duplicate :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	// Reserved-arg guard: arg0 and arg1 are used.
	if frame_validate_unused_args(frame, 2) != abi.OK {
		return abi.ERR_INVALID_ARGS
	}
	if !frame_validate_handle_arg(frame.arg0) {
		return abi.ERR_INVALID_HANDLE
	}

	src_handle    := abi.Handle(u32(frame.arg0))
	new_rights    := abi.Rights(u32(frame.arg1))

	new_handle, status := cap_duplicate(src_handle, new_rights)
	if status == abi.OK {
		// Secondary return value in rdx (arg2 on exit path).
		frame.arg2 = u64(u32(new_handle))
	}
	return status
}

// ---- SYS_HANDLE_REPLACE ----
//
// arg0 = handle to replace (consumed on success)
// arg1 = new rights mask (u32 in lower 32 bits; must be ⊆ current rights)
// args 2-5 must be zero
//
// On success:
//   rax (return) = OK
//   rdx (frame.arg2) = new handle (same slot, bumped generation)
//
// The original handle is invalidated.  ref_count is unchanged.
// New rights are computed as: old_rights & arg1 (anti-amplification).
sys_handle_replace :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	// Reserved-arg guard: arg0 and arg1 are used.
	if frame_validate_unused_args(frame, 2) != abi.OK {
		return abi.ERR_INVALID_ARGS
	}
	if !frame_validate_handle_arg(frame.arg0) {
		return abi.ERR_INVALID_HANDLE
	}

	h          := abi.Handle(u32(frame.arg0))
	new_rights := abi.Rights(u32(frame.arg1))

	new_handle, status := cap_replace(h, new_rights)
	if status == abi.OK {
		frame.arg2 = u64(u32(new_handle))
	}
	return status
}
