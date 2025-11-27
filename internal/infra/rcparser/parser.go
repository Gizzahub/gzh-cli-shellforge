package rcparser

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/spf13/afero"
)

// Parser parses RC files and extracts sections
type Parser struct {
	fs afero.Fs
}

// New creates a new RC file parser
func New(fs afero.Fs) *Parser {
	return &Parser{fs: fs}
}

// ParseFile reads and parses an RC file into sections
func (p *Parser) ParseFile(path string) (*domain.MigrationResult, error) {
	// Read file
	file, err := p.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read all lines
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse sections
	result := domain.NewMigrationResult()
	sections := p.parseSections(lines)

	// Process each section
	for i, section := range sections {
		// Generate module name and filename
		moduleName := domain.GenerateModuleName(section.Name, i)
		fileName := domain.GenerateFileName(section.Category, moduleName, i)

		// Infer dependencies and OS support
		requires := domain.InferDependencies(section.Content)
		osSupport := domain.InferOSSupport(section.Content)

		// Create module
		module := domain.Module{
			Name:        moduleName,
			File:        fileName,
			Requires:    requires,
			OS:          osSupport,
			Description: section.Description,
		}

		result.Sections = append(result.Sections, section)
		result.Modules = append(result.Modules, module)
	}

	// Generate manifest
	result.GenerateManifest()

	return result, nil
}

// parseSections extracts sections from lines
func (p *Parser) parseSections(lines []string) []domain.Section {
	var sections []domain.Section
	var currentSection *domain.Section
	var preambleContent strings.Builder
	foundFirstSection := false

	for lineNum, line := range lines {
		// Check for section headers (dashes, equals, hash, or all caps)
		if sectionName := p.detectSectionHeader(line); sectionName != "" {
			// Save previous section if exists
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(currentSection.Content)
				currentSection.LineEnd = lineNum - 1
				sections = append(sections, *currentSection)
			}

			// Start new section
			foundFirstSection = true
			category := domain.CategorizeSection(sectionName, "")
			currentSection = &domain.Section{
				Name:      sectionName,
				Content:   "",
				Category:  category,
				LineStart: lineNum + 1,
			}

			// Extract description from following comment lines
			currentSection.Description = p.extractDescription(lines, lineNum+1)
		} else {
			// Add content to current section or preamble
			if !foundFirstSection {
				// Before first section = preamble
				preambleContent.WriteString(line)
				preambleContent.WriteString("\n")
			} else if currentSection != nil {
				// Skip empty lines at start of section
				if currentSection.Content == "" && strings.TrimSpace(line) == "" {
					continue
				}
				currentSection.Content += line + "\n"
			}
		}
	}

	// Save last section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(currentSection.Content)
		currentSection.LineEnd = len(lines) - 1
		sections = append(sections, *currentSection)
	}

	// Add preamble if exists
	preamble := strings.TrimSpace(preambleContent.String())
	if preamble != "" {
		preambleSection := domain.Section{
			Name:        "Preamble",
			Content:     preamble,
			Category:    domain.CategoryInitD,
			LineStart:   0,
			LineEnd:     0,
			Description: "Initial configuration before first section",
		}
		// Re-categorize based on actual content
		preambleSection.Category = domain.CategorizeSection(preambleSection.Name, preambleSection.Content)
		// Prepend preamble
		sections = append([]domain.Section{preambleSection}, sections...)
	}

	return sections
}

// detectSectionHeader checks if a line is a section header and returns the section name
func (p *Parser) detectSectionHeader(line string) string {
	// Try standard patterns (---, ===, ##)
	matches := domain.SectionPattern.FindStringSubmatch(line)
	if matches != nil {
		// Check both capture groups
		if len(matches) > 1 && matches[1] != "" {
			return strings.TrimSpace(matches[1])
		}
		if len(matches) > 2 && matches[2] != "" {
			return strings.TrimSpace(matches[2])
		}
	}

	// Try ALL CAPS pattern
	matches = domain.AllCapsPattern.FindStringSubmatch(line)
	if matches != nil && len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractDescription extracts description from comment lines following a section header
func (p *Parser) extractDescription(lines []string, startLine int) string {
	var desc strings.Builder
	maxLines := 3 // Only check first 3 lines for description

	for i := 0; i < maxLines && startLine+i < len(lines); i++ {
		line := strings.TrimSpace(lines[startLine+i])

		// Stop at non-comment or empty line
		if !strings.HasPrefix(line, "#") || line == "#" {
			break
		}

		// Extract comment text
		comment := strings.TrimPrefix(line, "#")
		comment = strings.TrimSpace(comment)

		// Skip lines that look like section markers
		if strings.HasPrefix(comment, "---") || strings.HasPrefix(comment, "===") {
			continue
		}

		// Add to description
		if desc.Len() > 0 {
			desc.WriteString(" ")
		}
		desc.WriteString(comment)
	}

	return desc.String()
}
