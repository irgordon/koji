// AUTO-GENERATED from abi/KOJI_ABI_V1.h
// Do not edit manually.
// Regenerate with: python3 tools/abi/gen_abi.py
package koji_abi

// ---- Type Aliases ----

Handle :: distinct u32
Obj_Type :: distinct u32
Rights :: distinct u32
Status :: distinct i32
Syscall_Num :: distinct u32

// ---- Constants ----

ABI_MAGIC                    :: u32(0x4b4f4a49)   // "KOJI" in ASCII
ABI_VERSION_MAJOR            :: u32(1)
ABI_VERSION_MINOR            :: u32(1)
ABI_VERSION_PATCH            :: u32(0)

HANDLE_INVALID               :: Handle(0xffffffff)
HANDLE_GEN_SHIFT             :: u32(24)
HANDLE_GEN_MASK              :: u32(0xff000000)
HANDLE_INDEX_MASK            :: u32(0xffffff)
HANDLE_MAX_INDEX             :: u32(0xffffff)   // 16,777,215 entries
HANDLE_MAX_GEN               :: u8(0xff)   // 255 generations

OBJ_NONE                     :: Obj_Type(0)
OBJ_PROCESS                  :: Obj_Type(1)
OBJ_THREAD                   :: Obj_Type(2)
OBJ_VMAR                     :: Obj_Type(3)   // Virtual Memory Address Region
OBJ_VMO                      :: Obj_Type(4)   // Virtual Memory Object
OBJ_CHANNEL                  :: Obj_Type(5)   // IPC channel endpoint
OBJ_PORT                     :: Obj_Type(6)   // async notification port
OBJ_IRQ                      :: Obj_Type(7)   // IRQ binding object
OBJ_TYPE_COUNT               :: u32(8)

RIGHT_NONE                   :: Rights(0)
RIGHT_READ                   :: Rights(1)
RIGHT_WRITE                  :: Rights(2)
RIGHT_EXECUTE                :: Rights(4)
RIGHT_DUPLICATE              :: Rights(8)
RIGHT_TRANSFER               :: Rights(16)
RIGHT_MAP                    :: Rights(32)
RIGHT_SIGNAL                 :: Rights(64)
RIGHT_MANAGE                 :: Rights(128)

RIGHTS_ALL                   :: Rights(0xff)

OK                           :: Status(0)

ERR_INVALID_HANDLE           :: Status(1)
ERR_INVALID_SYSCALL          :: Status(2)
ERR_INVALID_ARGS             :: Status(3)
ERR_ACCESS_DENIED            :: Status(4)
ERR_NO_MEMORY                :: Status(5)
ERR_NOT_FOUND                :: Status(6)
ERR_ALREADY_EXISTS           :: Status(7)
ERR_BUFFER_TOO_SMALL         :: Status(8)
ERR_CHANNEL_CLOSED           :: Status(9)
ERR_TIMED_OUT                :: Status(10)
ERR_WOULD_BLOCK              :: Status(11)
ERR_INTERNAL                 :: Status(12)
ERR_COUNT                    :: u32(13)

SYS_HANDLE_CLOSE             :: Syscall_Num(0)
SYS_HANDLE_DUPLICATE         :: Syscall_Num(1)
SYS_HANDLE_REPLACE           :: Syscall_Num(2)
SYS_PROCESS_CREATE           :: Syscall_Num(3)
SYS_PROCESS_EXIT             :: Syscall_Num(4)
SYS_PROCESS_INFO             :: Syscall_Num(5)
SYS_THREAD_CREATE            :: Syscall_Num(6)
SYS_THREAD_EXIT              :: Syscall_Num(7)
SYS_THREAD_SUSPEND           :: Syscall_Num(8)
SYS_THREAD_RESUME            :: Syscall_Num(9)
SYS_CHANNEL_CREATE           :: Syscall_Num(10)
SYS_CHANNEL_SEND             :: Syscall_Num(11)
SYS_CHANNEL_RECV             :: Syscall_Num(12)
SYS_CHANNEL_CALL             :: Syscall_Num(13)   // send + recv
SYS_PORT_CREATE              :: Syscall_Num(14)
SYS_PORT_SIGNAL              :: Syscall_Num(15)
SYS_PORT_WAIT                :: Syscall_Num(16)
SYS_VMO_CREATE               :: Syscall_Num(17)
SYS_VMO_READ                 :: Syscall_Num(18)
SYS_VMO_WRITE                :: Syscall_Num(19)
SYS_VMAR_MAP                 :: Syscall_Num(20)
SYS_VMAR_UNMAP               :: Syscall_Num(21)
SYS_VMAR_PROTECT             :: Syscall_Num(22)
SYS_IRQ_BIND                 :: Syscall_Num(23)
SYS_IRQ_ACK                  :: Syscall_Num(24)
SYS_ABI_INFO                 :: Syscall_Num(25)

SYSCALL_COUNT                :: u32(26)

IPC_MAX_DATA_BYTES           :: u32(4096)   // max payload, excl. header
IPC_MAX_HANDLES              :: u32(8)

MAX_THREADS_PER_PROC         :: u32(256)

PAGE_SIZE                    :: u32(4096)

KERNEL_STACK_SIZE            :: u32(16384)   // 4 pages

// ---- Handle Helpers (bit manipulation) ----

handle_index :: #force_inline proc(h: Handle) -> u32 {
	return u32(h) & HANDLE_INDEX_MASK
}

handle_gen :: #force_inline proc(h: Handle) -> u8 {
	return u8((u32(h) & HANDLE_GEN_MASK) >> HANDLE_GEN_SHIFT)
}

handle_make :: #force_inline proc(index: u32, gen: u8) -> Handle {
	return Handle((u32(gen) << HANDLE_GEN_SHIFT) | (index & HANDLE_INDEX_MASK))
}

// ---- Structs ----

Syscall_Frame :: struct #packed {
	syscall_num:      u64,   // rax — syscall number
	arg0:             u64,   // rdi — argument 0 (typically: handle)
	arg1:             u64,   // rsi — argument 1
	arg2:             u64,   // rdx — argument 2 / secondary return
	arg3:             u64,   // r10 — argument 3 (rcx clobbered)
	arg4:             u64,   // r8  — argument 4
	arg5:             u64,   // r9  — argument 5
	user_rip:         u64,   // rcx saved by SYSCALL instruction
	user_rflags:      u64,   // r11 saved by SYSCALL instruction
	user_rsp:         u64,   // user stack pointer at syscall entry
}

Ipc_Header :: struct #packed {
	data_size:        u32,   // payload bytes following the header
	handle_count:     u32,   // number of handles being transferred
	ordinal:          u32,   // method / message type discriminant
	flags:            u32,   // reserved; must be zero in v1
}

// ---- Layout Assertions ----

#assert(size_of(Syscall_Frame) == 80)
#assert(size_of(Ipc_Header) == 16)
#assert(size_of(Handle) == 4)
#assert(size_of(Obj_Type) == 4)
#assert(size_of(Rights) == 4)
#assert(size_of(Status) == 4)
#assert(size_of(Syscall_Num) == 4)
