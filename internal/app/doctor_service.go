package app

import (
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// MissingDep describes one external prerequisite that is absent.
type MissingDep struct {
	Name    string   // binary name or path
	Modules []string // modules that declare this dep
	Kind    string   // "binary" or "path"
}

// DoctorResult is the output of a doctor check run.
type DoctorResult struct {
	Missing     []MissingDep
	CheckedOS   string
	ModuleCount int
}

// AllOK returns true when no prerequisites are missing.
func (r *DoctorResult) AllOK() bool {
	return len(r.Missing) == 0
}

// DoctorService checks whether declared prerequisites exist on the current host.
type DoctorService struct{}

// NewDoctorService creates a DoctorService.
func NewDoctorService() *DoctorService { return &DoctorService{} }

// Check runs the prerequisite check for all modules that apply to targetOS.
// Missing entries are grouped by dep name so each install hint appears once.
func (s *DoctorService) Check(manifest *domain.Manifest, targetOS string, lookup domain.PrereqLookup) *DoctorResult {
	// grouped[name] → list of module names that need it
	binMissing := make(map[string][]string)
	pathMissing := make(map[string][]string)
	moduleCount := 0

	for _, mod := range manifest.Modules {
		if !mod.AppliesTo(targetOS) {
			continue
		}
		moduleCount++

		for _, bin := range mod.RequiresBin {
			if !lookup.HasBinary(bin) {
				binMissing[bin] = append(binMissing[bin], mod.Name)
			}
		}

		for _, p := range mod.RequiresPath {
			expanded := domain.ExpandPath(p)
			if !lookup.PathExists(expanded) {
				// key on the original (unexpanded) path for readability
				pathMissing[p] = append(pathMissing[p], mod.Name)
			}
		}
	}

	var missing []MissingDep
	for name, mods := range binMissing {
		missing = append(missing, MissingDep{Name: name, Modules: mods, Kind: "binary"})
	}
	for p, mods := range pathMissing {
		missing = append(missing, MissingDep{Name: p, Modules: mods, Kind: "path"})
	}

	return &DoctorResult{
		Missing:     missing,
		CheckedOS:   targetOS,
		ModuleCount: moduleCount,
	}
}
