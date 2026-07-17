package app_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// mockManager is a domain.PackageManager test double that never shells out.
type mockManager struct {
	name         string
	installed    map[string]bool
	installErr   error
	installCalls []string // packages Install() was actually called with
}

func (m *mockManager) Name() string { return m.name }

func (m *mockManager) IsInstalled(pkg string) (bool, error) {
	return m.installed[pkg], nil
}

func (m *mockManager) Install(pkg string) error {
	m.installCalls = append(m.installCalls, pkg)
	if m.installErr != nil {
		return m.installErr
	}
	if m.installed == nil {
		m.installed = map[string]bool{}
	}
	m.installed[pkg] = true
	return nil
}

func managers(ms ...*mockManager) map[string]domain.PackageManager {
	out := map[string]domain.PackageManager{}
	for _, m := range ms {
		out[m.name] = m
	}
	return out
}

func TestPrepareService_Plan_MissingPackage(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "setup-mise", File: "mise.sh", Packages: map[string][]string{"brew": {"mise"}}},
	}}
	brew := &mockManager{name: "brew", installed: map[string]bool{}}

	svc := app.NewPrepareService(managers(brew))
	result, err := svc.Plan(manifest, "Mac")

	require.NoError(t, err)
	require.Len(t, result.Packages, 1)
	assert.Equal(t, "mise", result.Packages[0].Package)
	assert.False(t, result.Packages[0].Installed)
	assert.True(t, result.Packages[0].Supported)
	assert.False(t, result.AllSatisfied())
	assert.Empty(t, brew.installCalls, "Plan must never install")
}

func TestPrepareService_Plan_AlreadyInstalled(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "setup-mise", File: "mise.sh", Packages: map[string][]string{"brew": {"mise"}}},
	}}
	brew := &mockManager{name: "brew", installed: map[string]bool{"mise": true}}

	svc := app.NewPrepareService(managers(brew))
	result, err := svc.Plan(manifest, "Mac")

	require.NoError(t, err)
	assert.True(t, result.AllSatisfied())
}

func TestPrepareService_Plan_UnsupportedManagerDoesNotBlockOtherOS(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "gcloud", File: "gcloud.sh", OS: []string{"Mac"}, Packages: map[string][]string{"cask": {"google-cloud-sdk"}}},
	}}

	// No managers registered at all (e.g. running on an unsupported OS).
	svc := app.NewPrepareService(managers())
	result, err := svc.Plan(manifest, "Mac")

	require.NoError(t, err)
	require.Len(t, result.Packages, 1)
	assert.False(t, result.Packages[0].Supported)
	// Unsupported packages don't block AllSatisfied — prepare only manages
	// what it has a manager for.
	assert.True(t, result.AllSatisfied())
}

func TestPrepareService_Plan_FiltersByOS(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "mac-only", File: "mac.sh", OS: []string{"Mac"}, Packages: map[string][]string{"cask": {"google-cloud-sdk"}}},
		{Name: "linux-only", File: "linux.sh", OS: []string{"Linux"}, Packages: map[string][]string{"apt": {"chezmoi"}}},
	}}
	apt := &mockManager{name: "apt", installed: map[string]bool{}}

	svc := app.NewPrepareService(managers(apt))
	result, err := svc.Plan(manifest, "Linux")

	require.NoError(t, err)
	require.Len(t, result.Packages, 1)
	assert.Equal(t, "chezmoi", result.Packages[0].Package)
}

func TestPrepareService_Plan_InvalidManifestFailsBeforeInspectingPackages(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "dup", File: "a.sh"},
		{Name: "dup", File: "b.sh"},
	}}
	brew := &mockManager{name: "brew"}

	svc := app.NewPrepareService(managers(brew))
	_, err := svc.Plan(manifest, "Mac")

	require.Error(t, err)
	assert.Empty(t, brew.installCalls)
}

func TestPrepareService_Apply_SkipsAlreadyInstalled_Idempotent(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "setup-mise", File: "mise.sh", Packages: map[string][]string{"brew": {"mise"}}},
	}}
	brew := &mockManager{name: "brew", installed: map[string]bool{"mise": true}}

	svc := app.NewPrepareService(managers(brew))
	result, err := svc.Apply(manifest, "Mac")

	require.NoError(t, err)
	assert.Empty(t, brew.installCalls, "already installed package must not be reinstalled")
	assert.Empty(t, result.Installed)
	assert.Empty(t, result.Failed)
}

func TestPrepareService_Apply_InstallsMissingAndReVerifies(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "setup-mise", File: "mise.sh", Packages: map[string][]string{"brew": {"mise"}}},
	}}
	brew := &mockManager{name: "brew", installed: map[string]bool{}}

	svc := app.NewPrepareService(managers(brew))
	result, err := svc.Apply(manifest, "Mac")

	require.NoError(t, err)
	assert.Equal(t, []string{"mise"}, brew.installCalls)
	require.Len(t, result.Installed, 1)
	assert.Equal(t, "mise", result.Installed[0].Package)
	assert.True(t, result.AllSatisfied(), "post-install re-verify should show it installed")
}

func TestPrepareService_Apply_InstallFailurePropagates(t *testing.T) {
	manifest := &domain.Manifest{Modules: []domain.Module{
		{Name: "setup-mise", File: "mise.sh", Packages: map[string][]string{"brew": {"mise"}}},
	}}
	brew := &mockManager{name: "brew", installed: map[string]bool{}, installErr: errors.New("network unreachable")}

	svc := app.NewPrepareService(managers(brew))
	result, err := svc.Apply(manifest, "Mac")

	require.NoError(t, err, "install failure is reported via result.Failed, not a returned error")
	require.Len(t, result.Failed, 1)
	assert.Equal(t, "mise", result.Failed[0].Package)
	assert.False(t, result.AllSatisfied())
}
