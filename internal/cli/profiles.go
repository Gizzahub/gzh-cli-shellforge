package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain/shellmeta"
)

type profilesFlags struct {
	dataDir string
	verbose bool
}

func newProfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profiles",
		Short: "Query shell profile initialization metadata",
		Long: `Profiles provides information about shell initialization files and contexts.

This command helps you understand:
  - Which files are loaded for different shells (bash, zsh, fish)
  - How login vs interactive shells differ
  - What happens in various execution contexts (cron, SSH, containers)
  - Language version manager initialization (rbenv, nvm, pyenv)`,
	}

	cmd.AddCommand(newProfilesListCmd())
	cmd.AddCommand(newProfilesShowCmd())
	cmd.AddCommand(newProfilesCheckCmd())

	return cmd
}

func newProfilesListCmd() *cobra.Command {
	flags := &profilesFlags{}

	cmd := &cobra.Command{
		Use:   "list [category]",
		Short: "List available profile categories or items",
		Long: `List displays available profile categories or items within a category.

Categories:
  distributions    Supported Linux distributions
  managers         Language version managers (rbenv, nvm, pyenv, etc.)
  desktops         Desktop environments (GNOME, KDE, XFCE, etc.)
  modes            Shell execution modes (login, interactive, etc.)
  multiplexers     Terminal multiplexers (tmux, screen, zellij)`,
		Example: `  # List all categories
  gz-shellforge profiles list

  # List supported Linux distributions
  gz-shellforge profiles list distributions

  # List language version managers
  gz-shellforge profiles list managers`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			category := ""
			if len(args) > 0 {
				category = args[0]
			}
			return runProfilesList(category, flags)
		},
	}

	cmd.Flags().StringVar(&flags.dataDir, "data-dir", "", "Shell profiles data directory (default: embedded)")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func newProfilesShowCmd() *cobra.Command {
	flags := &profilesFlags{}

	cmd := &cobra.Command{
		Use:   "show <type> <name>",
		Short: "Show details about a specific profile item",
		Long: `Show displays detailed information about a specific profile item.

Types:
  os <name>            OS init files (mac, ubuntu, arch, fedora, debian)
  manager <name>       Language version manager (rbenv, nvm, pyenv, conda)
  desktop <name>       Desktop environment (gnome, kde, xfce, i3)
  mode <name>          Shell mode (login, interactive, non-login)
  multiplexer <name>   Terminal multiplexer (tmux, screen, zellij)`,
		Example: `  # Show macOS shell initialization
  gz-shellforge profiles show os mac

  # Show rbenv initialization info
  gz-shellforge profiles show manager rbenv

  # Show login shell mode details
  gz-shellforge profiles show mode login`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			itemType := args[0]
			name := args[1]
			return runProfilesShow(itemType, name, flags)
		},
	}

	cmd.Flags().StringVar(&flags.dataDir, "data-dir", "", "Shell profiles data directory (default: embedded)")
	cmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "Show detailed output")

	return cmd
}

func newProfilesCheckCmd() *cobra.Command {
	flags := &profilesFlags{}

	cmd := &cobra.Command{
		Use:   "check <context>",
		Short: "Check if shell profiles are loaded in a context",
		Long: `Check determines whether shell profiles are loaded in a given execution context.

Contexts:
  cron               Cron jobs
  at                 at command
  docker-exec        docker exec
  flatpak            Flatpak apps
  git-hooks          Git hooks
  github-actions     GitHub Actions
  gitlab-ci          GitLab CI
  jenkins            Jenkins
  ssh-forced-command SSH forced command`,
		Example: `  # Check if profiles load in cron
  gz-shellforge profiles check cron

  # Check GitHub Actions context
  gz-shellforge profiles check github-actions`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			context := args[0]
			return runProfilesCheck(context, flags)
		},
	}

	cmd.Flags().StringVar(&flags.dataDir, "data-dir", "", "Shell profiles data directory (default: embedded)")

	return cmd
}

func loadProfiles(flags *profilesFlags) (*shellmeta.ShellProfiles, error) {
	dataDir := flags.dataDir
	if dataDir == "" {
		// Try to find data directory relative to executable or current directory
		execPath, err := os.Executable()
		if err == nil {
			// Check relative to executable
			candidate := filepath.Join(filepath.Dir(execPath), "..", "data", "shell-profiles")
			if _, err := os.Stat(candidate); err == nil {
				dataDir = candidate
			}
		}

		if dataDir == "" {
			// Check relative to current directory
			candidate := filepath.Join("data", "shell-profiles")
			if _, err := os.Stat(candidate); err == nil {
				dataDir = candidate
			}
		}

		if dataDir == "" {
			return nil, fmt.Errorf("shell profiles data not found. Use --data-dir to specify location")
		}
	}

	loader := shellmeta.NewLoader(afero.NewOsFs())
	return loader.Load(dataDir)
}

func runProfilesList(category string, flags *profilesFlags) error {
	profiles, err := loadProfiles(flags)
	if err != nil {
		return err
	}

	if category == "" {
		// List all categories
		fmt.Println("Available profile categories:")
		fmt.Println()
		fmt.Println("  distributions    Supported Linux distributions")
		fmt.Println("  managers         Language version managers")
		fmt.Println("  desktops         Desktop environments")
		fmt.Println("  modes            Shell execution modes")
		fmt.Println("  multiplexers     Terminal multiplexers")
		fmt.Println()
		fmt.Println("Use 'profiles list <category>' to see items in a category.")
		return nil
	}

	category = strings.ToLower(category)

	switch category {
	case "distributions", "distros", "os":
		distros := profiles.ListSupportedDistributions()
		sort.Strings(distros)
		fmt.Println("Supported Linux distributions:")
		fmt.Println()
		for _, d := range distros {
			fmt.Printf("  %s\n", d)
		}
		fmt.Println()
		fmt.Println("Also supported: mac (macOS/Darwin)")

	case "managers", "version-managers", "lang":
		managers := profiles.ListLanguageVersionManagers()
		sort.Strings(managers)
		fmt.Println("Language version managers:")
		fmt.Println()
		for _, m := range managers {
			mgr := profiles.GetLanguageVersionManager(m)
			if mgr != nil && mgr.TypicalProblem != "" && flags.verbose {
				fmt.Printf("  %-12s  %s\n", m, mgr.TypicalProblem)
			} else {
				fmt.Printf("  %s\n", m)
			}
		}

	case "desktops", "desktop", "de":
		des := profiles.ListDesktopEnvironments()
		sort.Strings(des)
		fmt.Println("Desktop environments:")
		fmt.Println()
		for _, d := range des {
			de := profiles.GetDesktopEnvironment(d)
			if de != nil && de.Description != "" && flags.verbose {
				fmt.Printf("  %-12s  %s\n", d, de.Description)
			} else {
				fmt.Printf("  %s\n", d)
			}
		}

	case "modes", "mode", "shell-modes":
		fmt.Println("Shell execution modes:")
		fmt.Println()
		modes := []string{"login", "non-login", "interactive", "non-interactive", "restricted", "posix"}
		for _, m := range modes {
			mode := profiles.GetShellMode(m)
			if mode != nil {
				if flags.verbose && len(mode.UseCases) > 0 {
					fmt.Printf("  %-16s  %s\n", m, mode.UseCases[0])
				} else {
					fmt.Printf("  %s\n", m)
				}
			}
		}

	case "multiplexers", "mux", "terminal":
		fmt.Println("Terminal multiplexers:")
		fmt.Println()
		muxes := []string{"tmux", "screen", "zellij"}
		for _, m := range muxes {
			mux := profiles.GetTerminalMultiplexer(m)
			if mux != nil {
				fmt.Printf("  %s\n", m)
			}
		}

	default:
		return fmt.Errorf("unknown category: %s\n\nAvailable: distributions, managers, desktops, modes, multiplexers", category)
	}

	return nil
}

func runProfilesShow(itemType, name string, flags *profilesFlags) error {
	profiles, err := loadProfiles(flags)
	if err != nil {
		return err
	}

	itemType = strings.ToLower(itemType)
	name = strings.ToLower(name)

	switch itemType {
	case "os", "distribution", "distro":
		return showOSProfile(profiles, name, flags)

	case "manager", "version-manager", "lang":
		return showLanguageManager(profiles, name, flags)

	case "desktop", "de":
		return showDesktopEnvironment(profiles, name, flags)

	case "mode", "shell-mode":
		return showShellMode(profiles, name, flags)

	case "multiplexer", "mux":
		return showMultiplexer(profiles, name, flags)

	default:
		return fmt.Errorf("unknown type: %s\n\nAvailable: os, manager, desktop, mode, multiplexer", itemType)
	}
}

func showOSProfile(profiles *shellmeta.ShellProfiles, name string, flags *profilesFlags) error {
	loginFiles := profiles.GetLoginShellFiles(name)
	interactiveFiles := profiles.GetInteractiveShellFiles(name)
	defaultShell := profiles.GetDefaultShell(name)

	if loginFiles == nil && interactiveFiles == nil {
		return fmt.Errorf("OS not found: %s\n\nUse 'profiles list distributions' to see available options", name)
	}

	fmt.Printf("OS: %s\n", name)
	fmt.Println()

	if defaultShell != "" {
		fmt.Printf("Default shell: %s\n", defaultShell)
		fmt.Println()
	}

	if len(loginFiles) > 0 {
		fmt.Println("Login shell files:")
		for _, f := range loginFiles {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}

	if len(interactiveFiles) > 0 {
		fmt.Println("Interactive shell files:")
		for _, f := range interactiveFiles {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}

	// Show shell-specific files
	for _, shell := range []string{"bash", "zsh", "fish"} {
		files := profiles.GetInitFilesForOS(name, shell)
		if len(files) > 0 {
			fmt.Printf("%s files:\n", shell)
			for _, f := range files {
				fmt.Printf("  %s\n", f)
			}
			fmt.Println()
		}
	}

	return nil
}

func showLanguageManager(profiles *shellmeta.ShellProfiles, name string, flags *profilesFlags) error {
	mgr := profiles.GetLanguageVersionManager(name)
	if mgr == nil {
		return fmt.Errorf("language manager not found: %s\n\nUse 'profiles list managers' to see available options", name)
	}

	fmt.Printf("Language Version Manager: %s\n", name)
	fmt.Println()

	// InitCommand can be string or map
	switch cmd := mgr.InitCommand.(type) {
	case string:
		fmt.Printf("Init command: %s\n", cmd)
	case map[string]interface{}:
		fmt.Println("Init commands:")
		for shell, c := range cmd {
			fmt.Printf("  %-8s %v\n", shell+":", c)
		}
	}
	fmt.Println()

	// InitFiles can be []string or map
	switch files := mgr.InitFiles.(type) {
	case []interface{}:
		fmt.Println("Init files:")
		for _, f := range files {
			fmt.Printf("  %v\n", f)
		}
	case map[string]interface{}:
		fmt.Println("Init files:")
		for shell, f := range files {
			fmt.Printf("  %-8s %v\n", shell+":", f)
		}
	}

	if mgr.ShimsLocation != "" {
		fmt.Printf("\nShims location: %s\n", mgr.ShimsLocation)
	}

	if mgr.TypicalProblem != "" {
		fmt.Printf("\nTypical problem: %s\n", mgr.TypicalProblem)
	}

	if mgr.Note != "" {
		fmt.Printf("\nNote: %s\n", mgr.Note)
	}

	return nil
}

func showDesktopEnvironment(profiles *shellmeta.ShellProfiles, name string, flags *profilesFlags) error {
	de := profiles.GetDesktopEnvironment(name)
	if de == nil {
		return fmt.Errorf("desktop environment not found: %s\n\nUse 'profiles list desktops' to see available options", name)
	}

	fmt.Printf("Desktop Environment: %s\n", name)
	if de.Description != "" {
		fmt.Printf("Description: %s\n", de.Description)
	}
	fmt.Println()

	if len(de.Autostart) > 0 {
		fmt.Println("Autostart locations:")
		for _, f := range de.Autostart {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}

	if len(de.Environment) > 0 {
		fmt.Println("Environment files:")
		for _, f := range de.Environment {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}

	if len(de.Config) > 0 {
		fmt.Println("Config files:")
		for _, f := range de.Config {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}

	if len(de.Legacy) > 0 {
		fmt.Println("Legacy locations:")
		for _, f := range de.Legacy {
			fmt.Printf("  %s\n", f)
		}
	}

	return nil
}

func showShellMode(profiles *shellmeta.ShellProfiles, name string, flags *profilesFlags) error {
	mode := profiles.GetShellMode(name)
	if mode == nil {
		return fmt.Errorf("shell mode not found: %s\n\nUse 'profiles list modes' to see available options", name)
	}

	fmt.Printf("Shell Mode: %s\n", name)
	fmt.Println()

	if len(mode.Detection) > 0 {
		fmt.Println("Detection methods:")
		for _, d := range mode.Detection {
			fmt.Printf("  %s\n", d)
		}
		fmt.Println()
	}

	if len(mode.FilesLoaded) > 0 {
		fmt.Println("Files loaded by shell:")
		for shell, files := range mode.FilesLoaded {
			fmt.Printf("  %s:\n", shell)
			for _, f := range files {
				fmt.Printf("    %s\n", f)
			}
		}
		fmt.Println()
	}

	if len(mode.UseCases) > 0 {
		fmt.Println("Use cases:")
		for _, u := range mode.UseCases {
			fmt.Printf("  - %s\n", u)
		}
		fmt.Println()
	}

	if len(mode.Features) > 0 {
		fmt.Println("Features:")
		for _, f := range mode.Features {
			fmt.Printf("  - %s\n", f)
		}
		fmt.Println()
	}

	if len(mode.Restrictions) > 0 {
		fmt.Println("Restrictions:")
		for _, r := range mode.Restrictions {
			fmt.Printf("  - %s\n", r)
		}
	}

	return nil
}

func showMultiplexer(profiles *shellmeta.ShellProfiles, name string, flags *profilesFlags) error {
	mux := profiles.GetTerminalMultiplexer(name)
	if mux == nil {
		return fmt.Errorf("terminal multiplexer not found: %s\n\nAvailable: tmux, screen, zellij", name)
	}

	fmt.Printf("Terminal Multiplexer: %s\n", name)
	fmt.Println()

	if mux.NewSession != "" {
		fmt.Printf("New session:  %s\n", mux.NewSession)
	}
	if mux.NewWindow != "" {
		fmt.Printf("New window:   %s\n", mux.NewWindow)
	}
	if mux.NewPane != "" {
		fmt.Printf("New pane:     %s\n", mux.NewPane)
	}
	if mux.NewTab != "" {
		fmt.Printf("New tab:      %s\n", mux.NewTab)
	}
	if mux.DefaultShell != "" {
		fmt.Printf("Default shell: %s\n", mux.DefaultShell)
	}
	if mux.LoginShell != "" {
		fmt.Printf("Login shell:  %s\n", mux.LoginShell)
	}

	if len(mux.ShellProfile) > 0 {
		fmt.Println("\nShell profile behavior:")
		for k, v := range mux.ShellProfile {
			fmt.Printf("  %-12s %s\n", k+":", v)
		}
	}

	return nil
}

func runProfilesCheck(context string, flags *profilesFlags) error {
	profiles, err := loadProfiles(flags)
	if err != nil {
		return err
	}

	loaded := profiles.IsProfileLoadedInContext(context)

	if loaded {
		fmt.Printf("✓ Shell profiles ARE loaded in '%s' context\n", context)
	} else {
		fmt.Printf("✗ Shell profiles are NOT loaded in '%s' context\n", context)
		fmt.Println()
		fmt.Println("This means:")
		fmt.Println("  - PATH modifications won't be applied")
		fmt.Println("  - Aliases and functions won't be available")
		fmt.Println("  - Language version managers (rbenv, nvm, etc.) won't be initialized")
		fmt.Println()
		fmt.Println("Workarounds:")
		fmt.Println("  - Source profile explicitly: source ~/.profile")
		fmt.Println("  - Use login shell: bash -lc 'command'")
		fmt.Println("  - Set environment variables directly")
	}

	return nil
}
