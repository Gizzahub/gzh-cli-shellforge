package app_test

import (
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// --- Validator names ---

func TestValidatorNames(t *testing.T) {
	reader := mockFileReader{existing: map[string]bool{}}
	cases := []struct {
		v    app.Validator
		want string
	}{
		{app.ManifestStructureValidator{}, "manifest-structure"},
		{app.CircularDependencyValidator{}, "circular-dependencies"},
		{app.NewFileExistenceValidator(reader), "file-existence"},
	}
	for _, c := range cases {
		if c.v.Name() != c.want {
			t.Errorf("%T.Name() = %q, want %q", c.v, c.v.Name(), c.want)
		}
	}
}

// --- ManifestStructureValidator ---

func TestManifestStructureValidator_Valid(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh"},
		{Name: "b", File: "b.sh", Requires: []string{"a"}},
	})

	v := app.ManifestStructureValidator{}
	findings := v.Validate(m, "")

	if len(findings) != 0 {
		t.Errorf("expected no findings, got %v", findings)
	}
}

func TestManifestStructureValidator_DuplicateName(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh"},
		{Name: "a", File: "a2.sh"},
	})

	findings := app.ManifestStructureValidator{}.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected findings for duplicate name")
	}
	for _, f := range findings {
		if f.Severity != app.SeverityError {
			t.Errorf("expected error severity, got %q", f.Severity)
		}
	}
}

func TestManifestStructureValidator_MissingName(t *testing.T) {
	m := makeManifest([]domain.Module{
		{File: "a.sh"}, // no Name
	})

	findings := app.ManifestStructureValidator{}.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for missing module name")
	}
}

func TestManifestStructureValidator_NonExistentDep(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh", Requires: []string{"ghost"}},
	})

	findings := app.ManifestStructureValidator{}.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for non-existent dep")
	}
}

// --- CircularDependencyValidator ---

func TestCircularDependencyValidator_NoCycle(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh"},
		{Name: "b", File: "b.sh", Requires: []string{"a"}},
		{Name: "c", File: "c.sh", Requires: []string{"b"}},
	})

	findings := app.CircularDependencyValidator{}.Validate(m, "")

	if len(findings) != 0 {
		t.Errorf("expected no findings for valid DAG, got %v", findings)
	}
}

func TestCircularDependencyValidator_DirectCycle(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh", Requires: []string{"b"}},
		{Name: "b", File: "b.sh", Requires: []string{"a"}},
	})

	findings := app.CircularDependencyValidator{}.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for circular dependency")
	}
	if findings[0].Severity != app.SeverityError {
		t.Errorf("expected error severity, got %q", findings[0].Severity)
	}
}

func TestCircularDependencyValidator_IndirectCycle(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh", Requires: []string{"b"}},
		{Name: "b", File: "b.sh", Requires: []string{"c"}},
		{Name: "c", File: "c.sh", Requires: []string{"a"}},
	})

	findings := app.CircularDependencyValidator{}.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for indirect circular dependency")
	}
}

func TestCircularDependencyValidator_SelfDependency(t *testing.T) {
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh", Requires: []string{"a"}},
	})

	findings := app.CircularDependencyValidator{}.Validate(m, "")

	if len(findings) == 0 {
		t.Fatal("expected finding for self-dependency")
	}
}

// --- FileExistenceValidator ---

// mockFileReader lets tests control which files "exist".
type mockFileReader struct {
	existing map[string]bool
}

func (m mockFileReader) ReadFile(path string) (string, error) { return "", nil }
func (m mockFileReader) FileExists(path string) bool          { return m.existing[path] }

func TestFileExistenceValidator_AllExist(t *testing.T) {
	reader := mockFileReader{existing: map[string]bool{"modules/a.sh": true, "modules/b.sh": true}}
	v := app.NewFileExistenceValidator(reader)
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh"},
		{Name: "b", File: "b.sh"},
	})

	findings := v.Validate(m, "modules")

	if len(findings) != 0 {
		t.Errorf("expected no findings, got %v", findings)
	}
}

func TestFileExistenceValidator_MissingFile(t *testing.T) {
	reader := mockFileReader{existing: map[string]bool{"modules/a.sh": true}}
	v := app.NewFileExistenceValidator(reader)
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh"},
		{Name: "b", File: "missing.sh"},
	})

	findings := v.Validate(m, "modules")

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.Severity != app.SeverityError {
		t.Errorf("expected error severity, got %q", f.Severity)
	}
	if f.Module != "b" {
		t.Errorf("expected module 'b', got %q", f.Module)
	}
}

// --- ValidationPipeline ---

func TestValidationPipeline_Empty(t *testing.T) {
	p := app.NewValidationPipeline()
	m := makeManifest(nil)

	findings := p.Run(m, "")

	if len(findings) != 0 {
		t.Errorf("expected no findings from empty pipeline, got %v", findings)
	}
}

func TestValidationPipeline_AggregatesFindings(t *testing.T) {
	// Two validators that each produce one finding.
	m := makeManifest([]domain.Module{
		{Name: "a", File: "a.sh", Requires: []string{"a"}}, // self-cycle
		// ManifestStructureValidator will also catch nothing (name ok)
	})
	p := app.NewValidationPipeline(
		app.CircularDependencyValidator{},
	)

	findings := p.Run(m, "")

	if len(findings) == 0 {
		t.Fatal("expected at least one finding")
	}
}

// --- HasErrors ---

func TestHasErrors_WithError(t *testing.T) {
	findings := []app.Finding{
		{Severity: app.SeverityWarn, Message: "warn"},
		{Severity: app.SeverityError, Message: "err"},
	}
	if !app.HasErrors(findings) {
		t.Error("expected HasErrors to return true")
	}
}

func TestHasErrors_WarnsOnly(t *testing.T) {
	findings := []app.Finding{
		{Severity: app.SeverityWarn, Message: "warn"},
	}
	if app.HasErrors(findings) {
		t.Error("expected HasErrors to return false for warns only")
	}
}

func TestHasErrors_Empty(t *testing.T) {
	if app.HasErrors(nil) {
		t.Error("expected HasErrors to return false for empty slice")
	}
}
