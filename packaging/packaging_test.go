package packaging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"koji/internal/config"
)

func TestPackagingLayoutFilesExist(t *testing.T) {
	for _, path := range requiredPackagingFiles() {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected packaging file %s: %v", path, err)
		}
	}
}

func TestSystemdUnitsUseKojiServiceUser(t *testing.T) {
	for _, path := range systemdUnitFiles() {
		content := readPackagingFile(t, path)
		if !strings.Contains(content, "User=koji") || !strings.Contains(content, "Group=koji") {
			t.Fatalf("expected koji service user in %s", path)
		}
		if strings.Contains(content, "User=root") {
			t.Fatalf("unexpected root runtime user in %s", path)
		}
	}
}

func TestSystemdUnitsUseProductionRuntimePaths(t *testing.T) {
	daemon := readPackagingFile(t, filepath.Join("systemd", "kojid.service"))
	agent := readPackagingFile(t, filepath.Join("systemd", "koji-agent.service"))

	assertContainsAll(t, daemon, []string{
		"ExecStart=/usr/bin/kojid",
		"WorkingDirectory=/",
		"ReadWritePaths=/var/lib/koji",
		"RuntimeDirectory=koji",
	})
	assertContainsAll(t, agent, []string{
		"ExecStart=/usr/bin/koji-agent -config /etc/koji/agent.yaml",
		"WorkingDirectory=/",
		"ReadWritePaths=/run/koji",
		"RuntimeDirectory=koji",
	})
	assertContainsNone(t, daemon+agent, localPathFragments())
}

func TestPackagingExamplesAreValid(t *testing.T) {
	if _, err := config.Load(filepath.Join("examples", "koji.yaml")); err != nil {
		t.Fatalf("load daemon example: %v", err)
	}
	if _, err := config.LoadAgent(filepath.Join("examples", "agent.yaml")); err != nil {
		t.Fatalf("load agent example: %v", err)
	}
}

func TestPackagingExamplesContainNoLocalDeveloperPaths(t *testing.T) {
	for _, path := range exampleConfigFiles() {
		assertContainsNone(t, readPackagingFile(t, path), localPathFragments())
	}
}

func TestMakefileSeparatesStagingFromRuntimePaths(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	makefile := string(content)
	assertContainsAll(t, makefile, []string{
		"DESTDIR ?= build/rootfs",
		"$(DESTDIR)$(PREFIX)/bin",
		"$(DESTDIR)$(PREFIX)/share/koji",
		"$(DESTDIR)$(SYSCONFDIR)/koji",
		"$(DESTDIR)$(LOCALSTATEDIR)/koji",
		"$(DESTDIR)$(SYSTEMDUNITDIR)",
	})
	assertContainsNone(t, makefile, []string{
		"/Us" + "ers",
		"Documents" + "/Projects",
		"god" + "zilla",
	})
}

func TestMakefileDefinesReleaseTargets(t *testing.T) {
	makefile := readRepoFile(t, "Makefile")
	assertContainsAll(t, makefile, []string{
		"release:",
		"checksums:",
		"verify-release:",
		"koji-rootfs-$(GOOS)-$(GOARCH).tar.gz",
	})
	assertContainsAll(t, readPackagingFile(t, filepath.Join("scripts", "checksums.sh")), []string{
		"SHA256SUMS.txt",
	})
}

func TestReleaseWorkflowUsesPinnedToolchains(t *testing.T) {
	workflow := readRepoFile(t, filepath.Join(".github", "workflows", "release.yml"))
	assertContainsAll(t, workflow, []string{
		`- "v*"`,
		"go-version: \"1.25.0\"",
		"node-version: \"22.12.0\"",
		"npm ci",
		"go test ./...",
		"make release",
		"make verify-release",
		"actions/upload-artifact@v4",
		"actions/download-artifact@v4",
		"softprops/action-gh-release@v2",
	})
	assertContainsNone(t, workflow, []string{
		"go-version: latest",
		"node-version: latest",
		"@master",
		"@main",
		"@latest",
	})
}

func TestReleaseWorkflowHasSmokeGateBeforePublish(t *testing.T) {
	workflow := readRepoFile(t, filepath.Join(".github", "workflows", "release.yml"))
	assertContainsAll(t, workflow, []string{
		"build_release:",
		"smoke_test_release:",
		"publish_release:",
		"needs: build_release",
		"needs: smoke_test_release",
		"checksums_valid: ${{ steps.smoke.outputs.checksums_valid }}",
		"rootfs_layout_valid: ${{ steps.smoke.outputs.rootfs_layout_valid }}",
		"systemd_units_valid: ${{ steps.smoke.outputs.systemd_units_valid }}",
		"forbidden_paths_found: ${{ steps.smoke.outputs.forbidden_paths_found }}",
		"packaging/scripts/ci_verify_release_outputs.sh build/release",
	})
}

func TestReleaseScriptsValidateChecksumsAndForbiddenPaths(t *testing.T) {
	checksums := readPackagingFile(t, filepath.Join("scripts", "checksums.sh"))
	verification := readPackagingFile(t, filepath.Join("scripts", "verify_release.sh"))

	assertContainsAll(t, checksums, []string{
		"SHA256SUMS.txt",
		"kojid-linux-amd64",
		"koji-agent-linux-amd64",
		"koji-rootfs-linux-amd64.tar.gz",
	})
	assertContainsAll(t, verification, []string{
		"/Users/",
		"/home/",
		"Documents/Projects",
		"usr/share/koji/dist",
		"usr/lib/systemd/system",
		"SHA256SUMS.txt",
	})
}

func TestReleaseSmokeScriptsValidateDownloadedArtifacts(t *testing.T) {
	outputs := readPackagingFile(t, filepath.Join("scripts", "ci_verify_release_outputs.sh"))
	checksums := readPackagingFile(t, filepath.Join("scripts", "ci_verify_checksums.sh"))
	rootfs := readPackagingFile(t, filepath.Join("scripts", "ci_verify_rootfs_layout.sh"))
	systemd := readPackagingFile(t, filepath.Join("scripts", "ci_verify_systemd_units.sh"))

	assertContainsAll(t, outputs, []string{
		"checksums_valid=",
		"rootfs_layout_valid=",
		"systemd_units_valid=",
		"forbidden_paths_found=",
		"Artifact Smoke Test Summary",
		"kojid-linux-amd64",
		"koji-agent-linux-amd64",
		"koji-rootfs-linux-amd64.tar.gz",
		"SHA256SUMS.txt",
		"--help",
	})
	assertContainsAll(t, checksums, []string{
		"missing or empty checksum file",
		"missing checksum entry",
		"sha256sum -c SHA256SUMS.txt",
	})
	assertContainsAll(t, rootfs, []string{
		"usr/bin/kojid",
		"usr/bin/koji-agent",
		"usr/share/koji/dist",
		"etc/koji",
		"usr/lib/systemd/system",
		"var/lib/koji",
	})
	assertContainsAll(t, systemd, []string{
		"ExecStart=/usr/bin/kojid",
		"ExecStart=/usr/bin/koji-agent -config /etc/koji/agent.yaml",
		"RuntimeDirectory=koji",
		"WorkingDirectory=/",
	})
}

func requiredPackagingFiles() []string {
	return []string{
		filepath.Join("systemd", "kojid.service"),
		filepath.Join("systemd", "koji-agent.service"),
		filepath.Join("examples", "koji.yaml"),
		filepath.Join("examples", "agent.yaml"),
		"install.sh",
	}
}

func exampleConfigFiles() []string {
	return []string{
		filepath.Join("examples", "koji.yaml"),
		filepath.Join("examples", "agent.yaml"),
	}
}

func systemdUnitFiles() []string {
	return []string{
		filepath.Join("systemd", "kojid.service"),
		filepath.Join("systemd", "koji-agent.service"),
	}
}

func localPathFragments() []string {
	return []string{
		"/Us" + "ers",
		"Documents" + "/Projects",
		"god" + "zilla",
		"localhost" + ":5173",
		"zu" + "ki",
	}
}

func assertContainsAll(t *testing.T, content string, expected []string) {
	t.Helper()

	for _, value := range expected {
		if !strings.Contains(content, value) {
			t.Fatalf("expected %q in content", value)
		}
	}
}

func assertContainsNone(t *testing.T, content string, forbidden []string) {
	t.Helper()

	for _, value := range forbidden {
		if strings.Contains(content, value) {
			t.Fatalf("unexpected %q in content", value)
		}
	}
}

func readPackagingFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func readRepoFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
