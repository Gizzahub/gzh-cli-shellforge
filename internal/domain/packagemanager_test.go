package domain

import (
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
}

func (f *fakeRunner) Run(name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	f.calls = append(f.calls, call)
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

	installed, err := m.IsInstalled("mise")
	require.NoError(t, err)
	assert.True(t, installed)

	installed, err = m.IsInstalled("missing-pkg")
	require.NoError(t, err)
	assert.False(t, installed)

	require.NoError(t, m.Install("missing-pkg"))
}

func TestBrewFormulaManager_InstallFailure(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{}}
	m := &BrewFormulaManager{Runner: runner}

	err := m.Install("broken-pkg")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broken-pkg")
}

func TestBrewCaskManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"brew list --cask --versions google-cloud-sdk": true,
	}}
	m := &BrewCaskManager{Runner: runner}

	assert.Equal(t, "cask", m.Name())

	installed, err := m.IsInstalled("google-cloud-sdk")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestAptManager(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{
		"dpkg-query -W -f=${Status} chezmoi": true,
	}}
	m := &AptManager{Runner: runner}

	assert.Equal(t, "apt", m.Name())

	installed, err := m.IsInstalled("chezmoi")
	require.NoError(t, err)
	// fakeRunner returns "ok" as output, which does not contain
	// "install ok installed", so IsInstalled must report false here —
	// this exercises the output-parsing branch, not just the error branch.
	assert.False(t, installed)
}

func TestAptManager_IsInstalled_ParsesDpkgStatus(t *testing.T) {
	runner := &dpkgStatusRunner{status: "install ok installed"}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled("chezmoi")
	require.NoError(t, err)
	assert.True(t, installed)
}

func TestAptManager_IsInstalled_UnknownPackageIsMissingNotError(t *testing.T) {
	runner := &fakeRunner{ok: map[string]bool{}}
	m := &AptManager{Runner: runner}

	installed, err := m.IsInstalled("does-not-exist")
	require.NoError(t, err)
	assert.False(t, installed)
}

// dpkgStatusRunner returns a fixed dpkg-query style status line regardless of args.
type dpkgStatusRunner struct{ status string }

func (r *dpkgStatusRunner) Run(name string, args ...string) ([]byte, error) {
	return []byte(r.status), nil
}
