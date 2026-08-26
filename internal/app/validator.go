package app

import "github.com/gizzahub/gzh-cli-shellforge/internal/domain"

// SeverityError and SeverityWarn identify the severity of a validation finding.
const (
	SeverityError = "error"
	SeverityWarn  = "warn"
)

// Finding is a single validation result with a severity level.
type Finding struct {
	Severity string // SeverityError or SeverityWarn
	Module   string // module name, or empty for manifest-level findings
	Message  string
}

// IsError returns true for error-severity findings.
func (f Finding) IsError() bool { return f.Severity == SeverityError }

// Validator runs a single named check against a manifest and returns findings.
type Validator interface {
	Name() string
	Validate(m *domain.Manifest, modulesDir string) []Finding
}

// ValidationPipeline runs a sequence of validators and aggregates findings.
type ValidationPipeline struct {
	validators []Validator
}

// NewValidationPipeline creates a pipeline from the given validators.
func NewValidationPipeline(validators ...Validator) *ValidationPipeline {
	return &ValidationPipeline{validators: validators}
}

// Run executes all validators and returns all findings.
func (p *ValidationPipeline) Run(m *domain.Manifest, modulesDir string) []Finding {
	all := make([]Finding, 0, len(p.validators))
	for _, v := range p.validators {
		all = append(all, v.Validate(m, modulesDir)...)
	}
	if len(all) == 0 {
		return nil
	}
	return all
}

// HasErrors returns true if any finding has error severity.
func HasErrors(findings []Finding) bool {
	for _, f := range findings {
		if f.IsError() {
			return true
		}
	}
	return false
}
