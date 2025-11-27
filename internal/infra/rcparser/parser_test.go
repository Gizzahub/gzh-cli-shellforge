package rcparser

import (
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_ParseFile(t *testing.T) {
	t.Run("parses simple RC file with sections", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		content := `# Initial setup
export LANG=en_US.UTF-8

# --- PATH Setup ---
export PATH=/usr/local/bin:$PATH

# === NVM Initialization ===
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh"

## My Aliases
alias ll='ls -la'
alias gs='git status'
`
		afero.WriteFile(fs, "/test.zshrc", []byte(content), 0644)

		parser := New(fs)
		result, err := parser.ParseFile("/test.zshrc")

		require.NoError(t, err)
		require.NotNil(t, result)

		// Should have 4 sections: preamble + 3 explicit sections
		assert.Len(t, result.Sections, 4)
		assert.Len(t, result.Modules, 4)

		// Check preamble
		assert.Equal(t, "Preamble", result.Sections[0].Name)
		assert.Contains(t, result.Sections[0].Content, "export LANG")

		// Check PATH section
		assert.Equal(t, "PATH Setup", result.Sections[1].Name)
		assert.Equal(t, domain.CategoryInitD, result.Sections[1].Category)
		assert.Contains(t, result.Sections[1].Content, "export PATH")

		// Check NVM section
		assert.Equal(t, "NVM Initialization", result.Sections[2].Name)
		assert.Equal(t, domain.CategoryRcPreD, result.Sections[2].Category)
		assert.Contains(t, result.Sections[2].Content, "NVM_DIR")

		// Check Aliases section
		assert.Equal(t, "My Aliases", result.Sections[3].Name)
		assert.Equal(t, domain.CategoryRcPostD, result.Sections[3].Category)
		assert.Contains(t, result.Sections[3].Content, "alias ll")
	})

	t.Run("handles file not found", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		parser := New(fs)

		result, err := parser.ParseFile("/nonexistent.zshrc")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to open file")
	})

	t.Run("handles empty file", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		afero.WriteFile(fs, "/empty.zshrc", []byte(""), 0644)

		parser := New(fs)
		result, err := parser.ParseFile("/empty.zshrc")

		require.NoError(t, err)
		assert.Empty(t, result.Sections)
		assert.Empty(t, result.Modules)
	})

	t.Run("handles file with only preamble", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		content := `# Just some setup
export PATH=/usr/local/bin:$PATH
export EDITOR=vim
`
		afero.WriteFile(fs, "/preamble.zshrc", []byte(content), 0644)

		parser := New(fs)
		result, err := parser.ParseFile("/preamble.zshrc")

		require.NoError(t, err)
		assert.Len(t, result.Sections, 1)
		assert.Equal(t, "Preamble", result.Sections[0].Name)
		assert.Contains(t, result.Sections[0].Content, "export PATH")
	})

	t.Run("infers dependencies correctly", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		content := `# --- Brew Setup ---
if [ "$MACHINE" = "Mac" ]; then
  eval "$(brew shellenv)"
fi
`
		afero.WriteFile(fs, "/test.zshrc", []byte(content), 0644)

		parser := New(fs)
		result, err := parser.ParseFile("/test.zshrc")

		require.NoError(t, err)
		require.Len(t, result.Modules, 1)

		// Should infer both os-detection and brew-path dependencies
		assert.Contains(t, result.Modules[0].Requires, "os-detection")
		assert.Contains(t, result.Modules[0].Requires, "brew-path")
	})

	t.Run("infers OS support correctly", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		content := `# --- OS Specific ---
case $MACHINE in
  Mac)
    # Mac stuff
    ;;
  Linux)
    # Linux stuff
    ;;
esac
`
		afero.WriteFile(fs, "/test.zshrc", []byte(content), 0644)

		parser := New(fs)
		result, err := parser.ParseFile("/test.zshrc")

		require.NoError(t, err)
		require.Len(t, result.Modules, 1)

		// Should infer both Mac and Linux support
		assert.Equal(t, []string{"Mac", "Linux"}, result.Modules[0].OS)
	})

	t.Run("generates manifest", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		content := `# --- Test Section ---
echo "test"
`
		afero.WriteFile(fs, "/test.zshrc", []byte(content), 0644)

		parser := New(fs)
		result, err := parser.ParseFile("/test.zshrc")

		require.NoError(t, err)
		require.NotNil(t, result.Manifest)
		assert.Len(t, result.Manifest.Modules, 1)
		assert.Equal(t, "test-section", result.Manifest.Modules[0].Name)
	})
}

func TestParser_ParseSections(t *testing.T) {
	parser := New(afero.NewMemMapFs())

	t.Run("detects dashes style header", func(t *testing.T) {
		lines := []string{
			"# --- PATH Setup ---",
			"export PATH=/usr/local/bin:$PATH",
		}

		sections := parser.parseSections(lines)

		assert.Len(t, sections, 1)
		assert.Equal(t, "PATH Setup", sections[0].Name)
		assert.Contains(t, sections[0].Content, "export PATH")
	})

	t.Run("detects equals style header", func(t *testing.T) {
		lines := []string{
			"# === Tool Init ===",
			"source nvm.sh",
		}

		sections := parser.parseSections(lines)

		assert.Len(t, sections, 1)
		assert.Equal(t, "Tool Init", sections[0].Name)
	})

	t.Run("detects hash style header", func(t *testing.T) {
		lines := []string{
			"## Aliases",
			"alias ll='ls -la'",
		}

		sections := parser.parseSections(lines)

		assert.Len(t, sections, 1)
		assert.Equal(t, "Aliases", sections[0].Name)
	})

	t.Run("detects ALL CAPS header", func(t *testing.T) {
		lines := []string{
			"# PATH CONFIGURATION",
			"export PATH",
		}

		sections := parser.parseSections(lines)

		assert.Len(t, sections, 1)
		assert.Equal(t, "PATH CONFIGURATION", sections[0].Name)
	})

	t.Run("handles multiple sections", func(t *testing.T) {
		lines := []string{
			"# --- Section 1 ---",
			"content 1",
			"# === Section 2 ===",
			"content 2",
			"## Section 3",
			"content 3",
		}

		sections := parser.parseSections(lines)

		assert.Len(t, sections, 3)
		assert.Equal(t, "Section 1", sections[0].Name)
		assert.Equal(t, "Section 2", sections[1].Name)
		assert.Equal(t, "Section 3", sections[2].Name)
	})

	t.Run("captures preamble before first section", func(t *testing.T) {
		lines := []string{
			"# Initial comment",
			"export LANG=en_US.UTF-8",
			"",
			"# --- First Section ---",
			"content",
		}

		sections := parser.parseSections(lines)

		assert.Len(t, sections, 2)
		assert.Equal(t, "Preamble", sections[0].Name)
		assert.Contains(t, sections[0].Content, "LANG")
		assert.Equal(t, "First Section", sections[1].Name)
	})

	t.Run("tracks line numbers correctly", func(t *testing.T) {
		lines := []string{
			"preamble line 1",
			"preamble line 2",
			"# --- Section ---",
			"section line 1",
			"section line 2",
		}

		sections := parser.parseSections(lines)

		require.Len(t, sections, 2)

		// Preamble should start at line 0
		assert.Equal(t, 0, sections[0].LineStart)

		// Section should start at line 3 (after header)
		assert.Equal(t, 3, sections[1].LineStart)
		assert.Equal(t, 4, sections[1].LineEnd)
	})
}

func TestParser_DetectSectionHeader(t *testing.T) {
	parser := New(afero.NewMemMapFs())

	tests := []struct {
		name  string
		line  string
		want  string
	}{
		{
			name: "dashes style",
			line: "# --- PATH Setup ---",
			want: "PATH Setup",
		},
		{
			name: "equals style",
			line: "# === Tool Initialization ===",
			want: "Tool Initialization",
		},
		{
			name: "hash style",
			line: "## My Aliases",
			want: "My Aliases",
		},
		{
			name: "all caps",
			line: "# PATH CONFIGURATION",
			want: "PATH CONFIGURATION",
		},
		{
			name: "not a header - regular comment",
			line: "# This is just a comment",
			want: "",
		},
		{
			name: "not a header - code",
			line: "export PATH=/usr/local/bin:$PATH",
			want: "",
		},
		{
			name: "not a header - empty",
			line: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.detectSectionHeader(tt.line)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParser_ExtractDescription(t *testing.T) {
	parser := New(afero.NewMemMapFs())

	t.Run("extracts single line description", func(t *testing.T) {
		lines := []string{
			"# --- Section ---",
			"# This is a description",
			"code line",
		}

		desc := parser.extractDescription(lines, 1)
		assert.Equal(t, "This is a description", desc)
	})

	t.Run("extracts multi-line description", func(t *testing.T) {
		lines := []string{
			"# --- Section ---",
			"# First line of description",
			"# Second line of description",
			"code line",
		}

		desc := parser.extractDescription(lines, 1)
		assert.Equal(t, "First line of description Second line of description", desc)
	})

	t.Run("stops at non-comment line", func(t *testing.T) {
		lines := []string{
			"# --- Section ---",
			"# Description",
			"code line",
			"# This should not be included",
		}

		desc := parser.extractDescription(lines, 1)
		assert.Equal(t, "Description", desc)
	})

	t.Run("skips section marker lines", func(t *testing.T) {
		lines := []string{
			"# --- Section ---",
			"# ---",
			"# Real description",
			"code",
		}

		desc := parser.extractDescription(lines, 1)
		assert.Equal(t, "Real description", desc)
	})

	t.Run("handles no description", func(t *testing.T) {
		lines := []string{
			"# --- Section ---",
			"code line",
		}

		desc := parser.extractDescription(lines, 1)
		assert.Empty(t, desc)
	})

	t.Run("limits to 3 lines", func(t *testing.T) {
		lines := []string{
			"# --- Section ---",
			"# Line 1",
			"# Line 2",
			"# Line 3",
			"# Line 4 should not be included",
			"code",
		}

		desc := parser.extractDescription(lines, 1)
		assert.Equal(t, "Line 1 Line 2 Line 3", desc)
	})
}
