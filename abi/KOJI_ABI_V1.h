/*
 * KOJI Syscall ABI — Canonical Definition
 * Version: 1.1.0
 * Status:  Draft
 *
 * This file is the SINGLE SOURCE OF TRUTH for the KOJI ABI.
 * All language bindings (Odin, Go) are mechanically generated from this header.
 * DO NOT hand-edit generated artifacts; regenerate with: python3 tools/abi/gen_abi.py
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
 *
 * Changes from v1.0.0:
 *   - Syscall numbers reorganized: handles first (0-2), then process/thread
 *     lifecycle (3-9), IPC (10-13), ports (14-16), memory (17-22), IRQ (23-24),
 *     misc (25). Removed NOOP placeholder.
 *   - Error code ordering corrected: ERR_INVALID_ARGS=3, ERR_ACCESS_DENIED=4,
 *     ERR_NO_MEMORY=5. Replaced ERR_CANCELLED/ERR_PEER_CLOSED with
 *     ERR_CHANNEL_CLOSED/ERR_WOULD_BLOCK/ERR_INTERNAL to match kernel impl.
 *   - Syscall frame simplified: callee-saved registers (rbx, rbp, r12-r15)
 *     removed. Preserved by x86_64 calling convention; not needed in frame.
 *     Field order reordered to match NASM push sequence (syscall_num first).
 *   - IPC header fields renamed and reordered to match kernel impl:
 *     data_size, handle_count, ordinal, flags.
 *   - Added missing syscalls: SYS_PROCESS_INFO, SYS_THREAD_SUSPEND,
 *     SYS_THREAD_RESUME, SYS_VMO_READ, SYS_VMO_WRITE, SYS_VMAR_PROTECT,
 *     SYS_ABI_INFO. Renamed SYS_HANDLE_DUP to SYS_HANDLE_DUPLICATE.
 *   - IPC_MAX_MSG_SIZE renamed to IPC_MAX_DATA_BYTES (excludes header).
 */

#ifndef KOJI_ABI_V1_H
#define KOJI_ABI_V1_H

/* ======================================================================
 * 1. ABI Identity
 * ====================================================================== */

#define KOJI_ABI_MAGIC            ((koji_u32)0x4B4F4A49)  /* "KOJI" in ASCII */
#define KOJI_ABI_VERSION_MAJOR    ((koji_u32)1)
#define KOJI_ABI_VERSION_MINOR    ((koji_u32)1)
#define KOJI_ABI_VERSION_PATCH    ((koji_u32)0)

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

#define KOJI_HANDLE_GEN_SHIFT     ((koji_u32)24)
#define KOJI_HANDLE_GEN_MASK      ((koji_u32)0xFF000000)
#define KOJI_HANDLE_INDEX_MASK    ((koji_u32)0x00FFFFFF)

#define KOJI_HANDLE_MAX_INDEX     ((koji_u32)0x00FFFFFF)  /* 16,777,215 entries */
#define KOJI_HANDLE_MAX_GEN       ((koji_u8)0xFF)         /* 255 generations    */

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
#define KOJI_OBJ_IRQ              ((koji_obj_type_t)7)  /* IRQ binding object            */

#define KOJI_OBJ_TYPE_COUNT       ((koji_u32)8)

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
 *
 * Ordering: validation errors (1-5) precede resource errors (6-8),
 * then channel/blocking errors (9-11), then internal (12).
 * ====================================================================== */

typedef koji_i32 koji_status_t;

#define KOJI_OK                   ((koji_status_t)0)
#define KOJI_ERR_INVALID_HANDLE   ((koji_status_t)1)
#define KOJI_ERR_INVALID_SYSCALL  ((koji_status_t)2)
#define KOJI_ERR_INVALID_ARGS     ((koji_status_t)3)
#define KOJI_ERR_ACCESS_DENIED    ((koji_status_t)4)
#define KOJI_ERR_NO_MEMORY        ((koji_status_t)5)
#define KOJI_ERR_NOT_FOUND        ((koji_status_t)6)
#define KOJI_ERR_ALREADY_EXISTS   ((koji_status_t)7)
#define KOJI_ERR_BUFFER_TOO_SMALL ((koji_status_t)8)
#define KOJI_ERR_CHANNEL_CLOSED   ((koji_status_t)9)
#define KOJI_ERR_TIMED_OUT        ((koji_status_t)10)
#define KOJI_ERR_WOULD_BLOCK      ((koji_status_t)11)
#define KOJI_ERR_INTERNAL         ((koji_status_t)12)

#define KOJI_ERR_COUNT            ((koji_u32)13)

/* ======================================================================
 * 7. Syscall Numbers
 *
 * Stable numbering. Gaps may exist; new syscalls append only.
 * Numbers are never reused after removal (per VERSIONING_POLICY.md).
 *
 * Grouping (logical, not enforced by hardware):
 *   0–2   Handles
 *   3–9   Process / Thread lifecycle
 *   10–13 IPC channels
 *   14–16 Ports
 *   17–19 VMO (Virtual Memory Objects)
 *   20–22 VMAR (Virtual Memory Address Regions)
 *   23–24 IRQ
 *   25    Miscellaneous / introspection
 * ====================================================================== */

typedef koji_u32 koji_syscall_t;

/* -- Handles ---------------------------------------------------------- */
#define KOJI_SYS_HANDLE_CLOSE     ((koji_syscall_t)0)
#define KOJI_SYS_HANDLE_DUPLICATE ((koji_syscall_t)1)
#define KOJI_SYS_HANDLE_REPLACE   ((koji_syscall_t)2)

/* -- Process / Thread lifecycle --------------------------------------- */
#define KOJI_SYS_PROCESS_CREATE   ((koji_syscall_t)3)
#define KOJI_SYS_PROCESS_EXIT     ((koji_syscall_t)4)
#define KOJI_SYS_PROCESS_INFO     ((koji_syscall_t)5)
#define KOJI_SYS_THREAD_CREATE    ((koji_syscall_t)6)
#define KOJI_SYS_THREAD_EXIT      ((koji_syscall_t)7)
#define KOJI_SYS_THREAD_SUSPEND   ((koji_syscall_t)8)
#define KOJI_SYS_THREAD_RESUME    ((koji_syscall_t)9)

/* -- IPC channels ----------------------------------------------------- */
#define KOJI_SYS_CHANNEL_CREATE   ((koji_syscall_t)10)
#define KOJI_SYS_CHANNEL_SEND     ((koji_syscall_t)11)
#define KOJI_SYS_CHANNEL_RECV     ((koji_syscall_t)12)
#define KOJI_SYS_CHANNEL_CALL     ((koji_syscall_t)13)  /* send + recv */

/* -- Ports ------------------------------------------------------------ */
#define KOJI_SYS_PORT_CREATE      ((koji_syscall_t)14)
#define KOJI_SYS_PORT_SIGNAL      ((koji_syscall_t)15)
#define KOJI_SYS_PORT_WAIT        ((koji_syscall_t)16)

/* -- VMO (Virtual Memory Objects) ------------------------------------ */
#define KOJI_SYS_VMO_CREATE       ((koji_syscall_t)17)
#define KOJI_SYS_VMO_READ         ((koji_syscall_t)18)
#define KOJI_SYS_VMO_WRITE        ((koji_syscall_t)19)

/* -- VMAR (Virtual Memory Address Regions) --------------------------- */
#define KOJI_SYS_VMAR_MAP         ((koji_syscall_t)20)
#define KOJI_SYS_VMAR_UNMAP       ((koji_syscall_t)21)
#define KOJI_SYS_VMAR_PROTECT     ((koji_syscall_t)22)

/* -- IRQ -------------------------------------------------------------- */
#define KOJI_SYS_IRQ_BIND         ((koji_syscall_t)23)
#define KOJI_SYS_IRQ_ACK          ((koji_syscall_t)24)

/* -- Miscellaneous / introspection ----------------------------------- */
#define KOJI_SYS_ABI_INFO         ((koji_syscall_t)25)

#define KOJI_SYSCALL_COUNT        ((koji_u32)26)

/* ======================================================================
 * 8. Syscall Frame Layout
 *
 * When transitioning Ring 3 → Ring 0 via SYSCALL, the kernel saves
 * user context into this frame on the kernel stack.
 *
 * Field order matches the NASM push sequence in syscall_entry.asm:
 *   push rax (syscall_num), rdi (arg0), rsi (arg1), rdx (arg2),
 *   r10 (arg3), r8 (arg4), r9 (arg5), rcx (user_rip),
 *   r11 (user_rflags), [user_rsp from scratch]
 *
 * Callee-saved registers (rbx, rbp, r12–r15) are NOT included.
 * They are preserved by the x86_64 calling convention and are
 * not part of the syscall ABI contract.
 *
 * Size: 10 fields × 8 bytes = 80 bytes.
 * ====================================================================== */

typedef struct koji_syscall_frame {
    koji_u64 syscall_num;   /* rax — syscall number                    */
    koji_u64 arg0;          /* rdi — argument 0 (typically: handle)    */
    koji_u64 arg1;          /* rsi — argument 1                        */
    koji_u64 arg2;          /* rdx — argument 2 / secondary return     */
    koji_u64 arg3;          /* r10 — argument 3 (rcx clobbered)        */
    koji_u64 arg4;          /* r8  — argument 4                        */
    koji_u64 arg5;          /* r9  — argument 5                        */
    koji_u64 user_rip;      /* rcx saved by SYSCALL instruction        */
    koji_u64 user_rflags;   /* r11 saved by SYSCALL instruction        */
    koji_u64 user_rsp;      /* user stack pointer at syscall entry     */
} koji_syscall_frame_t;

/* ======================================================================
 * 9. IPC Message Header
 *
 * Placed at the start of every IPC message buffer.
 * Total size: 16 bytes (4 × u32, naturally aligned).
 * ====================================================================== */

typedef struct koji_ipc_header {
    koji_u32 data_size;     /* payload bytes following the header          */
    koji_u32 handle_count;  /* number of handles being transferred         */
    koji_u32 ordinal;       /* method / message type discriminant          */
    koji_u32 flags;         /* reserved; must be zero in v1                */
} koji_ipc_header_t;

/* ======================================================================
 * 10. Fixed Constants
 * ====================================================================== */

#define KOJI_IPC_MAX_DATA_BYTES   ((koji_u32)4096)   /* max payload, excl. header */
#define KOJI_IPC_MAX_HANDLES      ((koji_u32)8)
#define KOJI_MAX_THREADS_PER_PROC ((koji_u32)256)
#define KOJI_PAGE_SIZE            ((koji_u32)4096)
#define KOJI_KERNEL_STACK_SIZE    ((koji_u32)16384)  /* 4 pages */

#endif /* KOJI_ABI_V1_H */
