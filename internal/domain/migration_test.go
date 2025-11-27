package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCategorizeSection(t *testing.T) {
	tests := []struct {
		name        string
		sectionName string
		content     string
		want        TemplateCategory
	}{
		{
			name:        "preamble by name",
			sectionName: "Preamble",
			content:     "# Some initial setup",
			want:        CategoryInitD,
		},
		{
			name:        "PATH section goes to init.d",
			sectionName: "PATH Setup",
			content:     "export PATH=/usr/local/bin:$PATH",
			want:        CategoryInitD,
		},
		{
			name:        "PATH modification in content",
			sectionName: "Environment",
			content:     "PATH=/usr/local/bin:$PATH\nexport PATH",
			want:        CategoryInitD,
		},
		{
			name:        "aliases go to rc_post.d",
			sectionName: "My Aliases",
			content:     "alias ll='ls -la'\nalias gs='git status'",
			want:        CategoryRcPostD,
		},
		{
			name:        "functions go to rc_post.d",
			sectionName: "Helper Functions",
			content:     "function mkcd() {\n  mkdir -p $1 && cd $1\n}",
			want:        CategoryRcPostD,
		},
		{
			name:        "tool initialization goes to rc_pre.d",
			sectionName: "NVM Setup",
			content:     "export NVM_DIR=\"$HOME/.nvm\"\n[ -s \"$NVM_DIR/nvm.sh\" ] && . \"$NVM_DIR/nvm.sh\"",
			want:        CategoryRcPreD,
		},
		{
			name:        "conda initialization",
			sectionName: "Conda",
			content:     "__conda_setup=\"$('/usr/local/bin/conda' 'shell.bash' 'hook')\"",
			want:        CategoryRcPreD,
		},
		{
			name:        "unknown section defaults to rc_pre.d",
			sectionName: "Random Section",
			content:     "echo 'Hello World'",
			want:        CategoryRcPreD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CategorizeSection(tt.sectionName, tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInferDependencies(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "no dependencies",
			content: "alias ll='ls -la'",
			want:    []string{},
		},
		{
			name:    "MACHINE variable requires os-detection",
			content: "if [ \"$MACHINE\" = \"Mac\" ]; then\n  echo 'Mac'\nfi",
			want:    []string{"os-detection"},
		},
		{
			name:    "brew command requires brew-path",
			content: "if command -v brew >/dev/null; then\n  eval \"$(brew shellenv)\"\nfi",
			want:    []string{"brew-path"},
		},
		{
			name:    "both MACHINE and brew",
			content: "if [ \"$MACHINE\" = \"Mac\" ]; then\n  brew install something\nfi",
			want:    []string{"os-detection", "brew-path"},
		},
		{
			name:    "MACHINE with braces",
			content: "case ${MACHINE} in\n  Mac) echo 'mac';;\nesac",
			want:    []string{"os-detection"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferDependencies(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInferOSSupport(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "no OS-specific code defaults to all",
			content: "export EDITOR=vim",
			want:    []string{"Mac", "Linux"},
		},
		{
			name: "case statement with Mac and Linux",
			content: `case $MACHINE in
  Mac)
    # Mac stuff
    ;;
  Linux)
    # Linux stuff
    ;;
esac`,
			want: []string{"Mac", "Linux"},
		},
		{
			name: "Mac only",
			content: `case $MACHINE in
  Mac)
    # Mac stuff
    ;;
esac`,
			want: []string{"Mac"},
		},
		{
			name: "Linux only",
			content: `case "$MACHINE" in
  Linux)
    # Linux stuff
    ;;
esac`,
			want: []string{"Linux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferOSSupport(tt.content)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateModuleName(t *testing.T) {
	tests := []struct {
		name        string
		sectionName string
		index       int
		want        string
	}{
		{
			name:        "simple name",
			sectionName: "NVM Setup",
			index:       1,
			want:        "nvm-setup",
		},
		{
			name:        "all caps",
			sectionName: "PATH CONFIGURATION",
			index:       0,
			want:        "path-configuration",
		},
		{
			name:        "special characters",
			sectionName: "My Aliases (Git)",
			index:       2,
			want:        "my-aliases-git",
		},
		{
			name:        "multiple spaces and dashes",
			sectionName: "Tool   ---   Setup",
			index:       3,
			want:        "tool-setup",
		},
		{
			name:        "empty name generates generic",
			sectionName: "   ",
			index:       5,
			want:        "section-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateModuleName(tt.sectionName, tt.index)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateFileName(t *testing.T) {
	tests := []struct {
		name       string
		category   TemplateCategory
		moduleName string
		index      int
		want       string
	}{
		{
			name:       "preamble special case",
			category:   CategoryInitD,
			moduleName: "preamble",
			index:      0,
			want:       "init.d/00-preamble.sh",
		},
		{
			name:       "init.d with index",
			category:   CategoryInitD,
			moduleName: "path-setup",
			index:      0,
			want:       "init.d/10-path-setup.sh",
		},
		{
			name:       "init.d second item",
			category:   CategoryInitD,
			moduleName: "brew-path",
			index:      1,
			want:       "init.d/20-brew-path.sh",
		},
		{
			name:       "rc_pre.d",
			category:   CategoryRcPreD,
			moduleName: "nvm",
			index:      0,
			want:       "rc_pre.d/nvm.sh",
		},
		{
			name:       "rc_post.d",
			category:   CategoryRcPostD,
			moduleName: "aliases",
			index:      0,
			want:       "rc_post.d/aliases.sh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFileName(tt.category, tt.moduleName, tt.index)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSectionPattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string // Expected section names
	}{
		{
			name: "dashes style",
			input: `# --- PATH Setup ---
export PATH=/usr/local/bin:$PATH`,
			want: []string{"PATH Setup"},
		},
		{
			name: "equals style",
			input: `# === Tool Initialization ===
source nvm.sh`,
			want: []string{"Tool Initialization"},
		},
		{
			name: "hash style",
			input: `## My Aliases
alias ll='ls -la'`,
			want: []string{"My Aliases"},
		},
		{
			name: "multiple sections",
			input: `# --- Section 1 ---
content1
# === Section 2 ===
content2`,
			want: []string{"Section 1", "Section 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := SectionPattern.FindAllStringSubmatch(tt.input, -1)
			got := make([]string, 0, len(matches))
			for _, match := range matches {
				// The regex has two capture groups, check both
				if len(match) > 1 && match[1] != "" {
					got = append(got, match[1])
				} else if len(match) > 2 && match[2] != "" {
					got = append(got, match[2])
				}
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllCapsPattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name: "all caps section",
			input: `# PATH CONFIGURATION
export PATH`,
			want: []string{"PATH CONFIGURATION"},
		},
		{
			name: "multiple all caps",
			input: `# INITIALIZATION
setup
# ALIASES AND FUNCTIONS
alias ll='ls -la'`,
			want: []string{"INITIALIZATION", "ALIASES AND FUNCTIONS"},
		},
		{
			name: "not all caps",
			input: `# Path Setup
export PATH`,
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := AllCapsPattern.FindAllStringSubmatch(tt.input, -1)
			got := make([]string, 0, len(matches))
			for _, match := range matches {
				if len(match) > 1 {
					got = append(got, match[1])
				}
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationResult(t *testing.T) {
	t.Run("creates new result", func(t *testing.T) {
		result := NewMigrationResult()
		assert.NotNil(t, result)
		assert.Empty(t, result.Sections)
		assert.Empty(t, result.Modules)
		assert.Empty(t, result.Warnings)
	})

	t.Run("adds warnings", func(t *testing.T) {
		result := NewMigrationResult()
		result.AddWarning("Warning 1")
		result.AddWarning("Warning %d", 2)

		assert.Len(t, result.Warnings, 2)
		assert.Equal(t, "Warning 1", result.Warnings[0])
		assert.Equal(t, "Warning 2", result.Warnings[1])
	})

	t.Run("generates manifest", func(t *testing.T) {
		result := NewMigrationResult()
		result.Modules = []Module{
			{Name: "test", File: "test.sh"},
		}

		manifest := result.GenerateManifest()
		assert.NotNil(t, manifest)
		assert.Len(t, manifest.Modules, 1)
		assert.Equal(t, "test", manifest.Modules[0].Name)
	})
}
