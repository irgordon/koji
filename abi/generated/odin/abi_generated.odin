// ============================================================
// KOJI ABI — Generated Odin Bindings
// Source: abi/KOJI_ABI_V1.h (v1.0.0)
// DO NOT EDIT — regenerate from canonical header.
// ============================================================
package koji_abi

// ---- ABI Identity ----

ABI_MAGIC         :: 0x4B4F4A49   // "KOJI"
ABI_VERSION_MAJOR :: 1
ABI_VERSION_MINOR :: 0
ABI_VERSION_PATCH :: 0

// ---- Handle Encoding ----

Handle :: distinct u32

HANDLE_INVALID    :: Handle(0xFFFFFFFF)
HANDLE_GEN_SHIFT  :: 24
HANDLE_GEN_MASK   :: u32(0xFF000000)
HANDLE_INDEX_MASK :: u32(0x00FFFFFF)
HANDLE_MAX_INDEX  :: u32(0x00FFFFFF)
HANDLE_MAX_GEN    :: u8(0xFF)

handle_index :: #force_inline proc(h: Handle) -> u32 {
	return u32(h) & HANDLE_INDEX_MASK
}

handle_gen :: #force_inline proc(h: Handle) -> u8 {
	return u8((u32(h) & HANDLE_GEN_MASK) >> HANDLE_GEN_SHIFT)
}

handle_make :: #force_inline proc(index: u32, gen: u8) -> Handle {
	return Handle((u32(gen) << HANDLE_GEN_SHIFT) | (index & HANDLE_INDEX_MASK))
}

// ---- Object Taxonomy ----

Obj_Type :: distinct u32

OBJ_NONE    :: Obj_Type(0)
OBJ_PROCESS :: Obj_Type(1)
OBJ_THREAD  :: Obj_Type(2)
OBJ_VMAR    :: Obj_Type(3)
OBJ_VMO     :: Obj_Type(4)
OBJ_CHANNEL :: Obj_Type(5)
OBJ_PORT    :: Obj_Type(6)
OBJ_IRQ     :: Obj_Type(7)

OBJ_TYPE_COUNT :: 8

// ---- Rights Bitmask ----

Rights :: distinct u32

RIGHT_NONE      :: Rights(0)
RIGHT_READ      :: Rights(1 << 0)
RIGHT_WRITE     :: Rights(1 << 1)
RIGHT_EXECUTE   :: Rights(1 << 2)
RIGHT_DUPLICATE :: Rights(1 << 3)
RIGHT_TRANSFER  :: Rights(1 << 4)
RIGHT_MAP       :: Rights(1 << 5)
RIGHT_SIGNAL    :: Rights(1 << 6)
RIGHT_MANAGE    :: Rights(1 << 7)
RIGHTS_ALL      :: Rights(0xFF)

// ---- Status / Error Codes ----

Status :: distinct i32

OK                     :: Status(0)
ERR_INVALID_HANDLE     :: Status(1)
ERR_INVALID_SYSCALL    :: Status(2)
ERR_INVALID_ARGS       :: Status(3)
ERR_ACCESS_DENIED      :: Status(4)
ERR_NO_MEMORY          :: Status(5)
ERR_NOT_FOUND          :: Status(6)
ERR_ALREADY_EXISTS     :: Status(7)
ERR_BUFFER_TOO_SMALL   :: Status(8)
ERR_CHANNEL_CLOSED     :: Status(9)
ERR_TIMED_OUT          :: Status(10)
ERR_WOULD_BLOCK        :: Status(11)
ERR_INTERNAL           :: Status(12)

// ---- Syscall Numbers ----

Syscall_Num :: distinct u32

SYS_HANDLE_CLOSE       :: Syscall_Num(0)
SYS_HANDLE_DUPLICATE   :: Syscall_Num(1)
SYS_HANDLE_REPLACE     :: Syscall_Num(2)
SYS_PROCESS_CREATE     :: Syscall_Num(3)
SYS_PROCESS_EXIT       :: Syscall_Num(4)
SYS_PROCESS_INFO       :: Syscall_Num(5)
SYS_THREAD_CREATE      :: Syscall_Num(6)
SYS_THREAD_EXIT        :: Syscall_Num(7)
SYS_THREAD_SUSPEND     :: Syscall_Num(8)
SYS_THREAD_RESUME      :: Syscall_Num(9)
SYS_CHANNEL_CREATE     :: Syscall_Num(10)
SYS_CHANNEL_SEND       :: Syscall_Num(11)
SYS_CHANNEL_RECV       :: Syscall_Num(12)
SYS_CHANNEL_CALL       :: Syscall_Num(13)
SYS_PORT_CREATE        :: Syscall_Num(14)
SYS_PORT_SIGNAL        :: Syscall_Num(15)
SYS_PORT_WAIT          :: Syscall_Num(16)
SYS_VMO_CREATE         :: Syscall_Num(17)
SYS_VMO_READ           :: Syscall_Num(18)
SYS_VMO_WRITE          :: Syscall_Num(19)
SYS_VMAR_MAP           :: Syscall_Num(20)
SYS_VMAR_UNMAP         :: Syscall_Num(21)
SYS_VMAR_PROTECT       :: Syscall_Num(22)
SYS_IRQ_BIND           :: Syscall_Num(23)
SYS_IRQ_ACK            :: Syscall_Num(24)
SYS_ABI_INFO           :: Syscall_Num(25)

SYSCALL_COUNT          :: 26

// ---- Syscall Frame (matches NASM register save layout) ----

Syscall_Frame :: struct #packed {
	syscall_num: u64,   // rax on entry
	arg0:        u64,   // rdi
	arg1:        u64,   // rsi
	arg2:        u64,   // rdx
	arg3:        u64,   // r10 (rcx clobbered by SYSCALL)
	arg4:        u64,   // r8
	arg5:        u64,   // r9
	user_rip:    u64,   // rcx saved by SYSCALL
	user_rflags: u64,   // r11 saved by SYSCALL
	user_rsp:    u64,   // from kernel per-thread storage
}

// ---- IPC Message Header ----

IPC_MAX_DATA_BYTES   :: 4096
IPC_MAX_HANDLES      ::    8

Ipc_Header :: struct #packed {
	data_size:    u32,   // bytes of payload following header
	handle_count: u32,   // number of handles being transferred
	ordinal:      u32,   // method / message type discriminant
	flags:        u32,   // reserved, must be 0
}

// ---- Layout Assertions ----
// Odin does not have static_assert, but #assert works at compile time.

#assert(size_of(Handle)        == 4)
#assert(size_of(Obj_Type)      == 4)
#assert(size_of(Rights)        == 4)
#assert(size_of(Status)        == 4)
#assert(size_of(Syscall_Num)   == 4)
#assert(size_of(Syscall_Frame) == 80)
#assert(size_of(Ipc_Header)    == 16)
