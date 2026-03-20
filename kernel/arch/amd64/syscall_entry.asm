; ============================================================
; KOJI Kernel — AMD64 SYSCALL Entry / Exit (NASM)
;
; On SYSCALL:
;   rcx ← user RIP  (clobbered by hardware)
;   r11 ← user RFLAGS
;   rax  = syscall number
;   rdi, rsi, rdx, r10, r8, r9 = args 0–5
;
; We swap to the kernel stack, build a Syscall_Frame, and
; call into Odin's koji_syscall_dispatch(frame: ^Syscall_Frame).
; ============================================================

bits 64
default rel

section .data
align 8

; Per-CPU kernel stack pointer (single-CPU v1).
; Set during boot before enabling SYSCALL.
global kernel_syscall_rsp
kernel_syscall_rsp: dq 0

section .text

global syscall_entry_point
extern koji_syscall_dispatch

; ----- MSR addresses for SYSCALL setup -----
%define MSR_STAR    0xC0000081
%define MSR_LSTAR   0xC0000082
%define MSR_SFMASK  0xC0000084

; ----- Called once from Odin to wire up SYSCALL MSRs -----
global syscall_init
syscall_init:
    ; STAR: kernel CS/SS in bits [47:32], user CS/SS in bits [63:48]
    ; Kernel: CS=0x08 SS=0x10  |  User: CS=0x1B SS=0x23
    ; SYSRET loads CS from STAR[63:48]+16, SS from STAR[63:48]+8
    mov ecx, MSR_STAR
    xor eax, eax
    mov edx, 0x00180008     ; user base=0x18, kernel base=0x08
    wrmsr

    ; LSTAR: RIP on SYSCALL
    mov ecx, MSR_LSTAR
    lea rax, [syscall_entry_point]
    mov rdx, rax
    shr rdx, 32
    wrmsr

    ; SFMASK: clear IF (bit 9) on entry → interrupts off in kernel
    mov ecx, MSR_SFMASK
    mov eax, (1 << 9)
    xor edx, edx
    wrmsr

    ret

; ----- SYSCALL lands here -----
syscall_entry_point:
    ; Save user RSP, load kernel RSP
    ; swapgs would go here with per-CPU data; v1 uses a global
    mov [rel user_rsp_scratch], rsp
    mov rsp, [rel kernel_syscall_rsp]

    ; Build Syscall_Frame on kernel stack (matches abi_generated.odin)
    ; struct order: syscall_num, arg0..arg5, user_rip, user_rflags, user_rsp
    push qword [rel user_rsp_scratch]   ; user_rsp
    push r11                            ; user_rflags
    push rcx                            ; user_rip
    push r9                             ; arg5
    push r8                             ; arg4
    push r10                            ; arg3
    push rdx                            ; arg2
    push rsi                            ; arg1
    push rdi                            ; arg0
    push rax                            ; syscall_num

    ; Pass pointer to frame as arg0
    mov rdi, rsp
    call koji_syscall_dispatch
    ; rax now holds koji_status_t return value

    ; Restore frame → registers
    pop rdi                             ; discard saved syscall_num (rax = return)
    pop rdi                             ; arg0 (restore for caller convention)
    pop rsi                             ; arg1
    pop rdx                             ; arg2 / secondary return
    add rsp, 8                          ; skip arg3 (r10)
    add rsp, 8                          ; skip arg4 (r8)
    add rsp, 8                          ; skip arg5 (r9)
    pop rcx                             ; user_rip
    pop r11                             ; user_rflags
    pop rsp                             ; user_rsp

    o64 sysret

section .bss
align 8
user_rsp_scratch: resq 1

section .note.GNU-stack noalloc noexec nowrite progbits
