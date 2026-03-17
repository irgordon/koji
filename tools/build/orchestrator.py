"""
KOJI Phase 0.2 — Build Orchestrator

Validates isolated build environments for kernel (Odin) and userspace (Go).
Enforces cross-boundary separation rules.

Must be run from project root inside .venv.
Must not create, modify, or delete any files.
"""

import os
import sys
import subprocess


PROJECT_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

ODIN = "/opt/homebrew/bin/odin"
KERNEL_FILE = os.path.join(PROJECT_ROOT, "kernel", "main.odin")
USERSPACE_FILE = os.path.join(PROJECT_ROOT, "userspace", "build_check.go")

KERNEL_DIR = os.path.join(PROJECT_ROOT, "kernel")
USERSPACE_DIR = os.path.join(PROJECT_ROOT, "userspace")


def check_venv():
    venv_path = os.path.join(PROJECT_ROOT, ".venv")
    if not sys.prefix.startswith(venv_path):
        print("FAIL: not running inside .venv at project root")
        print(f"  sys.prefix: {sys.prefix}")
        print(f"  expected prefix: {venv_path}")
        return False
    print("PASS: running inside .venv")
    return True


def check_pyyaml():
    try:
        import yaml  # noqa: F401
        print("PASS: PyYAML importable")
        return True
    except ImportError:
        print("FAIL: PyYAML not importable")
        return False


def build_kernel():
    result = subprocess.run(
        [ODIN, "build", KERNEL_FILE, "-file"],
        capture_output=True,
        text=True,
        cwd=PROJECT_ROOT,
    )
    if result.returncode == 0:
        print("PASS: kernel/main.odin compiles successfully")
        return True
    print("FAIL: kernel/main.odin compilation failed")
    print(f"  stdout: {result.stdout}")
    print(f"  stderr: {result.stderr}")
    return False


def build_userspace():
    result = subprocess.run(
        ["go", "build", USERSPACE_FILE],
        capture_output=True,
        text=True,
        cwd=PROJECT_ROOT,
    )
    if result.returncode == 0:
        print("PASS: userspace/build_check.go compiles successfully")
        return True
    print("FAIL: userspace/build_check.go compilation failed")
    print(f"  stdout: {result.stdout}")
    print(f"  stderr: {result.stderr}")
    return False


def check_boundary_kernel():
    violations = []
    for name in os.listdir(KERNEL_DIR):
        path = os.path.join(KERNEL_DIR, name)
        if not os.path.isfile(path):
            continue
        with open(path, "r") as f:
            content = f.read()
        if "../userspace" in content:
            violations.append(name)
    if violations:
        print(f"FAIL: kernel/ contains '../userspace' reference in: {violations}")
        return False
    print("PASS: kernel/ has no cross-boundary references to userspace")
    return True


def check_boundary_userspace():
    violations = []
    for name in os.listdir(USERSPACE_DIR):
        path = os.path.join(USERSPACE_DIR, name)
        if not os.path.isfile(path):
            continue
        with open(path, "r") as f:
            content = f.read()
        if "../kernel" in content:
            violations.append(name)
    if violations:
        print(f"FAIL: userspace/ contains '../kernel' reference in: {violations}")
        return False
    print("PASS: userspace/ has no cross-boundary references to kernel")
    return True


results = {}

print("=== KOJI Build Orchestrator — Phase 0.2 ===")
print()

print("--- Environment Validation ---")
results["venv_active"] = check_venv()
results["pyyaml_available"] = check_pyyaml()
print()

if not results["venv_active"] or not results["pyyaml_available"]:
    print("ABORT: environment validation failed")
    sys.exit(1)

print("--- Build Validation ---")
results["kernel_build"] = build_kernel()
results["userspace_build"] = build_userspace()
print()

print("--- Boundary Enforcement ---")
results["kernel_boundary"] = check_boundary_kernel()
results["userspace_boundary"] = check_boundary_userspace()
print()

print("--- Summary ---")
all_passed = all(results.values())
for key, value in results.items():
    status = "PASS" if value else "FAIL"
    print(f"  {key}: {status}")
print()

if all_passed:
    print("RESULT: ALL CHECKS PASSED")
    sys.exit(0)
else:
    print("RESULT: ONE OR MORE CHECKS FAILED")
    sys.exit(1)
