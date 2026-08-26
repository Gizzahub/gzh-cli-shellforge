package app

import (
	"fmt"
	"path/filepath"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// ManifestStructureValidator checks module names, required fields, and dep refs.
type ManifestStructureValidator struct{}

// Name returns this validator's stable pipeline name.
func (ManifestStructureValidator) Name() string { return "manifest-structure" }

// Validate reports manifest structure violations.
func (ManifestStructureValidator) Validate(m *domain.Manifest, _ string) []Finding {
	var findings []Finding
	for _, err := range m.Validate() {
		findings = append(findings, Finding{Severity: SeverityError, Message: err.Error()})
	}
	return findings
}

// CircularDependencyValidator checks for cycles in module dependency graph.
type CircularDependencyValidator struct{}

// Name returns this validator's stable pipeline name.
func (CircularDependencyValidator) Name() string { return "circular-dependencies" }

// Validate reports circular module dependencies.
func (CircularDependencyValidator) Validate(m *domain.Manifest, _ string) []Finding {
	depMap := make(map[string][]string, len(m.Modules))
	for _, mod := range m.Modules {
		depMap[mod.Name] = mod.Requires
	}
	for name := range depMap {
		visited := make(map[string]bool)
		if detectCycle(name, depMap, visited, make(map[string]bool)) {
			return []Finding{{
				Severity: SeverityError,
				Message:  fmt.Sprintf("circular dependency involving module '%s'", name),
			}}
		}
	}
	return nil
}

func detectCycle(node string, graph map[string][]string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true
	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			if detectCycle(neighbor, graph, visited, recStack) {
				return true
			}
		} else if recStack[neighbor] {
			return true
		}
	}
	recStack[node] = false
	return false
}

// FileExistenceValidator checks that all referenced module files exist.
// Uses the FileReader interface defined in builder.go.
type FileExistenceValidator struct {
	reader FileReader
}

// NewFileExistenceValidator creates a FileExistenceValidator with the given reader.
func NewFileExistenceValidator(reader FileReader) *FileExistenceValidator {
	return &FileExistenceValidator{reader: reader}
}

// Name returns this validator's stable pipeline name.
func (*FileExistenceValidator) Name() string { return "file-existence" }

// Validate reports manifest module files that do not exist.
func (v *FileExistenceValidator) Validate(m *domain.Manifest, modulesDir string) []Finding {
	var findings []Finding
	for _, mod := range m.Modules {
		filePath := filepath.Join(modulesDir, mod.File)
		if !v.reader.FileExists(filePath) {
			findings = append(findings, Finding{
				Severity: SeverityError,
				Module:   mod.Name,
				Message:  fmt.Sprintf("module file not found: %s", filePath),
			})
		}
	}
	return findings
}
