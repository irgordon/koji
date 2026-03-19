; ============================================================
; KOJI Kernel — AMD64 Bootstrap (NASM)
; Entry point from bootloader (Limine or similar multiboot2).
; Sets up initial stack, enters long mode (assumed from bootloader),
; and jumps to Odin kernel_main.
; ============================================================

bits 64
default rel

section .bss
align 16
stack_bottom:
    resb 16384              ; 16 KiB kernel stack
stack_top:

section .data
align 8
gdt64:
    dq 0                                        ; null descriptor
.code: equ $ - gdt64
    dq (1 << 43) | (1 << 44) | (1 << 47) | (1 << 53)  ; code segment
.data: equ $ - gdt64
    dq (1 << 44) | (1 << 47) | (1 << 41)               ; data segment
.tss: equ $ - gdt64
    dq 0                                        ; TSS lo (filled at runtime)
    dq 0                                        ; TSS hi
.pointer:
    dw $ - gdt64 - 1        ; limit
    dq gdt64                ; base

section .text
global _start
extern kernel_main

_start:
    ; ----- load our own GDT -----
    lgdt [gdt64.pointer]

    ; reload segment registers
    mov ax, gdt64.data
    mov ds, ax
    mov es, ax
    mov fs, ax
    mov gs, ax
    mov ss, ax

    ; ----- set up kernel stack -----
    lea rsp, [stack_top]

    ; ----- clear RFLAGS -----
    push 0
    popfq

    ; ----- call into Odin -----
    xor edi, edi            ; arg0 = 0 (reserved for boot info pointer)
    call kernel_main

    ; ----- halt if kernel_main returns -----
.halt:
    cli
    hlt
    jmp .halt
