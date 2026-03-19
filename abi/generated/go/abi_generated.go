// AUTO-GENERATED from abi/KOJI_ABI_V1.h
// Do not edit manually.
// Regenerate with: python3 tools/abi/gen_abi.py

package kojiabi

import "unsafe"

// ---- Type Aliases ----

type Handle uint32
type ObjType uint32
type Rights uint32
type Status int32
type SyscallNum uint32

// ---- Constants ----

const (
	AbiMagic = uint32(0x4b4f4a49) // "KOJI" in ASCII
	AbiVersionMajor = uint32(1)
	AbiVersionMinor = uint32(1)
	AbiVersionPatch = uint32(0)

	HandleInvalid = Handle(0xffffffff)
	HandleGenShift = uint32(24)
	HandleGenMask = uint32(0xff000000)
	HandleIndexMask = uint32(0xffffff)
	HandleMaxIndex = uint32(0xffffff) // 16,777,215 entries
	HandleMaxGen = uint8(0xff) // 255 generations

	ObjNone = ObjType(0)
	ObjProcess = ObjType(1)
	ObjThread = ObjType(2)
	ObjVmar = ObjType(3) // Virtual Memory Address Region
	ObjVmo = ObjType(4) // Virtual Memory Object
	ObjChannel = ObjType(5) // IPC channel endpoint
	ObjPort = ObjType(6) // async notification port
	ObjIrq = ObjType(7) // IRQ binding object
	ObjTypeCount = uint32(8)

	RightNone = Rights(0)
	RightRead = Rights(1)
	RightWrite = Rights(2)
	RightExecute = Rights(4)
	RightDuplicate = Rights(8)
	RightTransfer = Rights(16)
	RightMap = Rights(32)
	RightSignal = Rights(64)
	RightManage = Rights(128)

	RightsAll = Rights(0xff)

	Ok = Status(0)

	ErrInvalidHandle = Status(1)
	ErrInvalidSyscall = Status(2)
	ErrInvalidArgs = Status(3)
	ErrAccessDenied = Status(4)
	ErrNoMemory = Status(5)
	ErrNotFound = Status(6)
	ErrAlreadyExists = Status(7)
	ErrBufferTooSmall = Status(8)
	ErrChannelClosed = Status(9)
	ErrTimedOut = Status(10)
	ErrWouldBlock = Status(11)
	ErrInternal = Status(12)
	ErrCount = uint32(13)

	SysHandleClose = SyscallNum(0)
	SysHandleDuplicate = SyscallNum(1)
	SysHandleReplace = SyscallNum(2)
	SysProcessCreate = SyscallNum(3)
	SysProcessExit = SyscallNum(4)
	SysProcessInfo = SyscallNum(5)
	SysThreadCreate = SyscallNum(6)
	SysThreadExit = SyscallNum(7)
	SysThreadSuspend = SyscallNum(8)
	SysThreadResume = SyscallNum(9)
	SysChannelCreate = SyscallNum(10)
	SysChannelSend = SyscallNum(11)
	SysChannelRecv = SyscallNum(12)
	SysChannelCall = SyscallNum(13) // send + recv
	SysPortCreate = SyscallNum(14)
	SysPortSignal = SyscallNum(15)
	SysPortWait = SyscallNum(16)
	SysVmoCreate = SyscallNum(17)
	SysVmoRead = SyscallNum(18)
	SysVmoWrite = SyscallNum(19)
	SysVmarMap = SyscallNum(20)
	SysVmarUnmap = SyscallNum(21)
	SysVmarProtect = SyscallNum(22)
	SysIrqBind = SyscallNum(23)
	SysIrqAck = SyscallNum(24)
	SysAbiInfo = SyscallNum(25)

	SyscallCount = uint32(26)

	IpcMaxDataBytes = uint32(4096) // max payload, excl. header
	IpcMaxHandles = uint32(8)

	MaxThreadsPerProc = uint32(256)

	PageSize = uint32(4096)

	KernelStackSize = uint32(16384) // 4 pages
)

// ---- Structs ----

type SyscallFrame struct {
	SyscallNum       uint64 // rax — syscall number
	Arg0             uint64 // rdi — argument 0 (typically: handle)
	Arg1             uint64 // rsi — argument 1
	Arg2             uint64 // rdx — argument 2 / secondary return
	Arg3             uint64 // r10 — argument 3 (rcx clobbered)
	Arg4             uint64 // r8  — argument 4
	Arg5             uint64 // r9  — argument 5
	UserRip          uint64 // rcx saved by SYSCALL instruction
	UserRflags       uint64 // r11 saved by SYSCALL instruction
	UserRsp          uint64 // user stack pointer at syscall entry
}

type IpcHeader struct {
	DataSize         uint32 // payload bytes following the header
	HandleCount      uint32 // number of handles being transferred
	Ordinal          uint32 // method / message type discriminant
	Flags            uint32 // reserved; must be zero in v1
}

// ---- Layout Assertions ----

var (
	_ [80 - unsafe.Sizeof(SyscallFrame{})]struct{} // fails if struct > 80 bytes
	_ [unsafe.Sizeof(SyscallFrame{}) - 80]struct{} // fails if struct < 80 bytes
	_ [16 - unsafe.Sizeof(IpcHeader{})]struct{} // fails if struct > 16 bytes
	_ [unsafe.Sizeof(IpcHeader{}) - 16]struct{} // fails if struct < 16 bytes
)
