// ============================================================
// KOJI Kernel — Entry Point
//
// kernel_main is called from boot.asm after stack and GDT setup.
//
// Boot Sequence (must remain in this order)
// -----------------------------------------
//   1. Serial console        — debug output becomes available
//   2. Capability table      — CNode slots zeroed, generations reset
//   3. Syscall dispatch      — handler table wired (Phase 1/2/5)
//   4. SYSCALL MSRs          — hardware entry point live
//   5. Halt                  — no scheduler yet; loops until Phase 4
//
// Subsystems initialized here must be idempotent with respect to
// a warm-restart (each init procedure re-zeros its state).
// ============================================================
package kernel

import abi "../abi/generated/odin"

// Foreign imports: NASM-assembled arch stubs
foreign import arch_boot    "arch/amd64/boot.o"
foreign import arch_syscall "arch/amd64/syscall_entry.o"

@(export, link_name="kernel_main")
kernel_main :: proc "c" (boot_info: rawptr) {
	// 1. Serial console
	serial_init()
	serial_puts("KOJI kernel v")
	serial_put_u32(abi.ABI_VERSION_MAJOR)
	serial_putc('.')
	serial_put_u32(abi.ABI_VERSION_MINOR)
	serial_putc('.')
	serial_put_u32(abi.ABI_VERSION_PATCH)
	serial_puts(" booting\n")

	// 2. Capability table (CNode)
	serial_puts("[cap]  initializing capability table\n")
	cap_table_init()
	serial_puts("[cap]  capability table ready\n")

	// 3. Syscall dispatch table
	serial_puts("[sys]  initializing syscall dispatch table\n")
	syscall_table_init()
	serial_puts("[sys]  dispatch table ready\n")

	// 4. SYSCALL MSR wiring
	serial_puts("[sys]  configuring SYSCALL MSRs\n")
	syscall_arch_init()
	serial_puts("[sys]  SYSCALL entry wired\n")

	serial_puts("[boot] kernel initialization complete — halting\n")

	// 5. Halt — no scheduler, no user-mode entry yet.
	halt_loop()
}

// Provided by arch NASM stubs
foreign arch_syscall {
	@(link_name="syscall_init")
	syscall_arch_init :: proc "c" () ---

	@(link_name="kernel_syscall_rsp")
	kernel_syscall_rsp: u64
}

halt_loop :: proc "c" () {
	for {
		// cli; hlt — requires inline asm, use a tight loop for now.
		// In production, boot.asm .halt handles this.
	}
}
