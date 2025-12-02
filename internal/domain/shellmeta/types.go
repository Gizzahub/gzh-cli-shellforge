package shellmeta

// ShellProfiles aggregates all shell profile metadata from multiple YAML files.
type ShellProfiles struct {
	Core       *CoreProfiles       `yaml:"-"`
	Contexts   *ContextProfiles    `yaml:"-"`
	Dev        *DevProfiles        `yaml:"-"`
	Automation *AutomationProfiles `yaml:"-"`
}

// CoreProfiles contains OS and shell basic initialization file mappings.
// Loaded from core.yaml
type CoreProfiles struct {
	OSProfiles     OSProfiles                `yaml:"os_profiles"`
	LoadPriority   LoadPriority              `yaml:"load_priority"`
	ShellDetection map[string]ShellDetection `yaml:"shell_detection"`
	OSDetection    OSDetection               `yaml:"os_detection"`
	DefaultShells  DefaultShells             `yaml:"default_shells"`
}

// OSProfiles contains OS-specific shell initialization files.
type OSProfiles struct {
	Linux LinuxProfiles `yaml:"linux"`
	Mac   OSProfile     `yaml:"mac"`
}

// LinuxProfiles contains Linux distribution-specific profiles.
type LinuxProfiles struct {
	Distributions map[string]OSProfile `yaml:"distributions"`
}

// OSProfile defines shell initialization files for an OS.
type OSProfile struct {
	LoginShell       []string            `yaml:"login_shell"`
	InteractiveShell []string            `yaml:"interactive_shell"`
	ShellTypes       map[string][]string `yaml:"shell_types"`
}

// LoadPriority defines the order in which shell files are loaded.
type LoadPriority struct {
	LoginShell       []string `yaml:"login_shell"`
	InteractiveShell []string `yaml:"interactive_shell"`
	Bash             []string `yaml:"bash,omitempty"`
	Zsh              []string `yaml:"zsh"`
	Fish             []string `yaml:"fish"`
}

// ShellDetection defines how to detect a shell and its version.
type ShellDetection struct {
	Command        string `yaml:"command"`
	VersionPattern string `yaml:"version_pattern"`
}

// OSDetection defines how to detect the operating system.
type OSDetection struct {
	Mac   MacDetection   `yaml:"mac"`
	Linux LinuxDetection `yaml:"linux"`
}

// MacDetection defines macOS detection patterns.
type MacDetection struct {
	Uname     string `yaml:"uname"`
	CheckFile string `yaml:"check_file"`
}

// LinuxDetection defines Linux detection patterns.
type LinuxDetection struct {
	Uname       string              `yaml:"uname"`
	DistroFiles map[string][]string `yaml:"distro_files"`
}

// DefaultShells defines default shell per OS.
type DefaultShells struct {
	Mac   string            `yaml:"mac"`
	Linux map[string]string `yaml:"linux"`
}

// ContextProfiles contains execution context information.
// Loaded from contexts.yaml
type ContextProfiles struct {
	SSHProfiles         SSHProfiles                   `yaml:"ssh_profiles"`
	XWindowProfiles     XWindowProfiles               `yaml:"x_window_profiles"`
	DesktopEnvironments map[string]DesktopEnvironment `yaml:"desktop_environments"`
	SessionTypes        map[string]SessionType        `yaml:"session_types"`
	SystemdUser         SystemdUser                   `yaml:"systemd_user"`
	DisplayManagers     map[string]DisplayManager     `yaml:"display_managers"`
	ShellModes          map[string]ShellMode          `yaml:"shell_modes"`
	XDGEnvironment      XDGEnvironment                `yaml:"xdg_environment"`
}

// SSHProfiles defines SSH-specific initialization files.
type SSHProfiles struct {
	ExecutionOrder SSHExecutionOrder `yaml:"execution_order"`
	User           []string          `yaml:"user"`
	System         []string          `yaml:"system"`
	Detection      SSHDetection      `yaml:"detection"`
}

// SSHExecutionOrder defines SSH rc file execution order.
type SSHExecutionOrder struct {
	Preferred string `yaml:"preferred"`
	Fallback  string `yaml:"fallback"`
}

// SSHDetection defines how to detect SSH sessions.
type SSHDetection struct {
	EnvVar     string `yaml:"env_var"`
	TTYPattern string `yaml:"tty_pattern"`
}

// XWindowProfiles defines X Window System initialization files.
type XWindowProfiles struct {
	DisplayManager DisplayManagerX `yaml:"display_manager"`
	ManualStart    []string        `yaml:"manual_start"`
	Resources      []string        `yaml:"resources"`
}

// DisplayManagerX defines Display Manager session entry point.
type DisplayManagerX struct {
	MainScript string   `yaml:"main_script"`
	Includes   []string `yaml:"includes"`
}

// DesktopEnvironment defines DE-specific configuration.
type DesktopEnvironment struct {
	Autostart   []string `yaml:"autostart,omitempty"`
	Legacy      []string `yaml:"legacy,omitempty"`
	Environment []string `yaml:"environment,omitempty"`
	Config      []string `yaml:"config,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

// SessionType defines session-specific detection and init files.
type SessionType struct {
	Detection SessionDetection `yaml:"detection"`
	InitFiles []string         `yaml:"init_files"`
}

// SessionDetection defines how to detect a session type.
type SessionDetection struct {
	TTYPattern   string `yaml:"tty_pattern,omitempty"`
	DisplayVar   string `yaml:"display_var,omitempty"`
	WaylandVar   string `yaml:"wayland_var,omitempty"`
	SSHVar       string `yaml:"ssh_var,omitempty"`
	SSHClientVar string `yaml:"ssh_client_var,omitempty"`
	SSHTTYVar    string `yaml:"ssh_tty_var,omitempty"`
}

// SystemdUser defines systemd user session configuration.
type SystemdUser struct {
	Environment []string         `yaml:"environment"`
	Units       []string         `yaml:"units"`
	Targets     SystemdTargets   `yaml:"targets"`
	Detection   SystemdDetection `yaml:"detection"`
}

// SystemdTargets defines systemd target names.
type SystemdTargets struct {
	Default   string `yaml:"default"`
	Graphical string `yaml:"graphical"`
}

// SystemdDetection defines how to detect systemd user session.
type SystemdDetection struct {
	Command string `yaml:"command"`
	EnvVar  string `yaml:"env_var"`
}

// DisplayManager defines display manager configuration.
type DisplayManager struct {
	Name           string   `yaml:"name"`
	InitFiles      []string `yaml:"init_files"`
	WaylandSupport bool     `yaml:"wayland_support"`
	DefaultSession string   `yaml:"default_session,omitempty"`
}

// ShellMode defines shell execution mode (login, interactive, etc).
type ShellMode struct {
	Detection    []string            `yaml:"detection,omitempty"`
	FilesLoaded  map[string][]string `yaml:"files_loaded,omitempty"`
	UseCases     []string            `yaml:"use_cases,omitempty"`
	Features     []string            `yaml:"features,omitempty"`
	Command      string              `yaml:"command,omitempty"`
	Restrictions []string            `yaml:"restrictions,omitempty"`
}

// XDGEnvironment defines XDG base directory specification.
type XDGEnvironment struct {
	BaseDirectories map[string]string `yaml:"base_directories"`
	SessionType     map[string]string `yaml:"session_type"`
	SetBy           []string          `yaml:"set_by"`
}

// DevProfiles contains development environment information.
// Loaded from dev.yaml
type DevProfiles struct {
	GUIAppContexts          GUIAppContexts                `yaml:"gui_app_contexts"`
	LanguageVersionManagers map[string]LanguageVersionMgr `yaml:"language_version_managers"`
}

// GUIAppContexts defines GUI application execution contexts.
type GUIAppContexts struct {
	IDEIntegratedTerminal IDEIntegratedTerminal `yaml:"ide_integrated_terminal"`
	DesktopFileExecution  DesktopFileExecution  `yaml:"desktop_file_execution"`
}

// IDEIntegratedTerminal defines IDE terminal issues and solutions.
type IDEIntegratedTerminal struct {
	Problem      string                    `yaml:"problem"`
	AffectedApps []string                  `yaml:"affected_apps"`
	Scenarios    map[string]LaunchScenario `yaml:"scenarios"`
	Solutions    map[string][]string       `yaml:"solutions"`
}

// LaunchScenario defines an application launch scenario.
type LaunchScenario struct {
	Description        string   `yaml:"description"`
	ShellProfileLoaded bool     `yaml:"shell_profile_loaded"`
	EnvironmentSource  string   `yaml:"environment_source"`
	TypicalIssues      []string `yaml:"typical_issues,omitempty"`
}

// DesktopFileExecution defines .desktop file execution behavior.
type DesktopFileExecution struct {
	Location    []string          `yaml:"location"`
	ExecField   ExecFieldInfo     `yaml:"exec_field"`
	Workarounds DesktopWorkaround `yaml:"workarounds"`
}

// ExecFieldInfo describes the Exec field behavior.
type ExecFieldInfo struct {
	Description        string `yaml:"description"`
	Example            string `yaml:"example"`
	ShellProfileLoaded bool   `yaml:"shell_profile_loaded"`
}

// DesktopWorkaround defines workarounds for desktop file execution.
type DesktopWorkaround struct {
	WrappedCommand    string `yaml:"wrapped_command"`
	CustomDesktopFile string `yaml:"custom_desktop_file"`
}

// LanguageVersionMgr defines a language version manager configuration.
// InitCommand and InitFiles can be strings, []string, or nested maps depending on complexity.
type LanguageVersionMgr struct {
	InitCommand           interface{} `yaml:"init_command"` // string or map[string]string
	InitFiles             interface{} `yaml:"init_files"`   // []string or map[string]interface{}
	ShimsLocation         string      `yaml:"shims_location,omitempty"`
	PathModification      bool        `yaml:"path_modification,omitempty"`
	FunctionDefinitions   bool        `yaml:"function_definitions,omitempty"`
	LazyLoad              bool        `yaml:"lazy_load,omitempty"`
	EnvironmentActivation bool        `yaml:"environment_activation,omitempty"`
	EnvFile               string      `yaml:"env_file,omitempty"`
	TypicalProblem        string      `yaml:"typical_problem,omitempty"`
	Note                  string      `yaml:"note,omitempty"`
}

// AutomationProfiles contains automation and isolated environment information.
// Loaded from automation.yaml
type AutomationProfiles struct {
	ScheduledExecution   ScheduledExecution             `yaml:"scheduled_execution"`
	UserSwitching        map[string]UserSwitch          `yaml:"user_switching"`
	ContainerContexts    ContainerContexts              `yaml:"container_contexts"`
	RemoteExecution      RemoteExecution                `yaml:"remote_execution"`
	TerminalMultiplexers map[string]TerminalMultiplexer `yaml:"terminal_multiplexers"`
	SystemEnvironment    SystemEnvironment              `yaml:"system_environment"`
}

// ScheduledExecution defines scheduled job environments.
type ScheduledExecution struct {
	Cron         CronConfig         `yaml:"cron"`
	At           AtConfig           `yaml:"at"`
	SystemdTimer SystemdTimerConfig `yaml:"systemd_timer"`
	MacOSLaunchd LaunchdConfig      `yaml:"macos_launchd"`
}

// CronConfig defines cron job environment.
type CronConfig struct {
	UserCrontab        string          `yaml:"user_crontab"`
	SystemCrontab      string          `yaml:"system_crontab"`
	CronD              string          `yaml:"cron_d"`
	Environment        CronEnvironment `yaml:"environment"`
	ShellProfileLoaded bool            `yaml:"shell_profile_loaded"`
	Workarounds        CronWorkarounds `yaml:"workarounds"`
}

// CronEnvironment defines cron's minimal environment.
type CronEnvironment struct {
	DefaultPath  string `yaml:"default_path"`
	DefaultShell string `yaml:"default_shell"`
	MinimalEnv   bool   `yaml:"minimal_env"`
	NoProfile    bool   `yaml:"no_profile"`
}

// CronWorkarounds defines workarounds for cron jobs.
type CronWorkarounds struct {
	SourceProfile string `yaml:"source_profile"`
	UseBashLogin  string `yaml:"use_bash_login"`
	SetInCrontab  string `yaml:"set_in_crontab"`
}

// AtConfig defines at command behavior.
type AtConfig struct {
	Command            string `yaml:"command"`
	Environment        string `yaml:"environment"`
	ShellProfileLoaded bool   `yaml:"shell_profile_loaded"`
}

// SystemdTimerConfig defines systemd timer environment.
type SystemdTimerConfig struct {
	UserTimers     string                  `yaml:"user_timers"`
	SystemTimers   string                  `yaml:"system_timers"`
	Environment    SystemdTimerEnvironment `yaml:"environment"`
	ServiceFileEnv []string                `yaml:"service_file_env"`
	Workarounds    SystemdTimerWorkarounds `yaml:"workarounds"`
}

// SystemdTimerEnvironment defines systemd timer environment settings.
type SystemdTimerEnvironment struct {
	NoProfile          bool `yaml:"no_profile"`
	UseEnvironmentFile bool `yaml:"use_environment_file"`
}

// SystemdTimerWorkarounds defines workarounds for systemd timers.
type SystemdTimerWorkarounds struct {
	UseBashLogin string `yaml:"use_bash_login"`
	EnvironmentD string `yaml:"environment_d"`
}

// LaunchdConfig defines macOS launchd configuration.
type LaunchdConfig struct {
	UserAgents    string             `yaml:"user_agents"`
	SystemDaemons string             `yaml:"system_daemons"`
	Environment   LaunchdEnvironment `yaml:"environment"`
	PlistEnv      PlistEnvConfig     `yaml:"plist_env"`
	Workarounds   LaunchdWorkarounds `yaml:"workarounds"`
	Notes         []string           `yaml:"notes,omitempty"`
}

// LaunchdEnvironment defines launchd environment settings.
type LaunchdEnvironment struct {
	NoProfile          bool `yaml:"no_profile"`
	UseEnvironmentDict bool `yaml:"use_environment_dict"`
}

// PlistEnvConfig defines plist environment variable configuration.
type PlistEnvConfig struct {
	Key     string `yaml:"key"`
	Example string `yaml:"example"`
}

// LaunchdWorkarounds defines workarounds for launchd.
type LaunchdWorkarounds struct {
	LaunchctlSetenv       string `yaml:"launchctl_setenv"`
	LaunchctlConfigPath   string `yaml:"launchctl_config_path,omitempty"`
	LaunchAgent           string `yaml:"launch_agent,omitempty"`
	SourceProfileInScript string `yaml:"source_profile_in_script"`
}

// UserSwitch defines user switching behavior.
type UserSwitch struct {
	Command            string      `yaml:"command"`
	Description        string      `yaml:"description"`
	ShellProfileLoaded interface{} `yaml:"shell_profile_loaded"`    // bool or string
	EnvPreserved       interface{} `yaml:"env_preserved,omitempty"` // bool or string (e.g., "partial")
	EnvReset           bool        `yaml:"env_reset,omitempty"`
	EnvFiltered        bool        `yaml:"env_filtered,omitempty"`
	EnvWhitelist       []string    `yaml:"env_whitelist,omitempty"`
	PwdPreserved       bool        `yaml:"pwd_preserved,omitempty"`
	PwdChanged         bool        `yaml:"pwd_changed,omitempty"`
	HomeChanged        bool        `yaml:"home_changed,omitempty"`
}

// ContainerContexts defines container and virtualization environments.
type ContainerContexts struct {
	Docker  DockerConfig  `yaml:"docker"`
	Chroot  ChrootConfig  `yaml:"chroot"`
	Flatpak FlatpakConfig `yaml:"flatpak"`
	Snap    SnapConfig    `yaml:"snap"`
	WSL     WSLConfig     `yaml:"wsl"`
	Termux  TermuxConfig  `yaml:"termux"`
}

// DockerConfig defines Docker execution contexts.
type DockerConfig struct {
	DockerExec DockerExecConfig `yaml:"docker_exec"`
	DockerRun  DockerRunConfig  `yaml:"docker_run"`
}

// DockerExecConfig defines docker exec behavior.
type DockerExecConfig struct {
	Command             string `yaml:"command"`
	ShellType           string `yaml:"shell_type"`
	LoginProfileLoaded  bool   `yaml:"login_profile_loaded,omitempty"`
	InteractiveRCLoaded bool   `yaml:"interactive_rc_loaded,omitempty"`
	ShellProfileLoaded  bool   `yaml:"shell_profile_loaded,omitempty"`
	Workaround          string `yaml:"workaround"`
}

// DockerRunConfig defines docker run behavior.
type DockerRunConfig struct {
	Command         string `yaml:"command"`
	ShellType       string `yaml:"shell_type"`
	DockerfileShell string `yaml:"dockerfile_shell"`
}

// ChrootConfig defines chroot environment.
type ChrootConfig struct {
	Command            string      `yaml:"command"`
	Description        string      `yaml:"description"`
	ShellProfileLoaded interface{} `yaml:"shell_profile_loaded"` // bool or string
	MinimalEnv         bool        `yaml:"minimal_env"`
}

// FlatpakConfig defines Flatpak sandbox environment.
type FlatpakConfig struct {
	Command             string `yaml:"command"`
	Description         string `yaml:"description"`
	FilesystemIsolation bool   `yaml:"filesystem_isolation"`
	EnvIsolated         bool   `yaml:"env_isolated"`
	ShellProfileLoaded  bool   `yaml:"shell_profile_loaded"`
}

// SnapConfig defines Snap confinement.
type SnapConfig struct {
	Command            string      `yaml:"command"`
	Description        string      `yaml:"description"`
	Confinement        []string    `yaml:"confinement"`
	ShellProfileLoaded interface{} `yaml:"shell_profile_loaded"` // bool or string
}

// WSLConfig defines Windows Subsystem for Linux.
type WSLConfig struct {
	WSLConf             string   `yaml:"wsl_conf"`
	EnvironmentHandling string   `yaml:"environment_handling"`
	InitFiles           []string `yaml:"init_files"`
	WindowsPathAppend   bool     `yaml:"windows_path_append"`
	Interop             bool     `yaml:"interop"`
}

// TermuxConfig defines Android Termux environment.
type TermuxConfig struct {
	Prefix           string   `yaml:"prefix"`
	NonStandardPaths bool     `yaml:"non_standard_paths"`
	InitFiles        []string `yaml:"init_files"`
}

// RemoteExecution defines remote and non-interactive execution.
type RemoteExecution struct {
	SSHNonInteractive SSHNonInteractiveConfig `yaml:"ssh_non_interactive"`
	SSHForcedCommand  SSHForcedCommandConfig  `yaml:"ssh_forced_command"`
	GitHooks          GitHooksConfig          `yaml:"git_hooks"`
	CICD              CICDConfig              `yaml:"ci_cd"`
}

// SSHNonInteractiveConfig defines SSH non-interactive execution.
type SSHNonInteractiveConfig struct {
	Command      string      `yaml:"command"`
	ShellType    string      `yaml:"shell_type"`
	BashBehavior interface{} `yaml:"bash_behavior"` // string or map (detailed behavior)
	ZshBehavior  string      `yaml:"zsh_behavior"`
	Workaround   string      `yaml:"workaround"`
}

// SSHForcedCommandConfig defines SSH forced command.
type SSHForcedCommandConfig struct {
	Location           string `yaml:"location"`
	Example            string `yaml:"example"`
	ShellProfileLoaded bool   `yaml:"shell_profile_loaded"`
}

// GitHooksConfig defines git hooks environment.
type GitHooksConfig struct {
	Location           string `yaml:"location"`
	Environment        string `yaml:"environment"`
	ShellProfileLoaded bool   `yaml:"shell_profile_loaded"`
	Workaround         string `yaml:"workaround"`
}

// CICDConfig defines CI/CD pipeline environments.
type CICDConfig struct {
	GithubActions CICDPlatformConfig `yaml:"github_actions"`
	GitlabCI      CICDPlatformConfig `yaml:"gitlab_ci"`
	Jenkins       CICDPlatformConfig `yaml:"jenkins"`
}

// CICDPlatformConfig defines a CI/CD platform configuration.
type CICDPlatformConfig struct {
	ShellDefault       string   `yaml:"shell_default,omitempty"`
	ShellOptions       []string `yaml:"shell_options,omitempty"`
	ShellConfigurable  bool     `yaml:"shell_configurable,omitempty"`
	ShellProfileLoaded bool     `yaml:"shell_profile_loaded"`
	Workaround         string   `yaml:"workaround"`
}

// TerminalMultiplexer defines terminal multiplexer configuration.
type TerminalMultiplexer struct {
	NewSession   string            `yaml:"new_session,omitempty"`
	NewWindow    string            `yaml:"new_window,omitempty"`
	NewPane      string            `yaml:"new_pane,omitempty"`
	NewTab       string            `yaml:"new_tab,omitempty"`
	DefaultShell string            `yaml:"default_shell,omitempty"`
	LoginShell   string            `yaml:"login_shell,omitempty"`
	ShellProfile map[string]string `yaml:"shell_profile,omitempty"`
}

// SystemEnvironment defines system-wide environment configuration.
type SystemEnvironment struct {
	PAM            PAMConfig            `yaml:"pam"`
	EtcEnvironment EtcEnvironmentConfig `yaml:"etc_environment"`
	ProfileD       ProfileDConfig       `yaml:"profile_d"`
	EnvironmentD   EnvironmentDConfig   `yaml:"environment_d"`
}

// PAMConfig defines PAM environment configuration.
type PAMConfig struct {
	SystemConfig string `yaml:"system_config"`
	UserConfig   string `yaml:"user_config"`
	LoadedAt     string `yaml:"loaded_at"`
	Scope        string `yaml:"scope"`
}

// EtcEnvironmentConfig defines /etc/environment configuration.
type EtcEnvironmentConfig struct {
	File     string `yaml:"file"`
	Format   string `yaml:"format"`
	LoadedAt string `yaml:"loaded_at"`
	Scope    string `yaml:"scope"`
}

// ProfileDConfig defines profile.d configuration.
type ProfileDConfig struct {
	Location string `yaml:"location"`
	LoadedBy string `yaml:"loaded_by"`
	Scope    string `yaml:"scope"`
}

// EnvironmentDConfig defines environment.d configuration.
type EnvironmentDConfig struct {
	System   string `yaml:"system"`
	User     string `yaml:"user"`
	Format   string `yaml:"format"`
	LoadedBy string `yaml:"loaded_by"`
	Scope    string `yaml:"scope"`
}
