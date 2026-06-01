package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	clierrors "github.com/gizzahub/gzh-cli-shellforge/internal/cli/errors"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/factory"
	"github.com/gizzahub/gzh-cli-shellforge/internal/cli/helpers"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

type doctorFlags struct {
	manifest string
	targetOS string
	verbose  bool
}

func newDoctorCmd() *cobra.Command {
	flags := &doctorFlags{}

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check that external prerequisites declared in modules are installed",
		Long: `Doctor checks the prerequisites declared by each module in the manifest
against the current system.  It reports missing binaries (requires_bin) and
missing filesystem paths (requires_path) grouped by tool, with the list of
modules that depend on each missing item.

Doctor never installs anything.  Exit code 0 = all present, 1 = at least one
prerequisite is missing.`,
		Example: `  # Check against auto-detected OS
  gz-shellforge doctor

  # Check against a specific OS
  gz-shellforge doctor --os Linux

  # Show verbose output
  gz-shellforge doctor --verbose`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor(flags)
		},
	}

	cmd.Flags().StringVarP(&flags.manifest, "manifest", "m", "manifest.yaml", "Path to manifest file")
	cmd.Flags().StringVar(&flags.targetOS, "os", "", "Target OS (auto-detected if omitted)")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show all checked modules, not just failures")

	return cmd
}

func runDoctor(flags *doctorFlags) error {
	targetOS := flags.targetOS
	if targetOS == "" {
		targetOS = helpers.DetectOS()
	}

	services := factory.NewServices()
	manifest, err := services.Parser.Parse(flags.manifest)
	if err != nil {
		return clierrors.WrapError("manifest parsing", err)
	}

	svc := app.NewDoctorService()
	result := svc.Check(manifest, targetOS, domain.OsPrereqLookup{})

	printDoctorResult(result, flags.verbose)

	if !result.AllOK() {
		os.Exit(1)
	}
	return nil
}

func printDoctorResult(result *app.DoctorResult, verbose bool) {
	fmt.Printf("Doctor — OS: %s, modules checked: %d\n\n", result.CheckedOS, result.ModuleCount)

	if result.AllOK() {
		fmt.Println("✓ All prerequisites satisfied.")
		return
	}

	// Sort for deterministic output
	missing := result.Missing
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Kind != missing[j].Kind {
			return missing[i].Kind < missing[j].Kind
		}
		return missing[i].Name < missing[j].Name
	})

	fmt.Printf("✗ Missing prerequisites (%d):\n\n", len(missing))
	for _, dep := range missing {
		kindLabel := "binary"
		if dep.Kind == "path" {
			kindLabel = "path  "
		}
		sort.Strings(dep.Modules)
		fmt.Printf("  [%s] %s\n", kindLabel, dep.Name)
		fmt.Printf("           needed by: %s\n", strings.Join(dep.Modules, ", "))
	}

	fmt.Printf("\nRun 'gz-shellforge doctor --verbose' for full module list.\n")
}
