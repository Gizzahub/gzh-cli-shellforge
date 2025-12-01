package shellmeta

import "strings"

// GetInitFilesForOS returns the initialization files for a specific OS and shell.
// os should be "mac" or a Linux distribution name (e.g., "ubuntu", "arch").
// shell should be "bash", "zsh", "fish", or "sh".
func (p *ShellProfiles) GetInitFilesForOS(os, shell string) []string {
	if p.Core == nil {
		return nil
	}

	os = strings.ToLower(os)
	shell = strings.ToLower(shell)

	// Check macOS
	if os == "mac" || os == "macos" || os == "darwin" {
		if files, ok := p.Core.OSProfiles.Mac.ShellTypes[shell]; ok {
			return files
		}
		return nil
	}

	// Check Linux distributions
	if profile, ok := p.Core.OSProfiles.Linux.Distributions[os]; ok {
		if files, ok := profile.ShellTypes[shell]; ok {
			return files
		}
	}

	return nil
}

// GetLoginShellFiles returns the login shell initialization files for an OS.
func (p *ShellProfiles) GetLoginShellFiles(os string) []string {
	if p.Core == nil {
		return nil
	}

	os = strings.ToLower(os)

	if os == "mac" || os == "macos" || os == "darwin" {
		return p.Core.OSProfiles.Mac.LoginShell
	}

	if profile, ok := p.Core.OSProfiles.Linux.Distributions[os]; ok {
		return profile.LoginShell
	}

	return nil
}

// GetInteractiveShellFiles returns the interactive shell initialization files for an OS.
func (p *ShellProfiles) GetInteractiveShellFiles(os string) []string {
	if p.Core == nil {
		return nil
	}

	os = strings.ToLower(os)

	if os == "mac" || os == "macos" || os == "darwin" {
		return p.Core.OSProfiles.Mac.InteractiveShell
	}

	if profile, ok := p.Core.OSProfiles.Linux.Distributions[os]; ok {
		return profile.InteractiveShell
	}

	return nil
}

// GetDefaultShell returns the default shell for an OS.
func (p *ShellProfiles) GetDefaultShell(os string) string {
	if p.Core == nil {
		return ""
	}

	os = strings.ToLower(os)

	if os == "mac" || os == "macos" || os == "darwin" {
		return p.Core.DefaultShells.Mac
	}

	if shell, ok := p.Core.DefaultShells.Linux[os]; ok {
		return shell
	}

	return ""
}

// GetLanguageVersionManager returns information about a language version manager.
func (p *ShellProfiles) GetLanguageVersionManager(name string) *LanguageVersionMgr {
	if p.Dev == nil {
		return nil
	}

	name = strings.ToLower(name)
	if mgr, ok := p.Dev.LanguageVersionManagers[name]; ok {
		return &mgr
	}

	return nil
}

// GetDesktopEnvironment returns information about a desktop environment.
func (p *ShellProfiles) GetDesktopEnvironment(name string) *DesktopEnvironment {
	if p.Contexts == nil {
		return nil
	}

	name = strings.ToLower(name)
	if de, ok := p.Contexts.DesktopEnvironments[name]; ok {
		return &de
	}

	return nil
}

// GetDisplayManager returns information about a display manager.
func (p *ShellProfiles) GetDisplayManager(name string) *DisplayManager {
	if p.Contexts == nil {
		return nil
	}

	name = strings.ToLower(name)
	if dm, ok := p.Contexts.DisplayManagers[name]; ok {
		return &dm
	}

	return nil
}

// GetShellMode returns information about a shell execution mode.
func (p *ShellProfiles) GetShellMode(mode string) *ShellMode {
	if p.Contexts == nil {
		return nil
	}

	mode = strings.ToLower(mode)
	// Handle common aliases
	switch mode {
	case "login":
		mode = "login_shell"
	case "nonlogin", "non-login":
		mode = "non_login_shell"
	case "interactive":
		mode = "interactive_shell"
	case "noninteractive", "non-interactive":
		mode = "non_interactive_shell"
	case "restricted":
		mode = "restricted_shell"
	case "posix":
		mode = "posix_mode"
	}

	if sm, ok := p.Contexts.ShellModes[mode]; ok {
		return &sm
	}

	return nil
}

// GetTerminalMultiplexer returns information about a terminal multiplexer.
func (p *ShellProfiles) GetTerminalMultiplexer(name string) *TerminalMultiplexer {
	if p.Automation == nil {
		return nil
	}

	name = strings.ToLower(name)
	if tm, ok := p.Automation.TerminalMultiplexers[name]; ok {
		return &tm
	}

	return nil
}

// GetUserSwitch returns information about a user switching command.
func (p *ShellProfiles) GetUserSwitch(name string) *UserSwitch {
	if p.Automation == nil {
		return nil
	}

	name = strings.ToLower(name)
	if us, ok := p.Automation.UserSwitching[name]; ok {
		return &us
	}

	return nil
}

// ListSupportedDistributions returns a list of supported Linux distributions.
func (p *ShellProfiles) ListSupportedDistributions() []string {
	if p.Core == nil {
		return nil
	}

	distros := make([]string, 0, len(p.Core.OSProfiles.Linux.Distributions))
	for name := range p.Core.OSProfiles.Linux.Distributions {
		distros = append(distros, name)
	}
	return distros
}

// ListLanguageVersionManagers returns a list of supported language version managers.
func (p *ShellProfiles) ListLanguageVersionManagers() []string {
	if p.Dev == nil {
		return nil
	}

	managers := make([]string, 0, len(p.Dev.LanguageVersionManagers))
	for name := range p.Dev.LanguageVersionManagers {
		managers = append(managers, name)
	}
	return managers
}

// ListDesktopEnvironments returns a list of supported desktop environments.
func (p *ShellProfiles) ListDesktopEnvironments() []string {
	if p.Contexts == nil {
		return nil
	}

	des := make([]string, 0, len(p.Contexts.DesktopEnvironments))
	for name := range p.Contexts.DesktopEnvironments {
		des = append(des, name)
	}
	return des
}

// IsProfileLoadedInContext checks if shell profiles are loaded in a given context.
func (p *ShellProfiles) IsProfileLoadedInContext(context string) bool {
	if p.Automation == nil {
		return true // Default to true if no automation info
	}

	context = strings.ToLower(context)

	switch context {
	case "cron":
		return p.Automation.ScheduledExecution.Cron.ShellProfileLoaded
	case "at":
		return p.Automation.ScheduledExecution.At.ShellProfileLoaded
	case "docker_exec", "docker-exec":
		return p.Automation.ContainerContexts.Docker.DockerExec.ShellProfileLoaded
	case "flatpak":
		return p.Automation.ContainerContexts.Flatpak.ShellProfileLoaded
	case "git_hooks", "git-hooks":
		return p.Automation.RemoteExecution.GitHooks.ShellProfileLoaded
	case "github_actions", "github-actions":
		return p.Automation.RemoteExecution.CICD.GithubActions.ShellProfileLoaded
	case "gitlab_ci", "gitlab-ci":
		return p.Automation.RemoteExecution.CICD.GitlabCI.ShellProfileLoaded
	case "jenkins":
		return p.Automation.RemoteExecution.CICD.Jenkins.ShellProfileLoaded
	case "ssh_forced_command", "ssh-forced-command":
		return p.Automation.RemoteExecution.SSHForcedCommand.ShellProfileLoaded
	}

	return true
}
