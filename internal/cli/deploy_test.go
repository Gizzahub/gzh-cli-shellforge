package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestNewDeployCmd(t *testing.T) {
	cmd := newDeployCmd()

	if cmd.Use != "deploy" {
		t.Errorf("Use = %q, want %q", cmd.Use, "deploy")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify flags exist
	flags := []string{"build-dir", "dry-run", "backup", "verbose"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Flag %q not found", flag)
		}
	}
}

func TestNewDeployCmd_DefaultBuildDir(t *testing.T) {
	cmd := newDeployCmd()

	flag := cmd.Flags().Lookup("build-dir")
	if flag == nil {
		t.Fatal("build-dir flag not found")
	}

	if flag.DefValue != "./build" {
		t.Errorf("build-dir default = %q, want %q", flag.DefValue, "./build")
	}
}

func TestNewDeployCmd_ShortFlags(t *testing.T) {
	cmd := newDeployCmd()

	// Check short flag for build-dir
	flag := cmd.Flags().ShorthandLookup("d")
	if flag == nil {
		t.Error("Short flag -d not found")
	}

	// Check short flag for verbose
	flag = cmd.Flags().ShorthandLookup("v")
	if flag == nil {
		t.Error("Short flag -v not found")
	}
}

func TestDeployCmd_RequiresBackupOrDryRun(t *testing.T) {
	cmd := newDeployCmd()
	cmd.SetArgs(nil)
	require.ErrorContains(t, cmd.Execute(), "--backup or --dry-run")
}

func TestDeployCmd_TargetFailureReturnsError(t *testing.T) {
	buildDir := t.TempDir()
	metadata := domain.BuildMetadata{Files: []domain.BuildFileInfo{{Source: "missing", DestPath: ".zshrc"}}}
	data, err := metadata.ToJSON()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, domain.MetadataFileName), data, 0o600))
	cmd := newDeployCmd()
	cmd.SetArgs([]string{"--backup", "--build-dir", buildDir})
	require.ErrorContains(t, cmd.Execute(), "deployment failed for 1 file(s)")
}

func TestDeployCmd_BackupFailureReturnsErrorAndPreservesTarget(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	target := filepath.Join(home, ".zshrc")
	require.NoError(t, os.MkdirAll(target, 0o700))
	sentinel := filepath.Join(target, "keep")
	require.NoError(t, os.WriteFile(sentinel, []byte("original"), 0o600))
	t.Setenv("HOME", home)
	buildDir := filepath.Join(root, "build")
	require.NoError(t, os.Mkdir(buildDir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, "config"), []byte("replacement"), 0o600))
	metadata := domain.BuildMetadata{Files: []domain.BuildFileInfo{{Source: "config", DestPath: ".zshrc"}}}
	data, err := metadata.ToJSON()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(buildDir, domain.MetadataFileName), data, 0o600))
	cmd := newDeployCmd()
	cmd.SetArgs([]string{"--backup", "--build-dir", buildDir})
	require.ErrorContains(t, cmd.Execute(), "deployment failed for 1 file(s)")
	preserved, err := os.ReadFile(sentinel)
	require.NoError(t, err)
	require.Equal(t, "original", string(preserved))
	entries, err := os.ReadDir(home)
	require.NoError(t, err)
	require.Len(t, entries, 1, "failed directory backup must not create a backup or overwrite destination")
}
