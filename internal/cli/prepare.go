// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

const (
	macOSName            = "Mac"
	linuxOSName          = "Linux"
	brewPackageManagerID = "brew"
)

type prepareFlags struct {
	manifest string
	targetOS string
	check    bool
	dryRun   bool
	verbose  bool
}

func newPrepareCmd() *cobra.Command {
	flags := &prepareFlags{}

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Install external packages declared by manifest modules",
		Long: `Prepare reads the "packages:" declarations on each manifest module and
installs whatever is missing for the current OS, using the manifest's
declared package manager (brew, cask on Mac; apt on Linux).

Already-installed packages are left untouched, and modules that don't apply
to the target OS are skipped entirely. A structurally invalid manifest fails
before prepare touches the system.

  --check    report missing packages without installing anything
  --dry-run  print the install plan without installing anything

With neither flag, prepare installs missing packages and re-verifies
afterward. Install failures are reported but don't stop the remaining
installs.`,
		Example: `  # Report missing packages only (non-zero exit if any are missing)
  gz-shellforge prepare --check

  # Preview the install plan
  gz-shellforge prepare --dry-run

  # Install missing packages
  gz-shellforge prepare`,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetOS := flags.targetOS
			if targetOS == "" {
				targetOS = helpers.DetectOS()
			}
			return runPrepare(cmd.Context(), flags, defaultPackageManagers(targetOS))
		},
	}

	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().StringVar(&flags.targetOS, "os", "", "Target OS (auto-detected if omitted)")
	cmd.Flags().BoolVar(&flags.check, "check", false, "Report missing packages without installing")
	cmd.Flags().BoolVar(&flags.dryRun, "dry-run", false, "Print the install plan without installing")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show all declared packages, not just missing ones")

	return cmd
}

// defaultPackageManagers registers the production package managers relevant
// to targetOS. Other OSes get no managers — their declared packages show up
// as unsupported rather than erroring.
func defaultPackageManagers(targetOS string) map[string]domain.PackageManager {
	switch targetOS {
	case macOSName:
		return map[string]domain.PackageManager{
			brewPackageManagerID: domain.NewBrewFormulaManager(),
			"cask":               domain.NewBrewCaskManager(),
		}
	case linuxOSName:
		return map[string]domain.PackageManager{
			"apt": domain.NewAptManager(),
		}
	default:
		return map[string]domain.PackageManager{}
	}
}

// runPrepare runs the check/dry-run/apply flow. managers is injected so tests
// can exercise it with fakes instead of shelling out to brew/apt.
func runPrepare(ctx context.Context, flags *prepareFlags, managers map[string]domain.PackageManager) error {
	targetOS := flags.targetOS
	if targetOS == "" {
		targetOS = helpers.DetectOS()
	}

	services := factory.NewServices()
	manifest, err := services.Parser.Parse(flags.manifest)
	if err != nil {
		return clierrors.WrapError("manifest parsing", err)
	}

	svc := app.NewPrepareService(managers)

	switch {
	case flags.check:
		result, err := svc.Plan(ctx, manifest, targetOS)
		if err != nil {
			return clierrors.WrapError("prepare check", err)
		}
		printPrepareResult(result, flags.verbose, "check")
		if !result.AllSatisfied() {
			return fmt.Errorf("prepare check: %d package(s) not satisfied", countMissing(result))
		}
		return nil

	case flags.dryRun:
		result, err := svc.Plan(ctx, manifest, targetOS)
		if err != nil {
			return clierrors.WrapError("prepare dry-run", err)
		}
		printPrepareResult(result, flags.verbose, "dry-run")
		return nil

	default:
		result, err := svc.Apply(ctx, manifest, targetOS)
		if err != nil {
			return clierrors.WrapError("prepare", err)
		}
		printPrepareResult(result, flags.verbose, "apply")
		if len(result.Failed) > 0 {
			return fmt.Errorf("prepare: %d failure(s)", len(result.Failed))
		}
		return nil
	}
}

func countMissing(result *app.PrepareResult) int {
	count := len(result.Failed)
	for _, p := range result.Packages {
		if p.Supported && !p.Installed {
			count++
		}
	}
	return count
}

func printPrepareResult(result *app.PrepareResult, verbose bool, mode string) {
	fmt.Printf("Prepare (%s) — OS: %s, packages declared: %d\n\n", mode, result.CheckedOS, len(result.Packages))

	if mode == "apply" && len(result.Installed) > 0 {
		names := make([]string, len(result.Installed))
		for i, p := range result.Installed {
			names[i] = fmt.Sprintf("[%s] %s", p.Manager, p.Package)
		}
		sort.Strings(names)
		fmt.Printf("Installed (%d):\n", len(names))
		for _, n := range names {
			fmt.Printf("  %s\n", n)
		}
		fmt.Println()
	}

	missing := make([]app.PackageStatus, 0)
	for _, p := range result.Packages {
		if p.Supported && !p.Installed {
			missing = append(missing, p)
		}
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Manager != missing[j].Manager {
			return missing[i].Manager < missing[j].Manager
		}
		return missing[i].Package < missing[j].Package
	})

	if len(missing) > 0 {
		fmt.Printf("Missing (%d):\n", len(missing))
		for _, p := range missing {
			mods := append([]string(nil), p.Modules...)
			sort.Strings(mods)
			fmt.Printf("  [%s] %s\n", p.Manager, p.Package)
			fmt.Printf("           needed by: %v\n", mods)
		}
		fmt.Println()
	}

	if len(result.Failed) > 0 {
		fmt.Printf("Failed (%d):\n", len(result.Failed))
		for _, f := range result.Failed {
			fmt.Printf("  [%s] %s — %v\n", f.Manager, f.Package, f.Err)
		}
		fmt.Println()
	}

	if verbose {
		fmt.Println("All declared packages:")
		all := append([]app.PackageStatus(nil), result.Packages...)
		sort.Slice(all, func(i, j int) bool {
			if all[i].Manager != all[j].Manager {
				return all[i].Manager < all[j].Manager
			}
			return all[i].Package < all[j].Package
		})
		for _, p := range all {
			status := "missing"
			switch {
			case !p.Supported:
				status = "unsupported"
			case p.Installed:
				status = "installed"
			}
			fmt.Printf("  [%s] %s — %s\n", p.Manager, p.Package, status)
		}
		fmt.Println()
	}

	if len(missing) == 0 && len(result.Failed) == 0 {
		fmt.Println("✓ All supported packages satisfied.")
	}
}
