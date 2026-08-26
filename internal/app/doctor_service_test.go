package app_test

import (
	"slices"
	"sort"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// mockLookup lets tests control which binaries/paths "exist".
type mockLookup struct {
	bins  map[string]bool
	paths map[string]bool
}

func (m mockLookup) HasBinary(name string) bool  { return m.bins[name] }
func (m mockLookup) PathExists(path string) bool { return m.paths[path] }

func newMock(bins, paths []string) mockLookup {
	m := mockLookup{bins: map[string]bool{}, paths: map[string]bool{}}
	for _, b := range bins {
		m.bins[b] = true
	}
	for _, p := range paths {
		m.paths[p] = true
	}
	return m
}

func makeManifest(modules []domain.Module) *domain.Manifest {
	return &domain.Manifest{Modules: modules}
}

func TestDoctorService_AllPresent(t *testing.T) {
	manifest := makeManifest([]domain.Module{
		{Name: "starship", File: "rc-post/starship.sh", RequiresBin: []string{"starship"}, OS: []string{"Mac"}},
	})
	lookup := newMock([]string{"starship"}, nil)

	result := app.NewDoctorService().Check(manifest, "Mac", lookup)

	if !result.AllOK() {
		t.Fatalf("expected AllOK, got missing: %v", result.Missing)
	}
	if result.ModuleCount != 1 {
		t.Errorf("ModuleCount = %d, want 1", result.ModuleCount)
	}
}

func TestDoctorService_MissingBinary(t *testing.T) {
	manifest := makeManifest([]domain.Module{
		{Name: "starship", File: "rc-post/starship.sh", RequiresBin: []string{"starship", "zsh"}},
		{Name: "mise-env", File: "rc-post/mise.sh", RequiresBin: []string{"mise", "zsh"}},
	})
	// only "zsh" is present
	lookup := newMock([]string{"zsh"}, nil)

	result := app.NewDoctorService().Check(manifest, "Linux", lookup)

	if result.AllOK() {
		t.Fatal("expected missing deps, got AllOK")
	}

	names := depNames(result.Missing)
	if !contains(names, "starship") {
		t.Errorf("expected 'starship' in missing, got %v", names)
	}
	if !contains(names, "mise") {
		t.Errorf("expected 'mise' in missing, got %v", names)
	}
	if contains(names, "zsh") {
		t.Errorf("'zsh' is present but listed as missing")
	}
}

func TestDoctorService_MissingPath(t *testing.T) {
	manifest := makeManifest([]domain.Module{
		{Name: "oh-my-zsh", File: "zshrc/00-omz.sh", RequiresPath: []string{"$HOME/.oh-my-zsh"}},
		{Name: "theme", File: "zshrc/theme.sh", RequiresPath: []string{"$HOME/.config/theme"}},
	})
	lookup := newMock(nil, []string{domain.ExpandPath("$HOME/.config/theme")})

	result := app.NewDoctorService().Check(manifest, "Mac", lookup)

	if result.AllOK() {
		t.Fatal("expected missing path dep")
	}
	if len(result.Missing) != 1 || result.Missing[0].Kind != "path" {
		t.Errorf("unexpected missing: %+v", result.Missing)
	}
}

func TestDoctorService_OSFiltering(t *testing.T) {
	manifest := makeManifest([]domain.Module{
		{Name: "mac-only", File: "mac.sh", RequiresBin: []string{"pbcopy"}, OS: []string{"Mac"}},
		{Name: "linux-only", File: "linux.sh", RequiresBin: []string{"xclip"}, OS: []string{"Linux"}},
	})
	// neither binary present
	lookup := newMock(nil, nil)

	result := app.NewDoctorService().Check(manifest, "Linux", lookup)

	names := depNames(result.Missing)
	if contains(names, "pbcopy") {
		t.Error("mac-only module should be filtered out on Linux")
	}
	if !contains(names, "xclip") {
		t.Error("linux-only module should be checked on Linux")
	}
	if result.ModuleCount != 1 {
		t.Errorf("ModuleCount = %d, want 1 (only linux-only)", result.ModuleCount)
	}
}

func TestDoctorService_GroupsByDepName(t *testing.T) {
	manifest := makeManifest([]domain.Module{
		{Name: "prompt", File: "prompt.sh", RequiresBin: []string{"starship"}},
		{Name: "starship-theme", File: "theme.sh", RequiresBin: []string{"starship"}},
	})
	lookup := newMock(nil, nil)

	result := app.NewDoctorService().Check(manifest, "Linux", lookup)

	if len(result.Missing) != 1 {
		t.Fatalf("expected 1 grouped entry, got %d: %v", len(result.Missing), result.Missing)
	}
	dep := result.Missing[0]
	if dep.Name != "starship" {
		t.Errorf("dep.Name = %q, want starship", dep.Name)
	}
	sort.Strings(dep.Modules)
	if dep.Modules[0] != "prompt" || dep.Modules[1] != "starship-theme" {
		t.Errorf("dep.Modules = %v, want [prompt starship-theme]", dep.Modules)
	}
}

// helpers

func depNames(deps []app.MissingDep) []string {
	names := make([]string, len(deps))
	for i, d := range deps {
		names[i] = d.Name
	}
	return names
}

func contains(slice []string, s string) bool {
	return slices.Contains(slice, s)
}
