package packaging

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"koji/internal/config"
	"koji/internal/db"
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

func TestMakefileDefinesBackupRestoreTargets(t *testing.T) {
	makefile := readRepoFile(t, "Makefile")
	assertContainsAll(t, makefile, []string{
		"backup:",
		"restore:",
		"verify-restore:",
		"packaging/scripts/backup.sh",
		"packaging/scripts/restore.sh $(BACKUP)",
		"packaging/scripts/verify_restore.sh",
	})
}

func TestBackupRestoreScriptsRecoverGovernanceData(t *testing.T) {
	requireSQLite(t)

	root := t.TempDir()
	dbPath := filepath.Join(root, "var", "lib", "koji", "koji.db")
	configDir := filepath.Join(root, "etc", "koji")
	backupRoot := filepath.Join(root, "backups")

	createRecoveryFixture(t, dbPath, configDir)
	archivePath := runBackupScript(t, dbPath, configDir, backupRoot)
	removeRuntimeState(t, dbPath, configDir)
	runRestoreScript(t, archivePath, dbPath, configDir)
	assertRestoredGovernanceData(t, dbPath, configDir)
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
		filepath.Join("scripts", "backup.sh"),
		filepath.Join("scripts", "restore.sh"),
		filepath.Join("scripts", "verify_restore.sh"),
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

func requireSQLite(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Skip("sqlite3 is required for backup and restore scripts")
	}
}

func createRecoveryFixture(t *testing.T, dbPath string, configDir string) {
	t.Helper()

	conn, err := db.Open(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("open fixture database: %v", err)
	}
	defer conn.Close()

	seedRecoveryData(t, conn)
	if err := os.MkdirAll(configDir, 0750); err != nil {
		t.Fatalf("create config dir: %v", err)
	}
	copyFile(t, filepath.Join("examples", "koji.yaml"), filepath.Join(configDir, "koji.yaml"))
	copyFile(t, filepath.Join("examples", "agent.yaml"), filepath.Join(configDir, "agent.yaml"))
}

func seedRecoveryData(t *testing.T, conn *sql.DB) {
	t.Helper()

	statements := []string{
		"INSERT INTO users (id, username, password_hash) VALUES (1, 'operator', 'hash')",
		"INSERT INTO user_capabilities (user_id, capability_name) VALUES (1, 'jobs.read')",
		"INSERT INTO jobs (id, created_by, action, target, status, approved_by, approved_at, decision_reason) VALUES ('job-1', 1, 'restart', 'kojid.service', 'approved', 1, '2026-01-01T00:00:00Z', 'maintenance')",
		"INSERT INTO audit_events (actor, action, target, status, message, user_id, outcome, reason_code, request_id) VALUES ('operator', 'job.approved', 'jobs:job-1', 'ok', 'approved', 1, 'success', 'approved', 'req-1')",
	}
	for _, statement := range statements {
		if _, err := conn.Exec(statement); err != nil {
			t.Fatalf("seed recovery data: %v", err)
		}
	}
}

func runBackupScript(t *testing.T, dbPath string, configDir string, backupRoot string) string {
	t.Helper()

	output := runPackagingCommand(t, []string{
		"KOJI_DB_PATH=" + dbPath,
		"KOJI_CONFIG_DIR=" + configDir,
		"KOJI_VERSION=test",
	}, "./scripts/backup.sh", backupRoot)
	archivePath := strings.TrimSpace(output)
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("expected backup archive %s: %v", archivePath, err)
	}
	return archivePath
}

func runRestoreScript(t *testing.T, archivePath string, dbPath string, configDir string) {
	t.Helper()

	runPackagingCommand(t, []string{
		"KOJI_DB_PATH=" + dbPath,
		"KOJI_CONFIG_DIR=" + configDir,
	}, "./scripts/restore.sh", archivePath)
}

func assertRestoredGovernanceData(t *testing.T, dbPath string, configDir string) {
	t.Helper()

	runPackagingCommand(t, nil, "./scripts/verify_restore.sh", dbPath)
	assertRestoredCount(t, dbPath, "users")
	assertRestoredCount(t, dbPath, "user_capabilities")
	assertRestoredCount(t, dbPath, "jobs")
	assertRestoredCount(t, dbPath, "audit_events")
	if _, err := os.Stat(filepath.Join(configDir, "koji.yaml")); err != nil {
		t.Fatalf("expected restored daemon config: %v", err)
	}
	if _, err := os.Stat(filepath.Join(configDir, "agent.yaml")); err != nil {
		t.Fatalf("expected restored agent config: %v", err)
	}
}

func assertRestoredCount(t *testing.T, dbPath string, table string) {
	t.Helper()

	output := runPackagingCommand(t, nil, "sqlite3", dbPath, "SELECT COUNT(*) FROM "+table+";")
	if strings.TrimSpace(output) != "1" {
		t.Fatalf("expected one restored %s row, got %q", table, output)
	}
}

func removeRuntimeState(t *testing.T, dbPath string, configDir string) {
	t.Helper()

	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("remove runtime database: %v", err)
	}
	if err := os.RemoveAll(configDir); err != nil {
		t.Fatalf("remove runtime config: %v", err)
	}
}

func copyFile(t *testing.T, source string, target string) {
	t.Helper()

	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("read %s: %v", source, err)
	}
	if err := os.WriteFile(target, content, 0640); err != nil {
		t.Fatalf("write %s: %v", target, err)
	}
}

func runPackagingCommand(t *testing.T, env []string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", name, err, output)
	}
	return string(output)
}

func readRepoFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join("..", path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}
