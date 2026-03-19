; ============================================================
; KOJI Kernel — x86 Port I/O Stubs (NASM)
;
; Thin wrappers for IN/OUT instructions.
; Called from Odin via foreign import.
; ============================================================

bits 64
default rel

section .text

; void port_outb(u16 port, u8 val)
;   rdi = port, sil = val
global port_outb
port_outb:
    mov dx, di
    mov al, sil
    out dx, al
    ret

; u8 port_inb(u16 port)
;   rdi = port → al = result
global port_inb
port_inb:
    mov dx, di
    in  al, dx
    movzx eax, al
    ret

; void port_outw(u16 port, u16 val)
global port_outw
port_outw:
    mov dx, di
    mov ax, si
    out dx, ax
    ret

; u16 port_inw(u16 port)
global port_inw
port_inw:
    mov dx, di
    in  ax, dx
    movzx eax, ax
    ret

; void port_outl(u16 port, u32 val)
global port_outl
port_outl:
    mov dx, di
    mov eax, esi
    out dx, eax
    ret

; u32 port_inl(u16 port)
global port_inl
port_inl:
    mov dx, di
    in  eax, dx
    ret
