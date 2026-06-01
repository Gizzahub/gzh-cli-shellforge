package app

import (
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// PrereqValidator checks requires_bin and requires_path for OS-applicable modules.
// Findings are always warn severity — missing prereqs do not gate deploys.
// Wraps DoctorService so the same check logic is shared with the doctor command.
type PrereqValidator struct {
	targetOS string
	lookup   domain.PrereqLookup
}

// NewPrereqValidator creates a PrereqValidator for the given OS and lookup implementation.
func NewPrereqValidator(targetOS string, lookup domain.PrereqLookup) *PrereqValidator {
	return &PrereqValidator{targetOS: targetOS, lookup: lookup}
}

func (*PrereqValidator) Name() string { return "prereq" }

func (v *PrereqValidator) Validate(m *domain.Manifest, _ string) []Finding {
	result := NewDoctorService().Check(m, v.targetOS, v.lookup)

	findings := make([]Finding, 0, len(result.Missing))
	for _, dep := range result.Missing {
		findings = append(findings, Finding{
			Severity: SeverityWarn,
			Message:  fmt.Sprintf("missing %s %q (needed by: %s)", dep.Kind, dep.Name, strings.Join(dep.Modules, ", ")),
		})
	}
	return findings
}
