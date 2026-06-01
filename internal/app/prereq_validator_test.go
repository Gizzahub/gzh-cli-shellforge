package app_test

import (
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

func TestPrereqValidator_Name(t *testing.T) {
	v := app.NewPrereqValidator("Mac", newMock(nil, nil))
	if v.Name() != "prereq" {
		t.Errorf("expected name 'prereq', got %q", v.Name())
	}
}

func TestPrereqValidator_AllPresent(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "starship", File: "starship.sh", RequiresBin: []string{"starship"}, OS: []string{"Mac"}},
	})
	v := app.NewPrereqValidator("Mac", newMock([]string{"starship"}, nil))

	findings := v.Validate(m, "")

	if len(findings) != 0 {
		t.Errorf("expected no findings when all prereqs present, got %v", findings)
	}
}

func TestPrereqValidator_MissingBinary(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "starship", File: "starship.sh", RequiresBin: []string{"starship"}},
	})
	v := app.NewPrereqValidator("Linux", newMock(nil, nil))

	findings := v.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for missing binary")
	}
	f := findings[0]
	if f.Severity != app.SeverityWarn {
		t.Errorf("prereq findings must be warn severity, got %q", f.Severity)
	}
	if !strings.Contains(f.Message, "starship") {
		t.Errorf("message should mention 'starship', got %q", f.Message)
	}
}

func TestPrereqValidator_MissingPath(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "oh-my-zsh", File: "omz.sh", RequiresPath: []string{"$HOME/.oh-my-zsh"}},
	})
	v := app.NewPrereqValidator("Mac", newMock(nil, nil))

	findings := v.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for missing path")
	}
	if findings[0].Severity != app.SeverityWarn {
		t.Errorf("prereq findings must be warn severity, got %q", findings[0].Severity)
	}
}

func TestPrereqValidator_OSFiltering(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "mac-tool", File: "mac.sh", RequiresBin: []string{"pbcopy"}, OS: []string{"Mac"}},
	})
	// Running check for Linux — mac-tool should be skipped.
	v := app.NewPrereqValidator("Linux", newMock(nil, nil))

	findings := v.Validate(m, "")

	if len(findings) != 0 {
		t.Errorf("expected no findings (module filtered by OS), got %v", findings)
	}
}

func TestPrereqValidator_IgnoresModulesDir(t *testing.T) {
	// modulesDir is irrelevant to prereq checks.
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh", RequiresBin: []string{"missing-tool"}},
	})
	v := app.NewPrereqValidator("Linux", newMock(nil, nil))

	findingsWithDir := v.Validate(m, "/some/dir")
	findingsNoDir := v.Validate(m, "")

	if len(findingsWithDir) != len(findingsNoDir) {
		t.Errorf("modulesDir should have no effect on prereq findings")
	}
}

func TestPrereqValidator_ImplementsValidator(t *testing.T) {
	// Compile-time check: *PrereqValidator implements app.Validator.
	var _ app.Validator = app.NewPrereqValidator("Linux", domain.OsPrereqLookup{})
}
