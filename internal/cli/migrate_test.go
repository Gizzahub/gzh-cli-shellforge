package cli

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateCmd_Structure(t *testing.T) {
	cmd := newMigrateCmd()

	assert.Contains(t, cmd.Use, "migrate")
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.NotEmpty(t, cmd.Example)
	assert.NotNil(t, cmd.RunE)

	// Verify it requires exactly 1 argument
	assert.NotNil(t, cmd.Args)
}

func TestMigrateCmd_Flags(t *testing.T) {
	cmd := newMigrateCmd()

	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue string
	}{
		{
			name:         "output-dir flag",
			flagName:     "output-dir",
			shorthand:    "o",
			defaultValue: "modules",
		},
		{
			name:         "manifest flag",
			flagName:     "manifest",
			shorthand:    "m",
			defaultValue: "manifest.yaml",
		},
		{
			name:      "dry-run flag",
			flagName:  "dry-run",
			shorthand: "",
		},
		{
			name:      "verbose flag",
			flagName:  "verbose",
			shorthand: "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag, "flag %s should exist", tt.flagName)

			if tt.shorthand != "" {
				assert.Equal(t, tt.shorthand, flag.Shorthand)
			}

			if tt.defaultValue != "" {
				assert.Equal(t, tt.defaultValue, flag.DefValue)
			}
		})
	}
}

func TestMigrateCmd_Help(t *testing.T) {
	cmd := newMigrateCmd()

	// Verify help text contains key information
	assert.Contains(t, cmd.Long, "monolithic")
	assert.Contains(t, cmd.Long, "modular structure")
	assert.Contains(t, cmd.Long, "init.d")
	assert.Contains(t, cmd.Long, "rc_pre.d")
	assert.Contains(t, cmd.Long, "rc_post.d")

	// Verify examples are present
	assert.Contains(t, cmd.Example, "--dry-run")
	assert.Contains(t, cmd.Example, "--output-dir")
	assert.Contains(t, cmd.Example, "manifest")
}

func TestMigrateCmd_FlagShortcuts(t *testing.T) {
	cmd := newMigrateCmd()

	tests := []struct {
		name      string
		flagName  string
		shorthand string
	}{
		{"output-dir", "output-dir", "o"},
		{"manifest", "manifest", "m"},
		{"verbose", "verbose", "v"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := cmd.Flags().Lookup(tt.flagName)
			require.NotNil(t, flag)
			assert.Equal(t, tt.shorthand, flag.Shorthand)
		})
	}
}

func TestMigrateCmd_Examples(t *testing.T) {
	cmd := newMigrateCmd()

	// Verify examples demonstrate key features
	examples := cmd.Example
	assert.Contains(t, examples, "gz-shellforge migrate")
	assert.Contains(t, examples, ".zshrc")
	assert.Contains(t, examples, ".bashrc")
	assert.Contains(t, examples, "--dry-run")
	assert.Contains(t, examples, "--output-dir")
	assert.Contains(t, examples, "-v")
}

func TestMigrateCmd_LongDescription(t *testing.T) {
	cmd := newMigrateCmd()

	longDesc := cmd.Long

	// Check for key concepts
	assert.Contains(t, longDesc, "monolithic")
	assert.Contains(t, longDesc, "modular")
	assert.Contains(t, longDesc, "section detection")
	assert.Contains(t, longDesc, "categorization")

	// Check for directory descriptions
	assert.Contains(t, longDesc, "init.d")
	assert.Contains(t, longDesc, "rc_pre.d")
	assert.Contains(t, longDesc, "rc_post.d")

	// Check for functionality descriptions
	assert.Contains(t, longDesc, "PATH")
	assert.Contains(t, longDesc, "Tool initialization")
	assert.Contains(t, longDesc, "Aliases")
	assert.Contains(t, longDesc, "manifest.yaml")
}

func TestMigrateCmd_UsageText(t *testing.T) {
	cmd := newMigrateCmd()

	usage := cmd.Use
	assert.Equal(t, "migrate <rc-file>", usage)
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxLines int
		want     []string
	}{
		{
			name:     "single line",
			content:  "line1",
			maxLines: 3,
			want:     []string{"line1"},
		},
		{
			name:     "multiple lines under limit",
			content:  "line1\nline2\nline3",
			maxLines: 5,
			want:     []string{"line1", "line2", "line3"},
		},
		{
			name:     "truncate at max lines",
			content:  "line1\nline2\nline3\nline4\nline5",
			maxLines: 3,
			want:     []string{"line1", "line2", "line3", "..."},
		},
		{
			name:     "empty content",
			content:  "",
			maxLines: 3,
			want:     []string{},
		},
		{
			name:     "exact max lines",
			content:  "line1\nline2\nline3",
			maxLines: 3,
			want:     []string{"line1", "line2", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitLines(tt.content, tt.maxLines)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrateCmd_Integration(t *testing.T) {
	// Verify the command is properly structured for integration
	cmd := newMigrateCmd()

	// Should have RunE function
	assert.NotNil(t, cmd.RunE)

	// Should accept exactly 1 argument
	err := cmd.Args(cmd, []string{})
	assert.Error(t, err, "should require at least 1 argument")

	err = cmd.Args(cmd, []string{"file1"})
	assert.NoError(t, err, "should accept 1 argument")

	err = cmd.Args(cmd, []string{"file1", "file2"})
	assert.Error(t, err, "should not accept more than 1 argument")
}

func TestMigrateCmd_RootIntegration(t *testing.T) {
	// Verify migrate command is registered with root
	rootCmd := NewRootCmd()

	var migrateCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "migrate" {
			migrateCmd = cmd
			break
		}
	}

	require.NotNil(t, migrateCmd, "migrate command should be registered with root")
	assert.Equal(t, "migrate", migrateCmd.Name())
}

func TestPrintAnalysisResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *app.MigrateResult
		verbose  bool
		contains []string
	}{
		{
			name: "analysis with sections",
			result: &app.MigrateResult{
				SourceFile: "/home/user/.zshrc",
				Sections: []domain.Section{
					{
						Name:      "PATH Setup",
						Content:   "export PATH=/usr/local/bin:$PATH",
						Category:  "init.d",
						LineStart: 1,
						LineEnd:   3,
					},
					{
						Name:      "Aliases",
						Content:   "alias gs='git status'",
						Category:  "rc_post.d",
						LineStart: 5,
						LineEnd:   7,
					},
				},
				ModulesCreated: 2,
			},
			verbose: false,
			contains: []string{
				"Migration Analysis",
				"/home/user/.zshrc",
				"Detected sections: 2",
				"Modules to create: 2",
				"PATH Setup",
				"Aliases",
				"init.d/",
				"rc_post.d/",
			},
		},
		{
			name: "analysis with verbose",
			result: &app.MigrateResult{
				SourceFile: "/home/user/.bashrc",
				Sections: []domain.Section{
					{
						Name:      "Environment",
						Content:   "export VAR=value\nexport FOO=bar\nexport BAZ=qux",
						Category:  "init.d",
						LineStart: 1,
						LineEnd:   3,
					},
				},
				ModulesCreated: 1,
			},
			verbose: true,
			contains: []string{
				"Migration Analysis",
				"Detected sections: 1",
				"Environment",
				"Content preview:",
				"export VAR=value",
			},
		},
		{
			name: "analysis with no sections",
			result: &app.MigrateResult{
				SourceFile:     "/home/user/.zshrc",
				Sections:       []domain.Section{},
				ModulesCreated: 0,
			},
			verbose: false,
			contains: []string{
				"Migration Analysis",
				"No sections detected",
				"unsegmented",
				"Tip:",
			},
		},
		{
			name: "analysis with warnings",
			result: &app.MigrateResult{
				SourceFile: "/home/user/.zshrc",
				Sections: []domain.Section{
					{Name: "Test", Content: "echo test", Category: "init.d"},
				},
				ModulesCreated: 1,
				Warnings:       []string{"Some content could not be categorized", "Missing section headers"},
			},
			verbose: false,
			contains: []string{
				"Migration Analysis",
				"Warnings:",
				"Some content could not be categorized",
				"Missing section headers",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printAnalysisResult(tt.result, tt.verbose)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}

func TestPrintMigrationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   *app.MigrateResult
		verbose  bool
		contains []string
	}{
		{
			name: "migration complete",
			result: &app.MigrateResult{
				SourceFile:     "/home/user/.zshrc",
				ModulesCreated: 3,
				ManifestPath:   "/output/manifest.yaml",
			},
			verbose: false,
			contains: []string{
				"Migration Complete",
				"Created 3 module files",
				"Generated manifest: /output/manifest.yaml",
				"Next steps:",
				"Review generated module files",
			},
		},
		{
			name: "migration with verbose",
			result: &app.MigrateResult{
				SourceFile:     "/home/user/.bashrc",
				ModulesCreated: 2,
				ManifestPath:   "/output/manifest.yaml",
				ModuleFilesPaths: []string{
					"/output/modules/init.d/10-path.sh",
					"/output/modules/rc_post.d/50-aliases.sh",
				},
			},
			verbose: true,
			contains: []string{
				"Migration Complete",
				"Created 2 module files",
				"Module files created:",
				"/output/modules/init.d/10-path.sh",
				"/output/modules/rc_post.d/50-aliases.sh",
			},
		},
		{
			name: "migration with warnings",
			result: &app.MigrateResult{
				SourceFile:     "/home/user/.zshrc",
				ModulesCreated: 1,
				ManifestPath:   "/output/manifest.yaml",
				Warnings:       []string{"Some lines could not be parsed", "Complex shell constructs detected"},
			},
			verbose: false,
			contains: []string{
				"Migration Complete",
				"Warnings:",
				"Some lines could not be parsed",
				"Complex shell constructs detected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			printMigrationResult(tt.result, tt.verbose)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected, "Output should contain: %s", expected)
			}
		})
	}
}
