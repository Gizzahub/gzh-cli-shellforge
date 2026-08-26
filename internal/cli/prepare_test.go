package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// fakePackageManager is a domain.PackageManager test double so prepare's CLI
// wiring can be exercised without shelling out to brew/apt.
type fakePackageManager struct {
	name         string
	installed    map[string]bool
	isInstalled  func(string) (bool, error)
	installErr   error
	installCalls []string
}

func (f *fakePackageManager) Name() string { return f.name }

func (f *fakePackageManager) IsInstalled(pkg string) (bool, error) {
	if f.isInstalled != nil {
		return f.isInstalled(pkg)
	}
	return f.installed[pkg], nil
}

func (f *fakePackageManager) Install(pkg string) error {
	f.installCalls = append(f.installCalls, pkg)
	if f.installErr != nil {
		return f.installErr
	}
	if f.installed == nil {
		f.installed = map[string]bool{}
	}
	f.installed[pkg] = true
	return nil
}

// writeTempManifest writes a single-module manifest declaring pkg under
// manager, and returns its path. The module has no os: restriction so it
// applies regardless of the test's targetOS.
func writeTempManifest(t *testing.T, manager, pkg string) string {
	t.Helper()
	content := "modules:\n" +
		"  - name: setup-mise\n" +
		"    file: init/56-setup-mise.sh\n" +
		"    packages:\n" +
		"      " + manager + ": [" + pkg + "]\n"

	path := filepath.Join(t.TempDir(), "manifest.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func TestPrepareCmd_Flags(t *testing.T) {
	cmd := newPrepareCmd()

	checkFlag := cmd.Flags().Lookup("check")
	require.NotNil(t, checkFlag)
	assert.Equal(t, "false", checkFlag.DefValue)

	dryRunFlag := cmd.Flags().Lookup("dry-run")
	require.NotNil(t, dryRunFlag)
	assert.Equal(t, "false", dryRunFlag.DefValue)

	manifestFlag := cmd.Flags().Lookup("manifest")
	require.NotNil(t, manifestFlag)
	assert.Equal(t, "manifest.yaml", manifestFlag.DefValue)
}

func TestPrepareCmd_Help(t *testing.T) {
	cmd := newPrepareCmd()

	assert.Equal(t, "prepare", cmd.Use)
	assert.Contains(t, cmd.Long, "--check")
	assert.Contains(t, cmd.Long, "--dry-run")
}

func TestPrepareCheck_DoesNotInstall(t *testing.T) {
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{"mise": true}}
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac", check: true}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.NoError(t, err, "already installed, so check should pass")
	assert.Empty(t, fake.installCalls, "check must never call Install")
}

func TestPrepareCheck_ReportsMissingAsError(t *testing.T) {
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{}}
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac", check: true}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.Error(t, err, "missing package should surface as a non-nil error (non-zero exit)")
	assert.Empty(t, fake.installCalls, "check must never call Install")
}

func TestPrepareDryRun_DoesNotInstall(t *testing.T) {
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{}} // missing
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac", dryRun: true}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.NoError(t, err, "dry-run reports the plan but always exits 0")
	assert.Empty(t, fake.installCalls, "dry-run must never call Install")
}

func TestPrepareApply_SkipsAlreadyInstalled_Idempotent(t *testing.T) {
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{"mise": true}}
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac"}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.NoError(t, err)
	assert.Empty(t, fake.installCalls, "already installed package must not be reinstalled")
}

func TestPrepareApply_InstallsMissingPackage(t *testing.T) {
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{}}
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac"}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.NoError(t, err)
	assert.Equal(t, []string{"mise"}, fake.installCalls)
}

func TestPrepareApply_InstallFailurePropagatesNonZero(t *testing.T) {
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{}, installErr: errors.New("boom")}
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac"}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.Error(t, err, "install failure must propagate as a non-nil error (non-zero exit)")
}

func TestPrepareApply_TransientDetectionFailurePropagatesNonZero(t *testing.T) {
	detectionErr := errors.New("temporary package status failure")
	checks := 0
	fake := &fakePackageManager{name: "brew", installed: map[string]bool{}}
	fake.isInstalled = func(string) (bool, error) {
		checks++
		if checks == 1 {
			return false, detectionErr
		}
		return false, nil
	}
	flags := &prepareFlags{manifest: writeTempManifest(t, "brew", "mise"), targetOS: "Mac"}

	err := runPrepare(flags, map[string]domain.PackageManager{"brew": fake})

	require.Error(t, err, "an initial detection failure must remain a non-zero apply result")
	assert.Equal(t, 2, checks)
	assert.Empty(t, fake.installCalls, "detection failure must not trigger installation")
}

func TestDefaultPackageManagers(t *testing.T) {
	mac := defaultPackageManagers("Mac")
	assert.Contains(t, mac, "brew")
	assert.Contains(t, mac, "cask")

	linux := defaultPackageManagers("Linux")
	assert.Contains(t, linux, "apt")

	other := defaultPackageManagers("Windows")
	assert.Empty(t, other)
}
