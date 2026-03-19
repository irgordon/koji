// ============================================================
// KOJI Phase 1 — CNode Invariant Tests
//
// This file is a standalone Odin program that validates the CNode
// behavioral invariants described in the Phase 1 requirements.
//
// It reimplements the minimal CNode and object-header logic using
// normal Odin (non-freestanding) to run on the host build machine.
//
// Run with:
//   odin run tests/cnode_test.odin -file
// Or via make:
//   make test
//
// Exit code 0 = all tests passed.
// Exit code 1 = at least one assertion failed.
//
// Tests:
//   T01 — HANDLE_INVALID (0xFFFFFFFF) is always rejected by lookup
//   T02 — Handle index 0 with gen 0 is a VALID handle (not a sentinel)
//   T03 — Out-of-range slot index (>= CAP_TABLE_SIZE) returns nil
//   T04 — Empty slot (never allocated) returns nil
//   T05 — Stale generation: close increments gen, old handle fails
//   T06 — Rights-subset: duplicate narrows rights, never amplifies
//   T07 — Anti-amplification: requesting RIGHTS_ALL gives only src rights
//   T08 — cap_replace: rights narrowed, old handle invalidated, new handle valid
//   T09 — cap_close on last ref triggers obj_deref (ref_count → 0)
//   T10 — cap_duplicate without RIGHT_DUPLICATE → ERR_ACCESS_DENIED
//   T11 — cap_alloc(nil, rights) fails cleanly
//   T12 — one of many closes does not destroy; last close destroys
//   T13 — obj_deref destroy hook runs once; extra deref at zero is no-op
//   T14 — obj_is_live rejects Dying and Dead
//   T15 — out-of-range syscall returns ERR_INVALID_SYSCALL
//   T16 — in-range nil dispatch slot returns ERR_INVALID_SYSCALL
//   T17 — malformed ingress fails before handler dispatch
//   T18 — SYS_ABI_INFO rejects non-zero unused args
//   T19 — nil ingress frame returns ERR_INVALID_ARGS
// ============================================================
package cnode_test

import "core:fmt"
import "core:os"
import abi "../abi/generated/odin"

// ---- Minimal type replicas (no foreign import, no freestanding) ----

Handle    :: distinct u32
Rights    :: distinct u32
Status    :: distinct i32
Obj_Type  :: distinct u32

HANDLE_INVALID      :: Handle(0xFFFFFFFF)
HANDLE_GEN_SHIFT    :: u32(24)
HANDLE_GEN_MASK     :: u32(0xFF000000)
HANDLE_INDEX_MASK   :: u32(0x00FFFFFF)
CAP_TABLE_SIZE      :: u32(4096)

RIGHT_NONE      :: Rights(0)
RIGHT_READ      :: Rights(1 << 0)
RIGHT_WRITE     :: Rights(1 << 1)
RIGHT_DUPLICATE :: Rights(1 << 3)
RIGHT_TRANSFER  :: Rights(1 << 4)
RIGHTS_ALL      :: Rights(0xFF)

OK                   :: Status(0)
ERR_INVALID_HANDLE   :: Status(1)
ERR_INVALID_SYSCALL  :: Status(2)
ERR_INVALID_ARGS     :: Status(3)
ERR_ACCESS_DENIED    :: Status(4)
ERR_NO_MEMORY        :: Status(5)

OBJ_NONE :: Obj_Type(0)

// ---- Object Header ----

Obj_State :: enum u32 { Live = 0, Dying = 1, Dead = 2 }
Obj_Destroy_Fn :: #type proc(obj: ^Obj_Header)

Obj_Header :: struct {
	obj_type:   Obj_Type,
	state:      Obj_State,
	ref_count:  u32,
	destroy_fn: Obj_Destroy_Fn,
}

obj_init :: proc(hdr: ^Obj_Header, t: Obj_Type, destroy: Obj_Destroy_Fn = nil) {
	hdr.obj_type   = t
	hdr.state      = .Live
	hdr.ref_count  = 0
	hdr.destroy_fn = destroy
}

obj_ref :: proc(hdr: ^Obj_Header) -> bool {
	if hdr == nil { return false }
	if hdr.ref_count == 0xFFFF_FFFF { return false }
	hdr.ref_count += 1
	return true
}

obj_deref :: proc(hdr: ^Obj_Header) -> bool {
	if hdr.ref_count == 0 { return false }
	hdr.ref_count -= 1
	if hdr.ref_count != 0 { return false }
	hdr.state = .Dying
	if hdr.destroy_fn != nil { hdr.destroy_fn(hdr) }
	hdr.state = .Dead
	return true
}

obj_is_live :: proc(hdr: ^Obj_Header) -> bool {
	return hdr.state == .Live
}

// ---- Handle Helpers ----

handle_index :: proc(h: Handle) -> u32 { return u32(h) & HANDLE_INDEX_MASK }
handle_gen   :: proc(h: Handle) -> u8  { return u8((u32(h) & HANDLE_GEN_MASK) >> HANDLE_GEN_SHIFT) }
handle_make  :: proc(index: u32, gen: u8) -> Handle {
	return Handle((u32(gen) << HANDLE_GEN_SHIFT) | (index & HANDLE_INDEX_MASK))
}

// ---- Cap Entry + Table ----

Cap_Entry :: struct {
	object:     ^Obj_Header,
	rights:     Rights,
	generation: u8,
	_pad:       [3]u8,
}

Cap_Table :: struct {
	entries: [CAP_TABLE_SIZE]Cap_Entry,
}

g_cap_table: Cap_Table

cap_table_init :: proc() {
	for i in 0..<CAP_TABLE_SIZE {
		g_cap_table.entries[i] = Cap_Entry{object=nil, rights=RIGHT_NONE, generation=0}
	}
}

cap_alloc :: proc(obj: ^Obj_Header, rights: Rights) -> Handle {
	if obj == nil { return HANDLE_INVALID }
	if !obj_is_live(obj) { return HANDLE_INVALID }
	for i in 0..<CAP_TABLE_SIZE {
		if g_cap_table.entries[i].object == nil {
			if !obj_ref(obj) { return HANDLE_INVALID }
			gen := g_cap_table.entries[i].generation
			g_cap_table.entries[i] = Cap_Entry{object=obj, rights=rights, generation=gen}
			return handle_make(i, gen)
		}
	}
	return HANDLE_INVALID
}

cap_lookup :: proc(h: Handle) -> ^Cap_Entry {
	if h == HANDLE_INVALID { return nil }
	idx := handle_index(h)
	gen := handle_gen(h)
	if idx >= CAP_TABLE_SIZE { return nil }
	entry := &g_cap_table.entries[idx]
	if entry.object == nil { return nil }
	if entry.generation != gen { return nil }
	if !obj_is_live(entry.object) { return nil }
	return entry
}

cap_has_rights :: proc(entry: ^Cap_Entry, required: Rights) -> bool {
	if entry == nil { return false }
	return (entry.rights & required) == required
}

cap_close :: proc(h: Handle) -> Status {
	entry := cap_lookup(h)
	if entry == nil { return ERR_INVALID_HANDLE }
	obj := entry.object
	entry.object     = nil
	entry.rights     = RIGHT_NONE
	entry.generation += 1
	obj_deref(obj)
	return OK
}

cap_duplicate :: proc(h: Handle, requested_rights: Rights) -> (Handle, Status) {
	src := cap_lookup(h)
	if src == nil { return HANDLE_INVALID, ERR_INVALID_HANDLE }
	if !cap_has_rights(src, RIGHT_DUPLICATE) { return HANDLE_INVALID, ERR_ACCESS_DENIED }
	effective := src.rights & requested_rights
	if !obj_ref(src.object) { return HANDLE_INVALID, ERR_NO_MEMORY }
	new_handle := cap_alloc(src.object, effective)
	if new_handle == HANDLE_INVALID {
		obj_deref(src.object)
		return HANDLE_INVALID, ERR_NO_MEMORY
	}
	obj_deref(src.object)
	return new_handle, OK
}

cap_replace :: proc(h: Handle, new_rights: Rights) -> (Handle, Status) {
	entry := cap_lookup(h)
	if entry == nil { return HANDLE_INVALID, ERR_INVALID_HANDLE }
	idx := handle_index(h)
	effective := entry.rights & new_rights
	entry.generation += 1
	entry.rights = effective
	return handle_make(idx, entry.generation), OK
}

// ---- Test Harness ----

g_pass := 0
g_fail := 0
g_destroy_count := 0

check :: proc(name: string, cond: bool) {
	if cond {
		fmt.printf("  PASS  %s\n", name)
		g_pass += 1
	} else {
		fmt.printf("  FAIL  %s\n", name)
		g_fail += 1
	}
}

counting_destroy_fn :: proc(hdr: ^Obj_Header) {
	g_destroy_count += 1
}

// ---- Tests ----

test_t01_handle_invalid_rejected :: proc() {
	fmt.println("[T01] HANDLE_INVALID always rejected")
	entry := cap_lookup(HANDLE_INVALID)
	check("cap_lookup(HANDLE_INVALID) == nil", entry == nil)
}

test_t02_handle_zero_is_valid :: proc() {
	fmt.println("[T02] Handle value 0 (index=0, gen=0) is a valid non-sentinel handle")
	cap_table_init()
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)
	h := cap_alloc(&obj, RIGHT_READ)
	// First slot should be index=0, gen=0 → handle value = 0
	check("first alloc returns handle value 0", h == Handle(0))
	entry := cap_lookup(h)
	check("cap_lookup(0) returns non-nil when slot occupied", entry != nil)
	check("HANDLE_INVALID != 0 (sentinel is 0xFFFFFFFF)", HANDLE_INVALID != Handle(0))
}

test_t03_out_of_range_index :: proc() {
	fmt.println("[T03] Out-of-range slot index returns nil")
	// Craft a handle with index = CAP_TABLE_SIZE (exactly out of range)
	bad_handle := handle_make(CAP_TABLE_SIZE, 0)
	entry := cap_lookup(bad_handle)
	check("cap_lookup with index == CAP_TABLE_SIZE returns nil", entry == nil)

	// And a handle well beyond range
	bad2 := handle_make(0x00FFFFFF, 0) // max index field (16M-1 > 4096)
	entry2 := cap_lookup(bad2)
	check("cap_lookup with max index (> table size) returns nil", entry2 == nil)
}

test_t04_empty_slot_access :: proc() {
	fmt.println("[T04] Empty slot (never allocated) returns nil")
	cap_table_init()
	// Slot 5 has never been allocated.
	h := handle_make(5, 0)
	entry := cap_lookup(h)
	check("cap_lookup on never-allocated slot returns nil", entry == nil)
}

test_t05_stale_generation :: proc() {
	fmt.println("[T05] Stale generation after close returns nil")
	cap_table_init()
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)
	h := cap_alloc(&obj, RIGHT_READ | RIGHT_DUPLICATE)
	check("initial alloc succeeds", h != HANDLE_INVALID)

	// Close the handle.
	st := cap_close(h)
	check("cap_close returns OK", st == OK)

	// The same handle value is now stale (generation bumped).
	stale_entry := cap_lookup(h)
	check("cap_lookup with old handle returns nil after close", stale_entry == nil)

	// If we allocate a new object at the same slot (slot 0 after table reset),
	// it should use the bumped generation.
	obj2: Obj_Header
	obj_init(&obj2, OBJ_NONE)
	h2 := cap_alloc(&obj2, RIGHT_READ)
	check("re-alloc at same slot gets new generation (handle != old handle)", h2 != h)
	// The new handle should still be at index 0 but with gen=1.
	check("new handle has same index as old handle", handle_index(h2) == handle_index(h))
	check("new handle has bumped generation", handle_gen(h2) == handle_gen(h) + 1)
}

test_t06_rights_subset_copy :: proc() {
	fmt.println("[T06] Rights-subset copy: duplicate cannot exceed source rights")
	cap_table_init()
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)

	// Allocate with READ | WRITE | DUPLICATE.
	h := cap_alloc(&obj, RIGHT_READ | RIGHT_WRITE | RIGHT_DUPLICATE)
	check("alloc with R|W|DUP", h != HANDLE_INVALID)

	// Duplicate requesting READ only.
	h2, st := cap_duplicate(h, RIGHT_READ)
	check("duplicate READ-only returns OK", st == OK)
	entry2 := cap_lookup(h2)
	check("duplicated handle is valid", entry2 != nil)
	check("duplicated rights = READ only", entry2 != nil && entry2.rights == RIGHT_READ)

	// Duplicate requesting WRITE only (src has WRITE, so effective = WRITE).
	h3, st3 := cap_duplicate(h, RIGHT_WRITE)
	check("duplicate WRITE-only returns OK", st3 == OK)
	entry3 := cap_lookup(h3)
	check("duplicated WRITE rights = WRITE only", entry3 != nil && entry3.rights == RIGHT_WRITE)
}

test_t07_anti_amplification :: proc() {
	fmt.println("[T07] Anti-amplification: requesting RIGHTS_ALL gives only source rights")
	cap_table_init()
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)

	// Allocate with READ | DUPLICATE only.
	src_rights := RIGHT_READ | RIGHT_DUPLICATE
	h := cap_alloc(&obj, src_rights)

	// Duplicate requesting ALL rights — effective should be READ | DUPLICATE only.
	h2, st := cap_duplicate(h, RIGHTS_ALL)
	check("duplicate with RIGHTS_ALL returns OK", st == OK)
	entry2 := cap_lookup(h2)
	check("effective rights == source rights (no amplification)",
		entry2 != nil && entry2.rights == src_rights)
}

test_t08_cap_replace :: proc() {
	fmt.println("[T08] cap_replace: rights narrowed, old handle invalidated, new handle valid")
	cap_table_init()
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)

	h := cap_alloc(&obj, RIGHT_READ | RIGHT_WRITE | RIGHT_DUPLICATE)
	old_gen := handle_gen(h)

	new_h, st := cap_replace(h, RIGHT_READ)
	check("cap_replace returns OK", st == OK)
	check("new handle is not HANDLE_INVALID", new_h != HANDLE_INVALID)
	check("new handle has same index", handle_index(new_h) == handle_index(h))
	check("new handle has bumped generation", handle_gen(new_h) == old_gen + 1)

	// Old handle should now be stale.
	check("old handle is stale after replace", cap_lookup(h) == nil)

	// New handle should be valid with narrowed rights.
	new_entry := cap_lookup(new_h)
	check("new handle is valid", new_entry != nil)
	check("new rights = READ only", new_entry != nil && new_entry.rights == RIGHT_READ)

	// ref_count should be unchanged (replace doesn't copy).
	check("ref_count unchanged after replace", obj.ref_count == 1)
}

test_t09_last_ref_destruction :: proc() {
	fmt.println("[T09] cap_close on last reference triggers obj_deref (ref_count → 0)")
	cap_table_init()
	g_destroy_count = 0

	obj: Obj_Header
	obj_init(&obj, OBJ_NONE, counting_destroy_fn)
	check("initial ref_count = 0", obj.ref_count == 0)

	h := cap_alloc(&obj, RIGHT_READ)
	check("alloc publishes one ref", obj.ref_count == 1)

	st := cap_close(h)
	check("cap_close returns OK", st == OK)
	check("destroy hook called exactly once", g_destroy_count == 1)
	check("ref_count == 0 after last close", obj.ref_count == 0)
	check("object state is Dead after last close", obj.state == .Dead)
}

test_t10_duplicate_requires_right :: proc() {
	fmt.println("[T10] cap_duplicate without RIGHT_DUPLICATE → ERR_ACCESS_DENIED")
	cap_table_init()
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)

	// Allocate WITHOUT RIGHT_DUPLICATE.
	h := cap_alloc(&obj, RIGHT_READ | RIGHT_WRITE)

	h2, st := cap_duplicate(h, RIGHT_READ)
	check("duplicate without RIGHT_DUPLICATE returns ERR_ACCESS_DENIED", st == ERR_ACCESS_DENIED)
	check("returned handle is HANDLE_INVALID", h2 == HANDLE_INVALID)
	check("ref_count unchanged after failed duplicate", obj.ref_count == 1)
}

test_t11_cap_alloc_nil_fails :: proc() {
	fmt.println("[T11] cap_alloc(nil, rights) fails cleanly")
	cap_table_init()
	h := cap_alloc(nil, RIGHT_READ)
	check("cap_alloc(nil, RIGHT_READ) == HANDLE_INVALID", h == HANDLE_INVALID)
}

test_t12_close_multi_handle_last_destroys :: proc() {
	fmt.println("[T12] closing one of multiple handles does not destroy; last close does")
	cap_table_init()
	g_destroy_count = 0

	obj: Obj_Header
	obj_init(&obj, OBJ_NONE, counting_destroy_fn)
	h1 := cap_alloc(&obj, RIGHT_READ | RIGHT_DUPLICATE)
	check("first handle alloc succeeds", h1 != HANDLE_INVALID)
	check("ref_count is 1 after first publish", obj.ref_count == 1)

	h2, st_dup := cap_duplicate(h1, RIGHT_READ)
	check("duplicate succeeds", st_dup == OK)
	check("second handle is valid", h2 != HANDLE_INVALID)
	check("ref_count is 2 after duplicate publish", obj.ref_count == 2)

	st1 := cap_close(h1)
	check("first close returns OK", st1 == OK)
	check("object still live after closing one handle", obj.state == .Live)
	check("destroy hook not called after first close", g_destroy_count == 0)
	check("ref_count is 1 after first close", obj.ref_count == 1)

	st2 := cap_close(h2)
	check("second close returns OK", st2 == OK)
	check("destroy hook called exactly once on last close", g_destroy_count == 1)
	check("object is Dead after last close", obj.state == .Dead)
	check("ref_count is 0 after last close", obj.ref_count == 0)
}

test_t13_obj_deref_once_and_extra_noop :: proc() {
	fmt.println("[T13] obj_deref runs destroy hook once; extra deref at zero is clean no-op")
	cap_table_init()
	g_destroy_count = 0

	obj: Obj_Header
	obj_init(&obj, OBJ_NONE, counting_destroy_fn)
	h := cap_alloc(&obj, RIGHT_READ)
	check("alloc succeeds", h != HANDLE_INVALID)
	check("ref_count is 1 before close", obj.ref_count == 1)

	st := cap_close(h)
	check("close returns OK", st == OK)
	check("destroy hook called once after close", g_destroy_count == 1)

	destroyed_again := obj_deref(&obj)
	check("extra obj_deref at zero returns false", destroyed_again == false)
	check("destroy hook still called exactly once", g_destroy_count == 1)
}

test_t14_obj_live_rejects_dying_dead :: proc() {
	fmt.println("[T14] obj_is_live rejects Dying and Dead")
	obj: Obj_Header
	obj_init(&obj, OBJ_NONE)
	check("fresh object is live", obj_is_live(&obj))
	obj.state = .Dying
	check("dying object is not live", !obj_is_live(&obj))
	obj.state = .Dead
	check("dead object is not live", !obj_is_live(&obj))
}

// ---- Minimal syscall dispatch model checks ----

SYSCALL_COUNT :: u32(abi.SYSCALL_COUNT)
SYS_ABI_INFO  :: u32(abi.SYS_ABI_INFO)

Syscall_Frame :: struct {
	syscall_num: u64,
	arg0:        u64,
	arg1:        u64,
	arg2:        u64,
	arg3:        u64,
	arg4:        u64,
	arg5:        u64,
}

Syscall_Handler :: #type proc(frame: ^Syscall_Frame) -> Status

dispatch_table: [SYSCALL_COUNT]Syscall_Handler
dispatch_counter := 0

frame_validate_ingress :: proc(frame: ^Syscall_Frame) -> Status {
	if frame == nil {
		return ERR_INVALID_ARGS
	}
	if frame.syscall_num >> 32 != 0 {
		return ERR_INVALID_ARGS
	}
	return OK
}

frame_validate_unused_args :: proc(frame: ^Syscall_Frame, used_count: u32) -> Status {
	args := [6]u64{frame.arg0, frame.arg1, frame.arg2, frame.arg3, frame.arg4, frame.arg5}
	for i := used_count; i < 6; i += 1 {
		if args[i] != 0 {
			return ERR_INVALID_ARGS
		}
	}
	return OK
}

sys_abi_info_model :: proc(frame: ^Syscall_Frame) -> Status {
	dispatch_counter += 1
	if frame_validate_unused_args(frame, 0) != OK {
		return ERR_INVALID_ARGS
	}
	frame.arg2 = 0x00010100
	return OK
}

koji_syscall_dispatch_model :: proc(frame: ^Syscall_Frame) -> Status {
	if frame_validate_ingress(frame) != OK {
		return ERR_INVALID_ARGS
	}
	num := u32(frame.syscall_num)
	if num >= SYSCALL_COUNT {
		return ERR_INVALID_SYSCALL
	}
	handler := dispatch_table[num]
	if handler == nil {
		return ERR_INVALID_SYSCALL
	}
	return handler(frame)
}

test_t15_unknown_syscall_status :: proc() {
	fmt.println("[T15] unknown syscall number returns ERR_INVALID_SYSCALL")
	dispatch_counter = 0
	dispatch_table = [SYSCALL_COUNT]Syscall_Handler{}
	frame := Syscall_Frame{syscall_num = u64(SYSCALL_COUNT)}
	st := koji_syscall_dispatch_model(&frame)
	check("out-of-range syscall returns ERR_INVALID_SYSCALL", st == ERR_INVALID_SYSCALL)
	check("no handler dispatched", dispatch_counter == 0)
}

test_t16_in_range_nil_slot_status :: proc() {
	fmt.println("[T16] in-range nil dispatch slot returns ERR_INVALID_SYSCALL")
	dispatch_counter = 0
	dispatch_table = [SYSCALL_COUNT]Syscall_Handler{}
	frame := Syscall_Frame{syscall_num = 1}
	st := koji_syscall_dispatch_model(&frame)
	check("nil slot syscall returns ERR_INVALID_SYSCALL", st == ERR_INVALID_SYSCALL)
	check("no handler dispatched", dispatch_counter == 0)
}

test_t17_malformed_ingress_before_dispatch :: proc() {
	fmt.println("[T17] malformed ingress fails before handler dispatch")
	dispatch_counter = 0
	dispatch_table = [SYSCALL_COUNT]Syscall_Handler{}
	dispatch_table[SYS_ABI_INFO] = sys_abi_info_model
	frame := Syscall_Frame{syscall_num = (1 << 32) | u64(SYS_ABI_INFO)}
	st := koji_syscall_dispatch_model(&frame)
	check("malformed ingress returns ERR_INVALID_ARGS", st == ERR_INVALID_ARGS)
	check("handler not dispatched for malformed ingress", dispatch_counter == 0)
}

test_t18_sys_abi_info_unused_args :: proc() {
	fmt.println("[T18] SYS_ABI_INFO rejects non-zero unused args")
	dispatch_counter = 0
	dispatch_table = [SYSCALL_COUNT]Syscall_Handler{}
	dispatch_table[SYS_ABI_INFO] = sys_abi_info_model
	frame := Syscall_Frame{syscall_num = u64(SYS_ABI_INFO), arg0 = 1}
	st := koji_syscall_dispatch_model(&frame)
	check("SYS_ABI_INFO with non-zero arg returns ERR_INVALID_ARGS", st == ERR_INVALID_ARGS)
	check("handler dispatch counted exactly once", dispatch_counter == 1)
}

test_t19_nil_ingress_frame_rejected :: proc() {
	fmt.println("[T19] nil ingress frame returns ERR_INVALID_ARGS")
	dispatch_counter = 0
	dispatch_table = [SYSCALL_COUNT]Syscall_Handler{}
	st := koji_syscall_dispatch_model(nil)
	check("nil frame returns ERR_INVALID_ARGS", st == ERR_INVALID_ARGS)
	check("no handler dispatched", dispatch_counter == 0)
}

// ---- Main ----

main :: proc() {
	fmt.println("=== KOJI Phase 1 CNode Invariant Tests ===")
	fmt.println()

	test_t01_handle_invalid_rejected()
	test_t02_handle_zero_is_valid()
	test_t03_out_of_range_index()
	test_t04_empty_slot_access()
	test_t05_stale_generation()
	test_t06_rights_subset_copy()
	test_t07_anti_amplification()
	test_t08_cap_replace()
	test_t09_last_ref_destruction()
	test_t10_duplicate_requires_right()
	test_t11_cap_alloc_nil_fails()
	test_t12_close_multi_handle_last_destroys()
	test_t13_obj_deref_once_and_extra_noop()
	test_t14_obj_live_rejects_dying_dead()
	test_t15_unknown_syscall_status()
	test_t16_in_range_nil_slot_status()
	test_t17_malformed_ingress_before_dispatch()
	test_t18_sys_abi_info_unused_args()
	test_t19_nil_ingress_frame_rejected()

	fmt.println()
	fmt.printf("=== Results: %d passed, %d failed ===\n", g_pass, g_fail)

	if g_fail > 0 {
		os.exit(1)
	}
}
