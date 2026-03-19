// ============================================================
// KOJI Kernel — Syscall Dispatch
//
// koji_syscall_dispatch is called from syscall_entry.asm with
// a pointer to the register-save Syscall_Frame.
//
// This is pure mechanism:
//   - validate syscall number
//   - dispatch to handler
//   - return status in rax
//
// No policy. No implicit behavior. No retry logic.
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Syscall Handler Signature ----

Syscall_Handler :: #type proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status

// ---- Dispatch Table ----
// Indexed by syscall number. nil = not implemented yet.

syscall_table: [abi.SYSCALL_COUNT]Syscall_Handler

syscall_table_init :: proc "c" () {
	// Wire up implemented syscalls.
	// Unimplemented slots remain nil → ERR_INVALID_SYSCALL.
	syscall_table[u32(abi.SYS_HANDLE_CLOSE)]     = sys_handle_close
	syscall_table[u32(abi.SYS_HANDLE_DUPLICATE)]  = sys_handle_duplicate
	syscall_table[u32(abi.SYS_ABI_INFO)]          = sys_abi_info
	// All others are nil — stubs added as subsystems come online.
}

// ---- Entry Point (called from NASM) ----

@(export, link_name="koji_syscall_dispatch")
koji_syscall_dispatch :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	num := u32(frame.syscall_num)

	if num >= abi.SYSCALL_COUNT {
		return abi.ERR_INVALID_SYSCALL
	}

	handler := syscall_table[num]
	if handler == nil {
		return abi.ERR_INVALID_SYSCALL
	}

	return handler(frame)
}

// ============================================================
// Implemented Syscall Handlers
// ============================================================

// ---- SYS_HANDLE_CLOSE ----
// arg0 (rdi) = handle to close

sys_handle_close :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	h := abi.Handle(u32(frame.arg0))
	return cap_close(h)
}

// ---- SYS_HANDLE_DUPLICATE ----
// arg0 (rdi) = source handle
// arg1 (rsi) = new rights mask
// Returns: new handle in rdx (frame.arg2 on return path)

sys_handle_duplicate :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	src_handle  := abi.Handle(u32(frame.arg0))
	new_rights  := abi.Rights(u32(frame.arg1))

	new_handle, status := cap_duplicate(src_handle, new_rights)
	if status == abi.OK {
		// Secondary return value goes in arg2 slot (mapped to rdx on exit)
		frame.arg2 = u64(u32(new_handle))
	}
	return status
}

// ---- SYS_ABI_INFO ----
// No arguments. Returns ABI version info.
// arg2 (rdx) = packed version: (major << 16) | (minor << 8) | patch

sys_abi_info :: proc "c" (frame: ^abi.Syscall_Frame) -> abi.Status {
	version := u64(abi.ABI_VERSION_MAJOR) << 16 |
	           u64(abi.ABI_VERSION_MINOR) << 8  |
	           u64(abi.ABI_VERSION_PATCH)
	frame.arg2 = version
	return abi.OK
}
