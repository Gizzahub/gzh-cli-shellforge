package shellmeta

import (
	"testing"

	"github.com/spf13/afero"
)

func TestLoader_Load(t *testing.T) {
	fs := afero.NewMemMapFs()
	setupTestFiles(t, fs)

	loader := NewLoader(fs)
	profiles, err := loader.Load("/data/shell-profiles")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify Core
	if profiles.Core == nil {
		t.Error("Core should not be nil")
	}

	// Verify Contexts
	if profiles.Contexts == nil {
		t.Error("Contexts should not be nil")
	}

	// Verify Dev
	if profiles.Dev == nil {
		t.Error("Dev should not be nil")
	}

	// Verify Automation
	if profiles.Automation == nil {
		t.Error("Automation should not be nil")
	}
}

func TestLoader_LoadCore(t *testing.T) {
	fs := afero.NewMemMapFs()
	setupTestFiles(t, fs)

	loader := NewLoader(fs)
	core, err := loader.LoadCore("/data/shell-profiles/core.yaml")
	if err != nil {
		t.Fatalf("LoadCore() error = %v", err)
	}

	// Check OS profiles
	if core.OSProfiles.Mac.LoginShell == nil {
		t.Error("Mac LoginShell should not be nil")
	}

	// Check default shells
	if core.DefaultShells.Mac != "zsh" {
		t.Errorf("Mac default shell = %v, want zsh", core.DefaultShells.Mac)
	}

	// Check Linux distributions
	if _, ok := core.OSProfiles.Linux.Distributions["ubuntu"]; !ok {
		t.Error("ubuntu distribution should exist")
	}
}

func TestLoader_LoadDev(t *testing.T) {
	fs := afero.NewMemMapFs()
	setupTestFiles(t, fs)

	loader := NewLoader(fs)
	dev, err := loader.LoadDev("/data/shell-profiles/dev.yaml")
	if err != nil {
		t.Fatalf("LoadDev() error = %v", err)
	}

	// Check language version managers
	if rbenv, ok := dev.LanguageVersionManagers["rbenv"]; !ok {
		t.Error("rbenv should exist")
	} else if rbenv.InitCommand == "" {
		t.Error("rbenv InitCommand should not be empty")
	}

	// Check nvm
	if nvm, ok := dev.LanguageVersionManagers["nvm"]; !ok {
		t.Error("nvm should exist")
	} else if !nvm.FunctionDefinitions {
		t.Error("nvm should have FunctionDefinitions = true")
	}
}

func TestLoader_LoadContexts(t *testing.T) {
	fs := afero.NewMemMapFs()
	setupTestFiles(t, fs)

	loader := NewLoader(fs)
	contexts, err := loader.LoadContexts("/data/shell-profiles/contexts.yaml")
	if err != nil {
		t.Fatalf("LoadContexts() error = %v", err)
	}

	// Check SSH profiles
	if len(contexts.SSHProfiles.User) == 0 {
		t.Error("SSH user profiles should not be empty")
	}

	// Check desktop environments
	if gnome, ok := contexts.DesktopEnvironments["gnome"]; !ok {
		t.Error("gnome should exist")
	} else if gnome.Description == "" {
		t.Error("gnome description should not be empty")
	}

	// Check shell modes
	if _, ok := contexts.ShellModes["login_shell"]; !ok {
		t.Error("login_shell mode should exist")
	}
}

func TestLoader_LoadAutomation(t *testing.T) {
	fs := afero.NewMemMapFs()
	setupTestFiles(t, fs)

	loader := NewLoader(fs)
	automation, err := loader.LoadAutomation("/data/shell-profiles/automation.yaml")
	if err != nil {
		t.Fatalf("LoadAutomation() error = %v", err)
	}

	// Check cron config
	if automation.ScheduledExecution.Cron.ShellProfileLoaded {
		t.Error("Cron should have ShellProfileLoaded = false")
	}

	// Check user switching
	if su, ok := automation.UserSwitching["su"]; !ok {
		t.Error("su should exist")
	} else if su.ShellProfileLoaded {
		t.Error("su should have ShellProfileLoaded = false")
	}

	// Check terminal multiplexers
	if _, ok := automation.TerminalMultiplexers["tmux"]; !ok {
		t.Error("tmux should exist")
	}
}

func TestLoader_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	loader := NewLoader(fs)

	_, err := loader.Load("/nonexistent")
	if err == nil {
		t.Error("Load() should return error for nonexistent directory")
	}
}

func TestLoader_InvalidYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/invalid/core.yaml", []byte("invalid: yaml: content:"), 0644)

	loader := NewLoader(fs)
	_, err := loader.LoadCore("/invalid/core.yaml")
	if err == nil {
		t.Error("LoadCore() should return error for invalid YAML")
	}
}

func TestLoader_InvalidContextsYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/invalid/contexts.yaml", []byte("invalid: yaml: content:"), 0644)

	loader := NewLoader(fs)
	_, err := loader.LoadContexts("/invalid/contexts.yaml")
	if err == nil {
		t.Error("LoadContexts() should return error for invalid YAML")
	}
}

func TestLoader_InvalidDevYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/invalid/dev.yaml", []byte("invalid: yaml: content:"), 0644)

	loader := NewLoader(fs)
	_, err := loader.LoadDev("/invalid/dev.yaml")
	if err == nil {
		t.Error("LoadDev() should return error for invalid YAML")
	}
}

func TestLoader_InvalidAutomationYAML(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "/invalid/automation.yaml", []byte("invalid: yaml: content:"), 0644)

	loader := NewLoader(fs)
	_, err := loader.LoadAutomation("/invalid/automation.yaml")
	if err == nil {
		t.Error("LoadAutomation() should return error for invalid YAML")
	}
}

func TestLoader_MissingContextsFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/partial", 0755)
	_ = afero.WriteFile(fs, "/partial/core.yaml", []byte("default_shells:\n  mac: zsh"), 0644)

	loader := NewLoader(fs)
	_, err := loader.Load("/partial")
	if err == nil {
		t.Error("Load() should return error when contexts.yaml is missing")
	}
}

func TestLoader_MissingDevFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/partial", 0755)
	_ = afero.WriteFile(fs, "/partial/core.yaml", []byte("default_shells:\n  mac: zsh"), 0644)
	_ = afero.WriteFile(fs, "/partial/contexts.yaml", []byte("ssh_profiles:\n  user: []"), 0644)

	loader := NewLoader(fs)
	_, err := loader.Load("/partial")
	if err == nil {
		t.Error("Load() should return error when dev.yaml is missing")
	}
}

func TestLoader_MissingAutomationFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = fs.MkdirAll("/partial", 0755)
	_ = afero.WriteFile(fs, "/partial/core.yaml", []byte("default_shells:\n  mac: zsh"), 0644)
	_ = afero.WriteFile(fs, "/partial/contexts.yaml", []byte("ssh_profiles:\n  user: []"), 0644)
	_ = afero.WriteFile(fs, "/partial/dev.yaml", []byte("language_version_managers: {}"), 0644)

	loader := NewLoader(fs)
	_, err := loader.Load("/partial")
	if err == nil {
		t.Error("Load() should return error when automation.yaml is missing")
	}
}

func setupTestFiles(t *testing.T, fs afero.Fs) {
	t.Helper()

	// Create directory
	_ = fs.MkdirAll("/data/shell-profiles", 0755)

	// core.yaml
	coreYAML := `
os_profiles:
  linux:
    distributions:
      ubuntu:
        login_shell: [/etc/profile, ~/.profile]
        interactive_shell: [/etc/bash.bashrc, ~/.bashrc]
        shell_types:
          bash: [~/.bashrc, ~/.bash_profile]
          zsh: [~/.zshrc, ~/.zprofile]
  mac:
    login_shell: [/etc/profile, ~/.bash_profile]
    interactive_shell: [/etc/bashrc, ~/.bashrc]
    shell_types:
      bash: [~/.bash_profile, ~/.bashrc]
      zsh: [~/.zshrc, ~/.zprofile]

load_priority:
  login_shell:
    - /etc/profile
    - ~/.bash_profile
  interactive_shell:
    - /etc/bashrc
    - ~/.bashrc
  zsh:
    - ~/.zshenv
    - ~/.zshrc
  fish:
    - ~/.config/fish/config.fish

shell_detection:
  bash:
    command: "bash --version"
    version_pattern: "GNU bash"

os_detection:
  mac:
    uname: "Darwin"
    check_file: "/System/Library/CoreServices/SystemVersion.plist"
  linux:
    uname: "Linux"
    distro_files:
      ubuntu: [/etc/lsb-release]

default_shells:
  mac: "zsh"
  linux:
    ubuntu: "bash"
`
	_ = afero.WriteFile(fs, "/data/shell-profiles/core.yaml", []byte(coreYAML), 0644)

	// contexts.yaml
	contextsYAML := `
ssh_profiles:
  user:
    - ~/.ssh/rc
    - ~/.ssh/environment
  system:
    - /etc/ssh/sshrc
  detection:
    env_var: "SSH_CONNECTION"
    tty_pattern: "^pts/"

x_window_profiles:
  display_manager:
    - ~/.xprofile
  manual_start:
    - ~/.xinitrc
  resources:
    - ~/.Xresources

desktop_environments:
  gnome:
    autostart: [~/.config/autostart/*.desktop]
    description: "GNOME Desktop Environment"
  kde:
    autostart: [~/.config/autostart/*.desktop]
    description: "KDE Plasma Desktop"

session_types:
  tty:
    detection:
      tty_pattern: "^tty[0-9]+$"
    init_files: [/etc/profile, ~/.profile]
  gui:
    detection:
      display_var: "DISPLAY"
    init_files: [~/.xprofile]

systemd_user:
  environment:
    - ~/.config/environment.d/*.conf
  units:
    - ~/.config/systemd/user/*.service
  targets:
    default: "default.target"
    graphical: "graphical-session.target"
  detection:
    command: "systemctl --user status"
    env_var: "XDG_RUNTIME_DIR"

display_managers:
  gdm:
    name: "GNOME Display Manager"
    init_files: [~/.xprofile]
    wayland_support: true
    default_session: "gnome"

shell_modes:
  login_shell:
    detection:
      - "shopt -q login_shell"
    files_loaded:
      bash: [/etc/profile, ~/.bash_profile]
    use_cases:
      - SSH login
      - Console login
  non_login_shell:
    detection:
      - "! shopt -q login_shell"
    files_loaded:
      bash: [~/.bashrc]
  interactive_shell:
    detection:
      - "[[ $- == *i* ]]"
    files_loaded:
      bash: [~/.bashrc]
    features:
      - Command line editing
      - Job control

xdg_environment:
  base_directories:
    XDG_CONFIG_HOME: ~/.config
    XDG_DATA_HOME: ~/.local/share
  session_type:
    XDG_SESSION_TYPE: "x11 or wayland"
  set_by:
    - Display manager
`
	_ = afero.WriteFile(fs, "/data/shell-profiles/contexts.yaml", []byte(contextsYAML), 0644)

	// dev.yaml
	devYAML := `
gui_app_contexts:
  ide_integrated_terminal:
    problem: "GUI apps don't inherit shell profile"
    affected_apps:
      - VSCode
      - IntelliJ IDEA
    scenarios:
      desktop_launcher:
        description: "App launched from menu"
        shell_profile_loaded: false
        environment_source: "systemd"
        typical_issues:
          - "rbenv not in PATH"
      terminal_launch:
        description: "App launched from terminal"
        shell_profile_loaded: true
        environment_source: "parent shell"
    solutions:
      vscode:
        - "Use terminal.integrated.shellArgs"
      general:
        - "Launch from terminal"

  desktop_file_execution:
    location:
      - ~/.local/share/applications/*.desktop
    exec_field:
      description: "Exec runs without profile"
      example: "Exec=/usr/bin/code %F"
      shell_profile_loaded: false
    workarounds:
      wrapped_command: "Exec=/bin/bash -lc 'code'"
      custom_desktop_file: "Modify .desktop file"

language_version_managers:
  rbenv:
    init_command: 'eval "$(rbenv init -)"'
    init_files: [~/.bashrc, ~/.zshrc]
    shims_location: ~/.rbenv/shims
    path_modification: true
    typical_problem: "ruby not found in IDE"

  nvm:
    init_command: "source ~/.nvm/nvm.sh"
    init_files: [~/.bashrc, ~/.zshrc]
    lazy_load: true
    function_definitions: true
    typical_problem: "nvm command not found"

  pyenv:
    init_command: 'eval "$(pyenv init -)"'
    init_files: [~/.bashrc, ~/.zshrc]
    shims_location: ~/.pyenv/shims
    path_modification: true

  conda:
    init_command: 'eval "$(conda shell.bash hook)"'
    init_files: [~/.bashrc, ~/.zshrc]
    path_modification: true
    environment_activation: true
    typical_problem: "conda not activated"
`
	_ = afero.WriteFile(fs, "/data/shell-profiles/dev.yaml", []byte(devYAML), 0644)

	// automation.yaml
	automationYAML := `
scheduled_execution:
  cron:
    user_crontab: "crontab -e"
    system_crontab: /etc/crontab
    cron_d: /etc/cron.d/*
    environment:
      default_path: "/usr/bin:/bin"
      default_shell: "/bin/sh"
      minimal_env: true
      no_profile: true
    shell_profile_loaded: false
    workarounds:
      source_profile: ". ~/.profile; script"
      use_bash_login: "bash -lc 'script'"
      set_in_crontab: "SHELL=/bin/bash"

  at:
    command: "at now + 1 hour"
    environment: "inherits current"
    shell_profile_loaded: false

  systemd_timer:
    user_timers: ~/.config/systemd/user/*.timer
    system_timers: /etc/systemd/system/*.timer
    environment:
      no_profile: true
      use_environment_file: true
    service_file_env:
      - "Environment=PATH=/usr/local/bin"
    workarounds:
      use_bash_login: "ExecStart=/bin/bash -lc 'script'"
      environment_d: "~/.config/environment.d/*.conf"

  macos_launchd:
    user_agents: ~/Library/LaunchAgents/*.plist
    system_daemons: /Library/LaunchDaemons/*.plist
    environment:
      no_profile: true
      use_environment_dict: true
    plist_env:
      key: "EnvironmentVariables"
      example: "<key>PATH</key>"
    workarounds:
      launchctl_setenv: "launchctl setenv VAR value"
      source_profile_in_script: "source ~/.profile"

user_switching:
  su:
    command: "su username"
    description: "Switch without login shell"
    shell_profile_loaded: false
    env_preserved: true
    pwd_preserved: true

  su_login:
    command: "su - username"
    description: "Switch with login shell"
    shell_profile_loaded: true
    env_reset: true
    pwd_changed: true
    home_changed: true

  sudo:
    command: "sudo command"
    description: "Run as root"
    shell_profile_loaded: false
    env_filtered: true
    env_whitelist: [PATH, TERM]

container_contexts:
  docker:
    docker_exec:
      command: "docker exec -it container bash"
      shell_type: "non-login, interactive"
      shell_profile_loaded: false
      workaround: "docker exec -it container bash -l"
    docker_run:
      command: "docker run -it image bash"
      shell_type: "depends on Dockerfile"
      dockerfile_shell: "SHELL [\"/bin/bash\", \"-l\", \"-c\"]"

  chroot:
    command: "chroot /new/root /bin/bash"
    description: "Change root"
    shell_profile_loaded: "depends"
    minimal_env: true

  flatpak:
    command: "flatpak run org.app"
    description: "Sandboxed"
    filesystem_isolation: true
    env_isolated: true
    shell_profile_loaded: false

  snap:
    command: "snap run app"
    description: "Confined"
    confinement: [strict, classic]
    shell_profile_loaded: "classic only"

  wsl:
    wsl_conf: /etc/wsl.conf
    environment_handling: "Windows/Linux"
    init_files: [/etc/profile, ~/.bashrc]
    windows_path_append: true
    interop: true

  termux:
    prefix: /data/data/com.termux/files/usr
    non_standard_paths: true
    init_files: [$PREFIX/etc/profile, ~/.bashrc]

remote_execution:
  ssh_non_interactive:
    command: "ssh user@host 'command'"
    shell_type: "non-interactive"
    bash_behavior: "sources ~/.bashrc"
    zsh_behavior: "sources ~/.zshenv"
    workaround: "ssh user@host 'bash -lc command'"

  ssh_forced_command:
    location: "~/.ssh/authorized_keys"
    example: 'command="/path/to/script" ssh-rsa'
    shell_profile_loaded: false

  git_hooks:
    location: ".git/hooks/*"
    environment: "minimal"
    shell_profile_loaded: false
    workaround: "source profile in hook"

  ci_cd:
    github_actions:
      shell_default: "bash"
      shell_options: ["-e", "-o pipefail"]
      shell_profile_loaded: false
      workaround: "source ~/.profile"
    gitlab_ci:
      shell_default: "bash"
      shell_profile_loaded: false
      workaround: "before_script: source ~/.profile"
    jenkins:
      shell_configurable: true
      shell_profile_loaded: false
      workaround: "#!/bin/bash -l"

terminal_multiplexers:
  tmux:
    new_session: "tmux new -s name"
    new_window: "Ctrl-b c"
    new_pane: "Ctrl-b %"
    default_shell: "set-option -g default-shell /bin/zsh"
    login_shell: "set -g default-command '${SHELL} -l'"
    shell_profile:
      new_session: "depends"
      new_window: "non-login"
      new_pane: "non-login"

  screen:
    new_session: "screen"
    new_window: "Ctrl-a c"
    login_shell: "shell -$SHELL"
    shell_profile:
      new_session: "login if configured"
      new_window: "non-login"

  zellij:
    new_tab: "Ctrl-t t"
    new_pane: "Ctrl-t n"
    default_shell: "configurable"
    shell_profile:
      new_tab: "depends"

system_environment:
  pam:
    system_config: /etc/security/pam_env.conf
    user_config: ~/.pam_environment
    loaded_at: "login"
    scope: "session"

  etc_environment:
    file: /etc/environment
    format: "KEY=value"
    loaded_at: "PAM start"
    scope: "all users"

  profile_d:
    location: /etc/profile.d/*.sh
    loaded_by: /etc/profile
    scope: "login shells"

  environment_d:
    system: /etc/environment.d/*.conf
    user: ~/.config/environment.d/*.conf
    format: "KEY=value"
    loaded_by: "systemd"
    scope: "systemd services"
`
	_ = afero.WriteFile(fs, "/data/shell-profiles/automation.yaml", []byte(automationYAML), 0644)
}
