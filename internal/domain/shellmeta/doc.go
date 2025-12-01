// Package shellmeta provides shell profile metadata loading and querying.
//
// This package loads shell initialization file mappings from YAML files and
// provides query methods to retrieve information about:
//   - OS-specific shell initialization files (Linux, macOS)
//   - Shell types (bash, zsh, fish)
//   - Desktop environments (GNOME, KDE, XFCE)
//   - Language version managers (rbenv, nvm, pyenv, conda)
//   - Execution contexts (cron, Docker, CI/CD, SSH)
//   - Terminal multiplexers (tmux, screen)
//
// Data Files:
//
// The metadata is stored in data/shell-profiles/ directory:
//   - core.yaml: OS and shell basic initialization file mappings
//   - contexts.yaml: Execution context information (SSH, X Window, desktop environments)
//   - dev.yaml: Development environment issues and solutions
//   - automation.yaml: Automation and isolated environment information
//
// Usage:
//
//	loader := shellmeta.NewLoader(afero.NewOsFs())
//	profiles, err := loader.Load("data/shell-profiles")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get init files for macOS bash
//	files := profiles.GetInitFilesForOS("mac", "bash")
//
//	// Check if profile is loaded in cron context
//	loaded := profiles.IsProfileLoadedInContext("cron") // false
//
//	// Get language version manager info
//	rbenv := profiles.GetLanguageVersionManager("rbenv")
package shellmeta
