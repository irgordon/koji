/*
 * KOJI Syscall ABI — Canonical Definition
 * Version: 1.0.0
 * Status:  Draft
 *
 * This file is the SINGLE SOURCE OF TRUTH for the KOJI ABI.
 * All language bindings (Odin, Go) are mechanically generated from this header.
 *
 * Rules:
 *   - No generated artifact may become the source of truth.
 *   - Field ordering, width, signedness, and constant values are binding.
 *   - Reserved fields must not be removed or reinterpreted.
 *   - All changes require a version bump per VERSIONING_POLICY.md.
 *
 * Syscall Calling Convention (x86_64, SYSCALL instruction):
 *   rax  = syscall number
 *   rdi  = arg0  (typically: handle)
 *   rsi  = arg1
 *   rdx  = arg2
 *   r10  = arg3  (rcx clobbered by SYSCALL)
 *   r8   = arg4
 *   r9   = arg5
 *
 *   Return:
 *   rax  = koji_status_t
 *   rdx  = out0  (syscall-specific secondary return)
 */

#ifndef KOJI_ABI_V1_H
#define KOJI_ABI_V1_H

/* ======================================================================
 * 1. ABI Identity
 * ====================================================================== */

#define KOJI_ABI_MAGIC       0x4B4F4A49   /* "KOJI" in ASCII */
#define KOJI_ABI_VERSION_MAJOR  1
#define KOJI_ABI_VERSION_MINOR  0
#define KOJI_ABI_VERSION_PATCH  0

/* ======================================================================
 * 2. Fixed-Width Type Conventions
 * ====================================================================== */

typedef unsigned char       koji_u8;
typedef unsigned short      koji_u16;
typedef unsigned int        koji_u32;
typedef unsigned long long  koji_u64;
typedef signed int          koji_i32;
typedef signed long long    koji_i64;

/* ======================================================================
 * 3. Handle Encoding
 *
 * A handle is a 32-bit opaque token.
 *
 *   Bits [31:24]  — generation counter (use-after-free detection)
 *   Bits [23:0]   — index into capability table
 *
 * KOJI_HANDLE_INVALID is the canonical "no handle" sentinel.
 * ====================================================================== */

typedef koji_u32 koji_handle_t;

#define KOJI_HANDLE_INVALID       ((koji_handle_t)0xFFFFFFFF)

#define KOJI_HANDLE_GEN_SHIFT     24
#define KOJI_HANDLE_GEN_MASK      ((koji_u32)0xFF000000)
#define KOJI_HANDLE_INDEX_MASK    ((koji_u32)0x00FFFFFF)

#define KOJI_HANDLE_MAX_INDEX     0x00FFFFFF   /* 16,777,215 entries */
#define KOJI_HANDLE_MAX_GEN       0xFF         /* 255 generations   */

/* ======================================================================
 * 4. Object Taxonomy
 *
 * Every kernel object has a type discriminant.
 * These values are stable and must not be reordered.
 * ====================================================================== */

typedef koji_u32 koji_obj_type_t;

#define KOJI_OBJ_NONE             ((koji_obj_type_t)0)
#define KOJI_OBJ_PROCESS          ((koji_obj_type_t)1)
#define KOJI_OBJ_THREAD           ((koji_obj_type_t)2)
#define KOJI_OBJ_VMAR             ((koji_obj_type_t)3)  /* Virtual Memory Address Region */
#define KOJI_OBJ_VMO              ((koji_obj_type_t)4)  /* Virtual Memory Object         */
#define KOJI_OBJ_CHANNEL          ((koji_obj_type_t)5)  /* IPC channel endpoint          */
#define KOJI_OBJ_PORT             ((koji_obj_type_t)6)  /* async notification port       */
#define KOJI_OBJ_IRQ              ((koji_obj_type_t)7)  /* IRQ binding object             */

#define KOJI_OBJ_TYPE_COUNT       8

/* ======================================================================
 * 5. Rights Bitmask Model
 *
 * Rights are a u32 bitmask attached to each capability.
 * A capability grants access to an object with specific rights.
 * Rights are monotonically non-increasing: you can drop rights,
 * never add them.
 * ====================================================================== */

typedef koji_u32 koji_rights_t;

#define KOJI_RIGHT_NONE           ((koji_rights_t)0)
#define KOJI_RIGHT_READ           ((koji_rights_t)(1 << 0))
#define KOJI_RIGHT_WRITE          ((koji_rights_t)(1 << 1))
#define KOJI_RIGHT_EXECUTE        ((koji_rights_t)(1 << 2))
#define KOJI_RIGHT_DUPLICATE      ((koji_rights_t)(1 << 3))
#define KOJI_RIGHT_TRANSFER       ((koji_rights_t)(1 << 4))
#define KOJI_RIGHT_MAP            ((koji_rights_t)(1 << 5))
#define KOJI_RIGHT_SIGNAL         ((koji_rights_t)(1 << 6))
#define KOJI_RIGHT_MANAGE         ((koji_rights_t)(1 << 7))

#define KOJI_RIGHTS_ALL           ((koji_rights_t)0xFF)

/* ======================================================================
 * 6. Error / Status Codes
 *
 * Every syscall returns a koji_status_t in rax.
 * Zero is success. All errors are positive.
 * ====================================================================== */

typedef koji_i32 koji_status_t;

#define KOJI_OK                   ((koji_status_t)0)
#define KOJI_ERR_INVALID_HANDLE   ((koji_status_t)1)
#define KOJI_ERR_INVALID_SYSCALL  ((koji_status_t)2)
#define KOJI_ERR_ACCESS_DENIED    ((koji_status_t)3)
#define KOJI_ERR_NO_MEMORY        ((koji_status_t)4)
#define KOJI_ERR_INVALID_ARGS     ((koji_status_t)5)
#define KOJI_ERR_NOT_FOUND        ((koji_status_t)6)
#define KOJI_ERR_ALREADY_EXISTS   ((koji_status_t)7)
#define KOJI_ERR_BUFFER_TOO_SMALL ((koji_status_t)8)
#define KOJI_ERR_WOULD_BLOCK      ((koji_status_t)9)
#define KOJI_ERR_TIMED_OUT        ((koji_status_t)10)
#define KOJI_ERR_CANCELLED        ((koji_status_t)11)
#define KOJI_ERR_PEER_CLOSED      ((koji_status_t)12)

#define KOJI_ERR_COUNT            13

/* ======================================================================
 * 7. Syscall Numbers
 *
 * Stable numbering. Gaps may exist; new syscalls append.
 * ====================================================================== */

typedef koji_u32 koji_syscall_t;

/* -- Lifecycle -------------------------------------------------------- */
#define KOJI_SYS_NOOP             ((koji_syscall_t)0)
#define KOJI_SYS_PROCESS_EXIT     ((koji_syscall_t)1)
#define KOJI_SYS_THREAD_EXIT      ((koji_syscall_t)2)

/* -- IPC -------------------------------------------------------------- */
#define KOJI_SYS_CHANNEL_CREATE   ((koji_syscall_t)3)
#define KOJI_SYS_CHANNEL_SEND     ((koji_syscall_t)4)
#define KOJI_SYS_CHANNEL_RECV     ((koji_syscall_t)5)
#define KOJI_SYS_CHANNEL_CALL     ((koji_syscall_t)6)   /* send + recv */
#define KOJI_SYS_PORT_CREATE      ((koji_syscall_t)7)
#define KOJI_SYS_PORT_SIGNAL      ((koji_syscall_t)8)
#define KOJI_SYS_PORT_WAIT        ((koji_syscall_t)9)

/* -- Handles ---------------------------------------------------------- */
#define KOJI_SYS_HANDLE_CLOSE     ((koji_syscall_t)10)
#define KOJI_SYS_HANDLE_DUP       ((koji_syscall_t)11)
#define KOJI_SYS_HANDLE_REPLACE   ((koji_syscall_t)12)

/* -- Memory ----------------------------------------------------------- */
#define KOJI_SYS_VMO_CREATE       ((koji_syscall_t)13)
#define KOJI_SYS_VMAR_MAP         ((koji_syscall_t)14)
#define KOJI_SYS_VMAR_UNMAP       ((koji_syscall_t)15)

/* -- Threads / Processes ---------------------------------------------- */
#define KOJI_SYS_PROCESS_CREATE   ((koji_syscall_t)16)
#define KOJI_SYS_THREAD_CREATE    ((koji_syscall_t)17)

/* -- IRQ -------------------------------------------------------------- */
#define KOJI_SYS_IRQ_BIND         ((koji_syscall_t)18)
#define KOJI_SYS_IRQ_ACK          ((koji_syscall_t)19)

#define KOJI_SYSCALL_COUNT        20

/* ======================================================================
 * 8. Syscall Frame Layout
 *
 * When transitioning Ring 3 → Ring 0 via SYSCALL, the kernel saves
 * user context into this frame on the kernel stack.
 * ====================================================================== */

typedef struct koji_syscall_frame {
    /* saved by SYSCALL instruction */
    koji_u64 user_rip;          /* RCX at SYSCALL entry */
    koji_u64 user_rflags;       /* R11 at SYSCALL entry */

    /* saved by kernel entry stub */
    koji_u64 user_rsp;
    koji_u64 rax;               /* syscall number       */
    koji_u64 rdi;               /* arg0                 */
    koji_u64 rsi;               /* arg1                 */
    koji_u64 rdx;               /* arg2                 */
    koji_u64 r10;               /* arg3                 */
    koji_u64 r8;                /* arg4                 */
    koji_u64 r9;                /* arg5                 */

    /* callee-saved (preserved across syscall) */
    koji_u64 rbx;
    koji_u64 rbp;
    koji_u64 r12;
    koji_u64 r13;
    koji_u64 r14;
    koji_u64 r15;
} koji_syscall_frame_t;

/* ======================================================================
 * 9. IPC Message Header
 *
 * Placed at the start of every IPC message buffer.
 * ====================================================================== */

typedef struct koji_ipc_header {
    koji_u32 msg_type;
    koji_u32 msg_size;          /* payload size in bytes (excluding header) */
    koji_u32 handle_count;      /* number of handles being transferred     */
    koji_u32 _reserved0;        /* must be zero                            */
} koji_ipc_header_t;

/* ======================================================================
 * 10. Fixed Constants
 * ====================================================================== */

#define KOJI_IPC_MAX_HANDLES      8
#define KOJI_IPC_MAX_MSG_SIZE     4096   /* bytes, including header */
#define KOJI_MAX_THREADS_PER_PROC 256
#define KOJI_PAGE_SIZE            4096
#define KOJI_KERNEL_STACK_SIZE    16384  /* 4 pages */

#endif /* KOJI_ABI_V1_H */
