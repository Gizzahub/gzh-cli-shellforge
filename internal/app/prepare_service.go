package app

import (
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// PackageStatus describes the state of one package declared by a module's
// "packages:" map.
type PackageStatus struct {
	Manager   string   // manifest packages key, e.g. "brew", "cask", "apt"
	Package   string   // package name
	Modules   []string // modules that declare this package
	Installed bool     // true if already installed (only meaningful when Supported)
	Supported bool     // false when no PackageManager is registered for Manager
}

// PrepareFailure records one package whose install status could not be
// determined or whose install attempt failed.
type PrepareFailure struct {
	Manager string
	Package string
	Err     error
}

// PrepareResult is the outcome of a prepare Plan or Apply run.
type PrepareResult struct {
	CheckedOS string
	Packages  []PackageStatus  // full plan, one entry per declared package
	Installed []PackageStatus  // packages actually installed by Apply (empty for Plan)
	Failed    []PrepareFailure // detection or install errors
}

// AllSatisfied reports whether every supported package is already installed
// and no failures occurred. Unsupported packages (no registered manager) do
// not block AllSatisfied — prepare only manages what it knows how to manage.
func (r *PrepareResult) AllSatisfied() bool {
	if len(r.Failed) > 0 {
		return false
	}
	for _, p := range r.Packages {
		if p.Supported && !p.Installed {
			return false
		}
	}
	return true
}

// PrepareService plans and (optionally) installs manifest-declared packages.
type PrepareService struct {
	managers map[string]domain.PackageManager
}

// NewPrepareService creates a PrepareService backed by the given package
// managers, keyed by their Name().
func NewPrepareService(managers map[string]domain.PackageManager) *PrepareService {
	return &PrepareService{managers: managers}
}

// Plan inspects packages declared by modules that apply to targetOS and
// reports install status without changing anything on the system. Used by
// both --check and --dry-run.
//
// The manifest is validated before any package is inspected — a structurally
// invalid manifest fails here, before prepare would ever touch the system.
func (s *PrepareService) Plan(manifest *domain.Manifest, targetOS string) (*PrepareResult, error) {
	if errs := manifest.Validate(); len(errs) > 0 {
		return nil, errs[0]
	}

	type pkgKey struct{ manager, name string }
	modulesFor := map[pkgKey][]string{}
	var order []pkgKey

	for _, mod := range manifest.Modules {
		if !mod.AppliesTo(targetOS) {
			continue
		}
		for manager, pkgs := range mod.Packages {
			for _, pkg := range pkgs {
				k := pkgKey{manager, pkg}
				if _, seen := modulesFor[k]; !seen {
					order = append(order, k)
				}
				modulesFor[k] = append(modulesFor[k], mod.Name)
			}
		}
	}

	result := &PrepareResult{CheckedOS: targetOS}
	for _, k := range order {
		status := PackageStatus{Manager: k.manager, Package: k.name, Modules: modulesFor[k]}

		mgr, ok := s.managers[k.manager]
		if !ok {
			result.Packages = append(result.Packages, status) // Supported=false
			continue
		}
		status.Supported = true

		installed, err := mgr.IsInstalled(k.name)
		if err != nil {
			result.Failed = append(result.Failed, PrepareFailure{Manager: k.manager, Package: k.name, Err: err})
			continue
		}
		status.Installed = installed
		result.Packages = append(result.Packages, status)
	}

	return result, nil
}

// Apply installs every missing, supported package (already-installed
// packages are left untouched — idempotent) and re-runs Plan afterward to
// confirm the final state. Install failures are collected in Failed rather
// than aborting the run, so one broken package doesn't block the rest;
// callers treat a non-empty Failed as a non-zero exit.
func (s *PrepareService) Apply(manifest *domain.Manifest, targetOS string) (*PrepareResult, error) {
	plan, err := s.Plan(manifest, targetOS)
	if err != nil {
		return nil, err
	}

	initialDetectionFailures := plan.Failed
	var installed []PackageStatus
	var installFailures []PrepareFailure
	for _, status := range plan.Packages {
		if status.Installed || !status.Supported {
			continue
		}
		mgr := s.managers[status.Manager]
		if err := mgr.Install(status.Package); err != nil {
			installFailures = append(installFailures, PrepareFailure{Manager: status.Manager, Package: status.Package, Err: err})
			continue
		}
		installed = append(installed, status)
	}

	final, err := s.Plan(manifest, targetOS)
	if err != nil {
		return nil, err
	}
	final.Installed = installed

	type failureKey struct{ manager, pkg string }
	finalDetectionFailures := make(map[failureKey]struct{}, len(final.Failed))
	for _, failure := range final.Failed {
		finalDetectionFailures[failureKey{failure.Manager, failure.Package}] = struct{}{}
	}
	for _, failure := range initialDetectionFailures {
		key := failureKey{failure.Manager, failure.Package}
		if _, duplicate := finalDetectionFailures[key]; duplicate {
			continue
		}
		final.Failed = append(final.Failed, failure)
	}
	final.Failed = append(final.Failed, installFailures...)
	return final, nil
}
