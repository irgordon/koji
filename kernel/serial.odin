// ============================================================
// KOJI Kernel — Serial Console (COM1, 0x3F8)
//
// Minimal polled-mode serial driver for early boot debug output.
// No interrupts, no buffering — mechanism only.
// ============================================================
package kernel

// x86 port I/O — requires NASM stubs (no inline asm in Odin yet)
foreign import arch_io "arch/amd64/io.o"

foreign arch_io {
	@(link_name="port_outb")
	port_outb :: proc "c" (port: u16, val: u8) ---

	@(link_name="port_inb")
	port_inb :: proc "c" (port: u16) -> u8 ---
}

COM1 :: u16(0x3F8)

serial_init :: proc "c" () {
	port_outb(COM1 + 1, 0x00)   // disable interrupts
	port_outb(COM1 + 3, 0x80)   // enable DLAB
	port_outb(COM1 + 0, 0x03)   // divisor lo: 38400 baud
	port_outb(COM1 + 1, 0x00)   // divisor hi
	port_outb(COM1 + 3, 0x03)   // 8N1
	port_outb(COM1 + 2, 0xC7)   // enable FIFO, clear, 14-byte threshold
	port_outb(COM1 + 4, 0x0B)   // IRQs enabled, RTS/DSR set
}

serial_is_transmit_empty :: #force_inline proc "c" () -> bool {
	return (port_inb(COM1 + 5) & 0x20) != 0
}

serial_putc :: proc "c" (c: u8) {
	for !serial_is_transmit_empty() {}
	port_outb(COM1, c)
}

serial_puts :: proc "c" (s: cstring) {
	p := cast([^]u8)s
	i := 0
	for p[i] != 0 {
		serial_putc(p[i])
		i += 1
	}
}

serial_put_u32 :: proc "c" (val: u32) {
	if val == 0 {
		serial_putc('0')
		return
	}
	buf: [10]u8
	i := 0
	v := val
	for v > 0 {
		buf[i] = u8(v % 10) + '0'
		v /= 10
		i += 1
	}
	for i > 0 {
		i -= 1
		serial_putc(buf[i])
	}
}
