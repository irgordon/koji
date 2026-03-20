package private_lifetime_test

import "core:fmt"
import "core:os"
import abi "../abi/generated/odin"

Obj_State :: enum u32 { Live = 0, Dying = 1, Dead = 2 }
Obj_Destroy_Fn :: #type proc(hdr: ^Obj_Header)

Obj_Header :: struct {
	obj_type:   abi.Obj_Type,
	state:      Obj_State,
	ref_count:  u32,
	destroy_fn: Obj_Destroy_Fn,
}

obj_init :: proc(hdr: ^Obj_Header, obj_type: abi.Obj_Type, destroy: Obj_Destroy_Fn = nil) {
	hdr.obj_type   = obj_type
	hdr.state      = .Live
	hdr.ref_count  = 0
	hdr.destroy_fn = destroy
}

Thread_State :: enum u32 {
	// Keep in lockstep with kernel/thread.odin (host-side focused invariant test).
	Stopped  = 0,
	Runnable = 1,
	Blocked  = 2,
	Dying    = 3,
	Dead     = 4,
}

Thread_Object :: struct {
	header:       Obj_Header,
	state:        Thread_State,
	addr_space_h: abi.Handle,
}

thread_state_is_valid :: proc(s: Thread_State) -> bool {
	return s >= .Stopped && s <= .Dead
}

thread_invariants_hold :: proc(t: ^Thread_Object) -> bool {
	if t == nil { return false }
	if t.header.obj_type != abi.OBJ_THREAD { return false }
	if !thread_state_is_valid(t.state) { return false }
	if t.state == .Dead && t.addr_space_h != abi.HANDLE_INVALID { return false }
	return true
}

thread_set_runnable :: proc(t: ^Thread_Object) -> abi.Status {
	if !thread_invariants_hold(t) { return abi.ERR_INVALID_ARGS }
	if t.state != .Stopped && t.state != .Blocked { return abi.ERR_INVALID_ARGS }
	t.state = .Runnable
	return abi.OK
}

thread_destroy :: proc(hdr: ^Obj_Header) {
	if hdr == nil || hdr.obj_type != abi.OBJ_THREAD { return }
	t := transmute(^Thread_Object)hdr
	t.addr_space_h = abi.HANDLE_INVALID
	t.state = .Dead
}

// Keep in lockstep with kernel/address_space.odin (host-side focused invariant test).
Addr_Space_State :: enum u32 { Detached = 0, Attached = 1, Dying = 2, Dead = 3 }

Addr_Space_Object :: struct {
	header:       Obj_Header,
	state:        Addr_Space_State,
	thread_count: u32,
}

addr_space_state_is_valid :: proc(s: Addr_Space_State) -> bool {
	return s >= .Detached && s <= .Dead
}

addr_space_invariants_hold :: proc(a: ^Addr_Space_Object) -> bool {
	if a == nil { return false }
	if a.header.obj_type != abi.OBJ_VMAR { return false }
	if !addr_space_state_is_valid(a.state) { return false }
	if a.state == .Attached && a.thread_count == 0 { return false }
	if a.state == .Detached && a.thread_count != 0 { return false }
	return true
}

addr_space_attach :: proc(a: ^Addr_Space_Object) -> abi.Status {
	if !addr_space_invariants_hold(a) { return abi.ERR_INVALID_ARGS }
	if a.state == .Dying || a.state == .Dead { return abi.ERR_INVALID_ARGS }
	a.thread_count += 1
	a.state = .Attached
	return abi.OK
}

addr_space_detach :: proc(a: ^Addr_Space_Object) -> abi.Status {
	if !addr_space_invariants_hold(a) { return abi.ERR_INVALID_ARGS }
	if a.state != .Attached || a.thread_count == 0 { return abi.ERR_INVALID_ARGS }
	a.thread_count -= 1
	if a.thread_count == 0 { a.state = .Detached }
	return abi.OK
}

check :: proc(name: string, cond: bool, pass, fail: ^int) {
	if cond {
		fmt.printf("  PASS  %s\n", name)
		pass^ += 1
	} else {
		fmt.printf("  FAIL  %s\n", name)
		fail^ += 1
	}
}

main :: proc() {
	pass, fail := 0, 0

	fmt.println("[L01] thread invariant holds after initialization")
	t := Thread_Object{}
	obj_init(&t.header, abi.OBJ_THREAD, thread_destroy)
	t.state = .Stopped
	t.addr_space_h = abi.HANDLE_INVALID
	check("thread invariants hold", thread_invariants_hold(&t), &pass, &fail)
	check("thread starts in stopped", t.state == .Stopped, &pass, &fail)

	fmt.println("[L02] type separation rejects wrong object type for thread path")
	t_bad := t
	t_bad.header.obj_type = abi.OBJ_VMAR
	check("thread invariants reject wrong type", !thread_invariants_hold(&t_bad), &pass, &fail)
	check("thread_set_runnable rejects wrong type", thread_set_runnable(&t_bad) == abi.ERR_INVALID_ARGS, &pass, &fail)

	fmt.println("[L03] dead thread must not retain address-space handle")
	t_dead := t
	t_dead.state = .Dead
	t_dead.addr_space_h = abi.Handle(1)
	check("dead thread with handle breaks invariant", !thread_invariants_hold(&t_dead), &pass, &fail)

	fmt.println("[L04] address-space invariant holds after initialization")
	a := Addr_Space_Object{}
	obj_init(&a.header, abi.OBJ_VMAR)
	a.state = .Detached
	a.thread_count = 0
	check("address-space invariants hold", addr_space_invariants_hold(&a), &pass, &fail)

	fmt.println("[L05] attach/detach transitions preserve invariants")
	check("attach succeeds", addr_space_attach(&a) == abi.OK, &pass, &fail)
	check("becomes attached with count 1", a.state == .Attached && a.thread_count == 1, &pass, &fail)
	check("detach succeeds", addr_space_detach(&a) == abi.OK, &pass, &fail)
	check("returns detached with count 0", a.state == .Detached && a.thread_count == 0, &pass, &fail)

	fmt.println("[L06] invalid attached state is rejected")
	a_bad := a
	a_bad.state = .Attached
	a_bad.thread_count = 0
	check("invalid attached state fails invariant", !addr_space_invariants_hold(&a_bad), &pass, &fail)
	check("attach rejects invalid pre-state", addr_space_attach(&a_bad) == abi.ERR_INVALID_ARGS, &pass, &fail)

	fmt.printf("\n=== Private lifetime tests: %d passed, %d failed ===\n", pass, fail)
	if fail != 0 {
		os.exit(1)
	}
}
