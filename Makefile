# ============================================================
# KOJI Kernel — Build System (v1, single-target: amd64)
#
# Prerequisites:
#   - odin (in PATH)
#   - nasm
#   - ld (GNU or LLVM lld)
#
# Targets:
#   make            — build kernel ELF
#   make clean      — remove build artifacts
#   make iso        — build bootable ISO (requires xorriso + limine)
#   make test       — run Phase 1 CNode invariant tests (requires odin)
# ============================================================

ARCH       := amd64
BUILD_DIR  := build
KERNEL_ELF := $(BUILD_DIR)/koji.elf

# ---- Tools ----
ODIN       := odin
NASM       := nasm
LD         := ld

# ---- Odin flags ----
ODIN_FLAGS := -target:freestanding_amd64_sysv \
              -build-mode:obj \
              -no-crt \
              -disable-red-zone \
              -default-to-nil-allocator \
              -o:minimal

# ---- NASM flags ----
NASM_FLAGS := -f elf64 -g -F dwarf

# ---- Linker flags ----
LD_FLAGS   := -T kernel/arch/$(ARCH)/linker.ld \
              -nostdlib \
              -static \
              -z max-page-size=0x1000

# ---- Sources ----
ASM_SRCS   := kernel/arch/$(ARCH)/boot.asm \
              kernel/arch/$(ARCH)/syscall_entry.asm \
              kernel/arch/$(ARCH)/io.asm

ASM_OBJS   := $(patsubst kernel/arch/$(ARCH)/%.asm,$(BUILD_DIR)/arch/%.o,$(ASM_SRCS))

ODIN_OBJ   := $(BUILD_DIR)/kernel.o

# ============================================================
# Targets
# ============================================================

.PHONY: all clean iso test

all: $(KERNEL_ELF)

# ---- Assemble NASM sources ----
$(BUILD_DIR)/arch/%.o: kernel/arch/$(ARCH)/%.asm | $(BUILD_DIR)/arch
	$(NASM) $(NASM_FLAGS) -o $@ $<

# ---- Compile Odin kernel package ----
$(ODIN_OBJ): kernel/*.odin abi/generated/odin/*.odin | $(BUILD_DIR)
	$(ODIN) build kernel/ $(ODIN_FLAGS) -out:$@

# ---- Link ----
$(KERNEL_ELF): $(ASM_OBJS) $(ODIN_OBJ)
	$(LD) $(LD_FLAGS) -o $@ $(ASM_OBJS) $(ODIN_OBJ)
	@echo ""
	@echo "=== KOJI kernel linked: $@ ==="
	@size $@ 2>/dev/null || true

# ---- Directories ----
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/arch:
	mkdir -p $(BUILD_DIR)/arch

# ---- Clean ----
clean:
	rm -rf $(BUILD_DIR)

# ---- Test (Phase 1 CNode invariant tests) ----
# Runs a standalone native Odin test binary that validates CNode behavioral
# invariants without requiring the kernel's freestanding build environment.
test:
	@echo "=== Running KOJI Phase 1 CNode invariant tests ==="
	$(ODIN) run tests/cnode_test.odin -file
	@echo "=== Tests complete ==="

# ---- ISO (placeholder — requires Limine setup) ----
iso: $(KERNEL_ELF)
	@echo "ISO generation requires Limine bootloader setup."
	@echo "See: https://github.com/limine-bootloader/limine"
	@echo "Kernel ELF ready at $(KERNEL_ELF)"
