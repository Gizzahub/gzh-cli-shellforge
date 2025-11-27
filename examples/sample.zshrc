# Sample .zshrc for migration testing
# This file demonstrates various section types and patterns

# --- Preamble ---
# Shell options and basic configuration
setopt auto_cd
setopt hist_ignore_dups
export LANG=en_US.UTF-8

# === OS Detection ===
case "$(uname -s)" in
  Darwin)
    export MACHINE="Mac"
    ;;
  Linux)
    export MACHINE="Linux"
    ;;
  *)
    export MACHINE="Unknown"
    ;;
esac

# --- PATH Initialization ---
export PATH="/usr/local/bin:$PATH"
export PATH="/usr/local/sbin:$PATH"

# --- Homebrew PATH (Mac only) ---
if [ "$MACHINE" = "Mac" ]; then
  if [ -d "/opt/homebrew" ]; then
    eval "$(/opt/homebrew/bin/brew shellenv)"
  elif [ -d "/usr/local/Homebrew" ]; then
    eval "$(/usr/local/Homebrew/bin/brew shellenv)"
  fi
fi

# === NVM Initialization ===
export NVM_DIR="$HOME/.nvm"
if [ -s "$NVM_DIR/nvm.sh" ]; then
  \. "$NVM_DIR/nvm.sh"
fi
if [ -s "$NVM_DIR/bash_completion" ]; then
  \. "$NVM_DIR/bash_completion"
fi

# === Python Environment ===
if command -v pyenv 1>/dev/null 2>&1; then
  eval "$(pyenv init -)"
fi

# === Go Environment ===
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"

# --- Git Aliases ---
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline --graph --decorate'
alias gd='git diff'

# --- System Aliases ---
alias ll='ls -la'
alias la='ls -A'
alias l='ls -CF'

# Platform-specific aliases
if [ "$MACHINE" = "Mac" ]; then
  alias updatebrew='brew update && brew upgrade && brew cleanup'
elif [ "$MACHINE" = "Linux" ]; then
  alias updateapt='sudo apt update && sudo apt upgrade -y'
fi

# --- Helper Functions ---
function mkcd() {
  mkdir -p "$1" && cd "$1"
}

function extract() {
  if [ -f "$1" ]; then
    case "$1" in
      *.tar.bz2)   tar xjf "$1"   ;;
      *.tar.gz)    tar xzf "$1"   ;;
      *.bz2)       bunzip2 "$1"   ;;
      *.rar)       unrar x "$1"   ;;
      *.gz)        gunzip "$1"    ;;
      *.tar)       tar xf "$1"    ;;
      *.tbz2)      tar xjf "$1"   ;;
      *.tgz)       tar xzf "$1"   ;;
      *.zip)       unzip "$1"     ;;
      *.Z)         uncompress "$1";;
      *.7z)        7z x "$1"      ;;
      *)           echo "'$1' cannot be extracted" ;;
    esac
  else
    echo "'$1' is not a valid file"
  fi
}

# === Prompt Customization ===
# Set up a simple prompt with git branch
autoload -Uz vcs_info
precmd() { vcs_info }
zstyle ':vcs_info:git:*' formats ' (%b)'
setopt PROMPT_SUBST
PROMPT='%F{green}%n@%m%f:%F{blue}%~%f%F{yellow}${vcs_info_msg_0_}%f$ '

# --- Completion System ---
autoload -Uz compinit
compinit

# === Custom Environment Variables ===
export EDITOR=vim
export VISUAL=vim
export PAGER=less

# Project-specific paths
if [ -d "$HOME/projects/bin" ]; then
  export PATH="$HOME/projects/bin:$PATH"
fi
