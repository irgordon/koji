VERIFY_DIR := verify

.PHONY: verify_env fix_permissions

verify_env: fix_permissions
	@echo "--- Phase 0: Verifying Toolchain ---"
	@$(VERIFY_DIR)/env_odin.sh
	@$(VERIFY_DIR)/env_nasm.sh
	@$(VERIFY_DIR)/env_go.sh
	@$(VERIFY_DIR)/env_qemu.sh
	@echo "--- Phase 0: Success ---"

fix_permissions:
	@chmod +x $(VERIFY_DIR)/*.sh
