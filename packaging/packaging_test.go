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
