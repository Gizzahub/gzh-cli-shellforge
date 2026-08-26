package domain

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeRunner lets tests control command output without invoking brew/apt.
type fakeRunner struct {
	// ok maps "name arg0 arg1..." to whether the command should succeed.
	ok map[string]bool
	// calls records every command invocation for assertions.
	calls [][]string
	ctxs  []context.Context
}

func (f *fakeRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	f.calls = append(f.calls, call)
	f.ctxs = append(f.ctxs, ctx)
	if f.ok[joinCall(call)] {
		return []byte("ok"), nil
	}
	return []byte(""), errors.New("not found")
}

func joinCall(call []string) string {
	return strings.Join(call, " ")
}

func TestBrewFormulaManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"brew list --formula --versions mise": true,
		"brew install missing-pkg":            true,
	}}
	m := &BrewFormulaManager{Runner: runner}

	assert.Equal(t, "brew", m.Name())

	installed, err := m.IsInstalled(context.Background(), "mise")
	require.NoError(t, err)
	assert.True(t, installed)

	installed, err = m.IsInstalled(context.Background(), "missing-pkg")
	require.NoError(t, err)
	assert.False(t, installed)

	require.NoError(t, m.Install(context.Background(), "missing-pkg"))
}

func TestBrewFormulaManager_InstallFailure(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{}}
	m := &BrewFormulaManager{Runner: runner}

	err := m.Install(context.Background(), "broken-pkg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broken-pkg")
}

func TestBrewCaskManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"brew list --cask --versions google-cloud-sdk": true,
	}}
	m := &BrewCaskManager{Runner: runner}

	assert.Equal(t, "cask", m.Name())

	installed, err := m.IsInstalled(context.Background(), "google-cloud-sdk")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestAptManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"dpkg-query -W -f=${Status} chezmoi": true,
	}}
	m := &AptManager{Runner: runner}

	assert.Equal(t, "apt", m.Name())

	installed, err := m.IsInstalled(context.Background(), "chezmoi")
	require.NoError(t, err)
	// fakeRunner returns "ok" as output, which does not contain
	// "install ok installed", so IsInstalled must report false here —
	// this exercises the output-parsing branch, not just the error branch.
	assert.False(t, installed)
}

func TestAptManager_IsInstalled_ParsesDpkgStatus(t *testing.T) {
	runner := &dpkgStatusRunner{status: "install ok installed"}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "chezmoi")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestAptManager_IsInstalled_UnknownPackageIsMissingNotError(t *testing.T) {
	runner := &commandResultRunner{err: commandExitError{code: 1}}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "does-not-exist")
	require.NoError(t, err)
	assert.False(t, installed)
}

func TestAptManager_IsInstalled_FatalDpkgErrorIsReported(t *testing.T) {
	cause := commandExitError{code: 2}
	runner := &commandResultRunner{output: []byte("database is locked"), err: cause}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "chezmoi")

	assert.False(t, installed)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "chezmoi")
	assert.Contains(t, err.Error(), "database is locked")
}

func TestAptManager_IsInstalled_CommandFailureIsReported(t *testing.T) {
	cause := errors.New("dpkg-query executable not found")
	runner := &commandResultRunner{err: cause}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled(context.Background(), "chezmoi")

	assert.False(t, installed)
	require.Error(t, err)
	assert.ErrorIs(t, err, cause)
	assert.Contains(t, err.Error(), "chezmoi")
}

// dpkgStatusRunner returns a fixed dpkg-query style status line regardless of args.
type dpkgStatusRunner struct{ status string }

func (r *dpkgStatusRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return []byte(r.status), nil
}

type commandResultRunner struct {
	output []byte
	err    error
}

func (r *commandResultRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return r.output, r.err
}

type commandExitError struct{ code int }

func (e commandExitError) Error() string { return "command failed" }
func (e commandExitError) ExitCode() int { return e.code }

func TestBrewFormulaManager_ForwardsContextToRunner(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{"brew list --formula --versions mise": true}}
	manager := &BrewFormulaManager{Runner: runner}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	installed, err := manager.IsInstalled(ctx, "mise")

	require.NoError(t, err)
	assert.True(t, installed)
	require.Len(t, runner.ctxs, 1)
	assert.Same(t, ctx, runner.ctxs[0])
}

func TestBrewManagers_PropagateContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		manager PackageManager
		wantErr error
	}{
		{name: "formula", manager: &BrewFormulaManager{Runner: &commandResultRunner{err: context.Canceled}}, wantErr: context.Canceled},
		{name: "cask", manager: &BrewCaskManager{Runner: &commandResultRunner{err: context.DeadlineExceeded}}, wantErr: context.DeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installed, err := tt.manager.IsInstalled(context.Background(), "pkg")

			assert.False(t, installed)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
