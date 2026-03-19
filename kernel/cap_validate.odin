// ============================================================
// KOJI Kernel — Capability Validation Helpers
//
// Pure validation functions used at syscall ingress and in
// individual syscall handlers.  No mutation, no side effects.
//
// Syscall Ingress Validation (Phase 2)
// -------------------------------------
// frame_validate_ingress checks structural invariants that must
// hold for every syscall before dispatch:
//   - upper 32 bits of frame.syscall_num must be zero
//   - (per-syscall unused-arg checks are done by each handler)
//
// Per-Argument Validation
// -----------------------
// frame_validate_handle_arg verifies that the upper 32 bits of a
// u64 handle argument are zero (handles are 32-bit values).
//
// frame_validate_unused_args checks that args beyond the number
// used by a given syscall are zero — malformed frames return
// ERR_INVALID_ARGS before any state mutation occurs.
//
// Type Checking
// -------------
// cap_check_type asserts that the capability's backing object
// matches an expected type, preventing thread caps from being
// used as address-space caps and vice versa.
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Constants ----

// SYSCALL_MAX_ARGS is the number of argument registers in the ABI (arg0..arg5).
// Used by frame_validate_unused_args to know the upper bound.
SYSCALL_MAX_ARGS :: u32(6)

// ---- Ingress Validation ----

// frame_validate_ingress validates structural fields that must be correct
// for every syscall, regardless of the specific handler invoked.
//
// Returns OK if the frame is well-formed, ERR_INVALID_ARGS otherwise.
frame_validate_ingress :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	// Upper 32 bits of syscall_num must be zero.
	// (syscall numbers are 32-bit values packed into a u64 register.)
	if frame.syscall_num >> 32 != 0 {
		return abi.ERR_INVALID_ARGS
	}
	return abi.OK
}

// ---- Per-Argument Validation ----

// frame_validate_handle_arg verifies that a handle argument occupies
// only the lower 32 bits of its 64-bit register slot.
// Returns false if the upper 32 bits are non-zero.
frame_validate_handle_arg :: #force_inline proc "c" (val: u64) -> bool {
	return val >> 32 == 0
}

// frame_validate_unused_args checks that all argument registers
// beyond [0, used_count) are zero.  used_count must be <= SYSCALL_MAX_ARGS.
//
// Returns ERR_INVALID_ARGS if any unused arg is non-zero.
// Returns OK if all unused args are zero.
frame_validate_unused_args :: proc "c" (frame: ^abi.Syscall_Frame, used_count: u32) -> abi.Status {
	args := [SYSCALL_MAX_ARGS]u64{frame.arg0, frame.arg1, frame.arg2, frame.arg3, frame.arg4, frame.arg5}
	for i := used_count; i < SYSCALL_MAX_ARGS; i += 1 {
		if args[i] != 0 {
			return abi.ERR_INVALID_ARGS
		}
	}
	return abi.OK
}

// ---- Type Checking ----

// cap_check_type verifies that the capability referenced by h has the
// expected object type.
//
// Returns:
//   ERR_INVALID_HANDLE  — handle is invalid or slot is empty/stale
//   ERR_INVALID_ARGS    — handle is valid but object type does not match
//   OK                  — handle is valid and type matches
cap_check_type :: proc "c" (h: abi.Handle, expected: abi.Obj_Type) -> abi.Status {
	entry := cap_lookup(h)
	if entry == nil {
		return abi.ERR_INVALID_HANDLE
	}
	if entry.object.obj_type != expected {
		return abi.ERR_INVALID_ARGS
	}
	return abi.OK
}
