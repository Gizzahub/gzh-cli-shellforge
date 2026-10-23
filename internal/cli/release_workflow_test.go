package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Exercise the shipped command graph with actual isolated configuration files.
func TestReleaseWorkflow_BuildPreviewDeployBackupAndRestore(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	require.NoError(t, os.Mkdir(home, 0o700))
	t.Setenv("HOME", home)
	manifest := filepath.Join(root, "manifest.yaml")
	module := filepath.Join(root, "module.sh")
	build := filepath.Join(root, "build")
	backup := filepath.Join(root, "backups")
	require.NoError(t, os.WriteFile(module, []byte("export RELEASE_TEST=ready\n"), 0o600))
	require.NoError(t, os.WriteFile(manifest, []byte("shell:\n  type: zsh\nmodules:\n  - name: release-test\n    file: module.sh\n    target: zshrc\n"), 0o600))
	run := func(args ...string) {
		t.Helper()
		cmd := NewRootCmd()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute(), "args=%v output=%s", args, out.String())
	}
	run("build", "--manifest", manifest, "--config-dir", root, "--output-dir", build, "--verbose")
	generated, err := os.ReadFile(filepath.Join(build, ".zshrc"))
	require.NoError(t, err)
	require.Contains(t, string(generated), "export RELEASE_TEST=ready")
	run("build", "--manifest", manifest, "--config-dir", root, "--dry-run", "--verbose")
	run("validate", "--manifest", manifest, "--config-dir", root, "--verbose")
	run("doctor", "--manifest", manifest, "--verbose")
	run("list", "--manifest", manifest, "--config-dir", root, "--verbose")
	target := filepath.Join(home, ".zshrc")
	require.NoError(t, os.WriteFile(target, []byte("original\n"), 0o600))
	run("deploy", "--build-dir", build, "--dry-run", "--backup", "--verbose")
	original, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "original\n", string(original))
	entries, err := os.ReadDir(home)
	require.NoError(t, err)
	require.Len(t, entries, 1, "preview must not create backups")
	run("deploy", "--build-dir", build, "--backup", "--verbose")
	deployed, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, generated, deployed)
	backups, err := filepath.Glob(target + ".backup.*")
	require.NoError(t, err)
	require.Len(t, backups, 1)
	saved, err := os.ReadFile(backups[0])
	require.NoError(t, err)
	require.Equal(t, original, saved)
	for _, format := range []string{"summary", "unified", "context", "side-by-side"} {
		run("diff", target, backups[0], "--format", format, "--verbose")
	}
	run("diff", target, target, "--verbose")
	run("backup", "--file", target, "--backup-dir", backup, "--no-git", "--verbose")
	snapshots, err := os.ReadDir(filepath.Join(backup, "snapshots", "zshrc"))
	require.NoError(t, err)
	require.NotEmpty(t, snapshots)
	snapshot := snapshots[0].Name()
	run("restore", "--file", target, "--snapshot", snapshot, "--backup-dir", backup, "--no-git", "--dry-run", "--verbose")
	run("restore", "--file", target, "--snapshot", snapshot, "--backup-dir", backup, "--no-git", "--verbose")
	run("cleanup", "--file", target, "--backup-dir", backup, "--no-git", "--dry-run", "--verbose")
	run("cleanup", "--file", target, "--backup-dir", backup, "--no-git", "--verbose")
	run("template", "list")
	run("template", "generate", "env", "release-env", "--config-dir", root, "--field", "var_name=RELEASE_ENV", "--field", "var_value=ready", "--verbose")
}

func TestReleaseWorkflow_ProfilesQueries(t *testing.T) {
	flags := &profilesFlags{dataDir: "../../data/shell-profiles", verbose: true}
	for _, category := range []string{"", "distributions", "managers", "desktops", "modes", "multiplexers"} {
		require.NoError(t, runProfilesList(category, flags), category)
	}
	for _, item := range [][2]string{{"os", "ubuntu"}, {"os", "mac"}, {"manager", "nvm"}, {"manager", "rbenv"}, {"desktop", "gnome"}, {"desktop", "kde"}, {"mode", "login"}, {"mode", "interactive"}, {"multiplexer", "tmux"}, {"multiplexer", "screen"}, {"multiplexer", "zellij"}} {
		require.NoError(t, runProfilesShow(item[0], item[1], flags), "%v", item)
		require.Error(t, runProfilesShow(item[0], "does-not-exist", flags))
	}
	require.Error(t, runProfilesList("does-not-exist", flags))
	require.Error(t, runProfilesShow("does-not-exist", "name", flags))
	for _, context := range []string{"ssh", "cron", "container", "unknown"} {
		require.NoError(t, runProfilesCheck(context, flags))
	}
}
