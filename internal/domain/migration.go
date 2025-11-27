package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Section represents a parsed section from an RC file
type Section struct {
	Name        string           // Section name (extracted from header comment)
	Content     string           // Raw section content (without header comment)
	Category    TemplateCategory // Target category (init.d, rc_pre.d, rc_post.d)
	LineStart   int              // Starting line number
	LineEnd     int              // Ending line number
	Description string           // Optional description for manifest
}

// MigrationResult contains the result of migrating an RC file
type MigrationResult struct {
	Sections []Section         // Detected sections
	Modules  []Module          // Generated module definitions
	Manifest *Manifest         // Generated manifest
	Warnings []string          // Warnings during migration
}

// SectionPattern defines regex patterns for detecting section headers
// Matches: # --- Name ---, # === Name ===, ## Name
var SectionPattern = regexp.MustCompile(`(?m)^(?:#\s*(?:---|===)\s*(.+?)\s*(?:---|===)|##\s+(.+?))\s*$`)

// AllCapsPattern detects ALL CAPS section headers
var AllCapsPattern = regexp.MustCompile(`(?m)^#\s*([A-Z][A-Z\s]{3,})\s*$`)

// CategorizeSection determines the category based on content analysis
// Content patterns take priority over name patterns
func CategorizeSection(name, content string) TemplateCategory {
	nameLower := strings.ToLower(name)
	contentLower := strings.ToLower(content)

	// Preamble is always first (goes to init.d)
	if strings.Contains(nameLower, "preamble") {
		return CategoryInitD
	}

	// Priority 1: Check content patterns (most reliable)

	// PATH manipulation in content → init.d
	if containsAny(contentLower, []string{"export path=", "path=", "$path:"}) {
		return CategoryInitD
	}

	// Aliases/functions in content → rc_post.d
	if containsAny(contentLower, []string{"alias ", "function ", "prompt"}) {
		return CategoryRcPostD
	}

	// Tool initialization in content → rc_pre.d
	if containsAny(contentLower, []string{"nvm", "rbenv", "pyenv", "conda", "asdf", "sdk", "nvm_dir"}) {
		return CategoryRcPreD
	}

	// Priority 2: Check name patterns (less reliable, more generic)

	// Specific tool names in section name → rc_pre.d
	if containsAny(nameLower, []string{"nvm", "rbenv", "pyenv", "conda", "asdf", "tool"}) {
		return CategoryRcPreD
	}

	// Alias/function sections → rc_post.d
	if containsAny(nameLower, []string{"alias", "function", "prompt", "custom"}) {
		return CategoryRcPostD
	}

	// Init/path/setup sections → init.d
	if containsAny(nameLower, []string{"path", "initialization", "setup", "init"}) {
		return CategoryInitD
	}

	// Default: rc_pre.d is the safest default
	return CategoryRcPreD
}

// InferDependencies analyzes content and infers required dependencies
func InferDependencies(content string) []string {
	deps := make([]string, 0)
	seen := make(map[string]bool)

	// Check for common dependency patterns
	if strings.Contains(content, "$MACHINE") || strings.Contains(content, "${MACHINE}") {
		if !seen["os-detection"] {
			deps = append(deps, "os-detection")
			seen["os-detection"] = true
		}
	}

	if strings.Contains(content, "brew ") || strings.Contains(content, "$(brew ") {
		if !seen["brew-path"] {
			deps = append(deps, "brew-path")
			seen["brew-path"] = true
		}
	}

	return deps
}

// InferOSSupport analyzes content for OS-specific patterns
func InferOSSupport(content string) []string {
	// Check for case $MACHINE pattern
	if strings.Contains(content, "case $MACHINE") || strings.Contains(content, "case \"$MACHINE\"") {
		// Look for Mac and Linux branches
		hasMac := strings.Contains(content, "Mac)")
		hasLinux := strings.Contains(content, "Linux)")

		if hasMac && hasLinux {
			return []string{"Mac", "Linux"}
		}
		if hasMac {
			return []string{"Mac"}
		}
		if hasLinux {
			return []string{"Linux"}
		}
	}

	// Default: works on all platforms
	return []string{"Mac", "Linux"}
}

// GenerateModuleName creates a valid module name from section name
func GenerateModuleName(sectionName string, index int) string {
	// Convert to lowercase and replace spaces/special chars with hyphens
	name := strings.ToLower(sectionName)
	name = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`-+`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	// If empty, generate a generic name
	if name == "" {
		name = fmt.Sprintf("section-%d", index)
	}

	return name
}

// GenerateFileName creates a filename for the module
func GenerateFileName(category TemplateCategory, moduleName string, index int) string {
	// Use index prefix for init.d to control load order
	if category == CategoryInitD {
		if moduleName == "preamble" {
			return "init.d/00-preamble.sh"
		}
		// Start from 10 to leave room for standard modules
		return fmt.Sprintf("init.d/%02d-%s.sh", (index+1)*10, moduleName)
	}

	return fmt.Sprintf("%s/%s.sh", category, moduleName)
}

// containsAny checks if the text contains any of the substrings
func containsAny(text string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(text, substr) {
			return true
		}
	}
	return false
}

// NewMigrationResult creates a new migration result
func NewMigrationResult() *MigrationResult {
	return &MigrationResult{
		Sections: make([]Section, 0),
		Modules:  make([]Module, 0),
		Warnings: make([]string, 0),
	}
}

// AddWarning adds a warning to the migration result
func (m *MigrationResult) AddWarning(format string, args ...interface{}) {
	m.Warnings = append(m.Warnings, fmt.Sprintf(format, args...))
}

// GenerateManifest creates a manifest from the migration result
func (m *MigrationResult) GenerateManifest() *Manifest {
	manifest := &Manifest{
		Modules: m.Modules,
	}
	m.Manifest = manifest
	return manifest
}
