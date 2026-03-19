// ============================================================
// KOJI Kernel — IPC Channel Substrate (Phase 6)
//
// Defines the internal endpoint (channel) object and the kernel-side
// message representation used for all IPC operations.
//
// This file covers:
//   • Channel_Object — endpoint representation and lifecycle
//   • Msg_Envelope   — kernel-side message buffer
//   • Header and payload validation helpers
//   • Capability-transfer validation
//
// What is NOT here (deferred per phase plan):
//   • Blocking send/recv semantics (requires Phase 4 scheduler)
//   • Actual message queuing (requires Phase 3 object lifetime for peers)
//   • SYS_CHANNEL_SEND/RECV/CALL handlers (blocked by CCR-002)
//
// Channel Model
// -------------
//   Channels are created in pairs (a "peer" relationship).
//   Each endpoint is a Channel_Object.  The two endpoints reference
//   each other via peer_handle (a capability handle to the peer endpoint).
//
//   State: Open → Peer_Closed (peer destroyed) → Closed (self destroyed)
//
// Message Envelope
// ----------------
//   Every message occupies a Msg_Envelope in the kernel.
//   The envelope validates:
//     • data_size  ≤ IPC_MAX_DATA_BYTES
//     • handle_count ≤ IPC_MAX_HANDLES
//     • header.flags == 0 (reserved; must be zero in v1)
//     • each transferred handle is valid and held by the sender
//
// Badge Model
// -----------
//   Badges are u32 values attached at channel creation or delegation.
//   They are stamped into the received message header's ordinal field
//   on the receiver side by the kernel.  A zero badge is valid.
//   Badges are per-capability (stored in Cap_Entry.rights upper bits
//   is NOT the design; badges are stored separately in the future).
//   For Phase 6 the badge field exists in the envelope but is not yet
//   enforced by capability delegation.
//
// CCR References
// --------------
//   CCR-002: IPC blocking, synchronisation, and scheduling interaction
// ============================================================
package kernel

import abi "../abi/generated/odin"

// ---- Channel State ----

Channel_State :: enum u32 {
	Open        = 0, // both endpoints alive
	Peer_Closed = 1, // peer endpoint was destroyed; reads may still drain
	Closed      = 2, // this endpoint destroyed; send attempts fail
}

// ---- Channel Object ----
//
// Obj_Header MUST be the first field (see kernel/object.odin).
Channel_Object :: struct {
	header:      Obj_Header,    // lifetime management; first field — do not reorder
	state:       Channel_State,
	peer_handle: abi.Handle,    // handle to the peer channel endpoint; HANDLE_INVALID if none
	badge:       u32,           // badge value stamped into received messages (Phase 6+)
	_pad:        [4]u8,
}

// ---- Channel Pool ----

CHANNEL_POOL_SIZE :: 128 // max concurrent channel endpoints for v1

g_channel_pool: [CHANNEL_POOL_SIZE]Channel_Object

// ---- channel_alloc ----
//
// Allocates a channel endpoint from the static pool.
// Returns nil if the pool is exhausted.
// ref_count is set to 1 via obj_init; caller must call cap_alloc.
channel_alloc :: proc "c" (badge: u32) -> ^Channel_Object {
	// A slot is available when ref_count == 0 (fresh or fully destroyed).
	for i in 0 ..< CHANNEL_POOL_SIZE {
		c := &g_channel_pool[i]
		if c.header.ref_count == 0 {
			c^ = Channel_Object{}
			obj_init(&c.header, abi.OBJ_CHANNEL, channel_destroy)
			c.state       = .Open
			c.peer_handle = abi.HANDLE_INVALID
			c.badge       = badge
			return c
		}
	}
	return nil
}

// ---- channel_destroy ----
//
// Destruction hook: called when the last capability to this endpoint is
// closed.  Notifies the peer (if alive) that this side is gone.
@(private)
channel_destroy :: proc "c" (hdr: ^Obj_Header) {
	c := transmute(^Channel_Object)hdr
	c.state = .Closed

	// Mark peer as peer-closed so it knows sends will fail.
	if c.peer_handle != abi.HANDLE_INVALID {
		peer_entry := cap_lookup(c.peer_handle)
		if peer_entry != nil && peer_entry.object.obj_type == abi.OBJ_CHANNEL {
			peer_chan := transmute(^Channel_Object)(peer_entry.object)
			if peer_chan.state == .Open {
				peer_chan.state = .Peer_Closed
			}
		}
		// Do NOT close peer_handle here — the peer's own lifecycle manages it.
		c.peer_handle = abi.HANDLE_INVALID
	}
}

// ============================================================
// Msg_Envelope — Kernel-Side Message Representation
// ============================================================

// Msg_Envelope is the kernel's internal representation of an in-flight
// IPC message.  It holds validated, kernel-owned copies of all message
// components.  User-mode buffers are never directly referenced after
// initial copy-in (no TOCTOU).
//
// Phase 6: validation helpers and structure only.  Actual queuing and
// blocking semantics are deferred to the scheduler phase (CCR-002).
Msg_Envelope :: struct {
	header:       abi.Ipc_Header,                   // validated IPC header
	data:         [abi.IPC_MAX_DATA_BYTES]u8,        // payload bytes (data_size valid bytes)
	handles:      [abi.IPC_MAX_HANDLES]abi.Handle,  // transferred handles (handle_count entries)
	sender_badge: u32,                              // badge from the sending capability
	_pad:         [4]u8,
}

// ---- ipc_validate_header ----
//
// Validates an IPC header received from user mode.
// Returns:
//   OK                  — header is well-formed
//   ERR_INVALID_ARGS    — data_size exceeds max, handle_count exceeds max,
//                         or flags is non-zero (reserved in v1)
//   ERR_BUFFER_TOO_SMALL — (reserved for future use; not currently returned)
ipc_validate_header :: proc "c" (hdr: ^abi.Ipc_Header) -> abi.Status {
	if hdr.data_size > abi.IPC_MAX_DATA_BYTES {
		return abi.ERR_INVALID_ARGS
	}
	if hdr.handle_count > abi.IPC_MAX_HANDLES {
		return abi.ERR_INVALID_ARGS
	}
	// flags must be zero in v1 (reserved for future extensions).
	if hdr.flags != 0 {
		return abi.ERR_INVALID_ARGS
	}
	return abi.OK
}

// ---- ipc_validate_cap_transfer ----
//
// Validates that each handle in a transfer list:
//   1. is a valid, live capability (cap_lookup succeeds)
//   2. holds RIGHT_TRANSFER
//
// Returns:
//   OK                — all handles are valid for transfer
//   ERR_INVALID_HANDLE — at least one handle is invalid or stale
//   ERR_ACCESS_DENIED  — at least one handle lacks RIGHT_TRANSFER
ipc_validate_cap_transfer :: proc "c" (handles: [^]abi.Handle, count: u32) -> abi.Status {
	for i in 0 ..< count {
		entry := cap_lookup(handles[i])
		if entry == nil {
			return abi.ERR_INVALID_HANDLE
		}
		if !cap_has_rights(entry, abi.RIGHT_TRANSFER) {
			return abi.ERR_ACCESS_DENIED
		}
	}
	return abi.OK
}

// ---- ipc_validate_payload_size ----
//
// Verifies that data_size bytes fit within the provided buffer length.
// Returns ERR_BUFFER_TOO_SMALL if buffer_len < data_size.
ipc_validate_payload_size :: proc "c" (data_size: u32, buffer_len: u32) -> abi.Status {
	if buffer_len < data_size {
		return abi.ERR_BUFFER_TOO_SMALL
	}
	return abi.OK
}

// ---- Compile-time invariant ----
#assert(size_of(Msg_Envelope) % 8 == 0, "Msg_Envelope must be 8-byte aligned in size")
