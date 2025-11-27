package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
)

// RCFileReader defines the interface for reading RC files
type RCFileReader interface {
	ReadFile(path string) (string, error)
}

// MigrationService handles migration of monolithic RC files to modular structure
type MigrationService struct {
	reader FileReader
	writer FileWriter
}

// NewMigrationService creates a new migration service
func NewMigrationService(reader FileReader, writer FileWriter) *MigrationService {
	return &MigrationService{
		reader: reader,
		writer: writer,
	}
}

// MigrateResult contains the result of an RC file migration
type MigrateResult struct {
	SourceFile      string
	Sections        []domain.Section
	ModulesCreated  int
	ManifestPath    string
	Warnings        []string
	ModuleFilesPaths []string
}

// Analyze parses an RC file and returns detected sections without creating files
func (s *MigrationService) Analyze(rcFilePath string) (*MigrateResult, error) {
	content, err := s.reader.ReadFile(rcFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read RC file: %w", err)
	}

	result := domain.NewMigrationResult()
	sections := s.extractSections(content)
	result.Sections = sections

	// Generate modules (but don't write files in analyze mode)
	for i, section := range sections {
		moduleName := domain.GenerateModuleName(section.Name, i)
		fileName := domain.GenerateFileName(section.Category, moduleName, i)

		deps := domain.InferDependencies(section.Content)
		osSupport := domain.InferOSSupport(section.Content)

		module := domain.Module{
			Name:        moduleName,
			File:        fileName,
			Requires:    deps,
			OS:          osSupport,
			Description: section.Description,
		}

		result.Modules = append(result.Modules, module)
	}

	result.GenerateManifest()

	migrateResult := &MigrateResult{
		SourceFile:     rcFilePath,
		Sections:       sections,
		ModulesCreated: len(result.Modules),
		Warnings:       result.Warnings,
	}

	return migrateResult, nil
}

// Migrate performs the full migration: analyze, create module files, and generate manifest
func (s *MigrationService) Migrate(rcFilePath, outputDir, manifestPath string) (*MigrateResult, error) {
	content, err := s.reader.ReadFile(rcFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read RC file: %w", err)
	}

	result := domain.NewMigrationResult()
	sections := s.extractSections(content)
	result.Sections = sections

	if len(sections) == 0 {
		result.AddWarning("no sections detected in RC file - file may be unsegmented")
	}

	// Track created file paths
	modulePaths := make([]string, 0, len(sections))

	// Create module files
	for i, section := range sections {
		moduleName := domain.GenerateModuleName(section.Name, i)
		fileName := domain.GenerateFileName(section.Category, moduleName, i)

		deps := domain.InferDependencies(section.Content)
		osSupport := domain.InferOSSupport(section.Content)

		module := domain.Module{
			Name:        moduleName,
			File:        fileName,
			Requires:    deps,
			OS:          osSupport,
			Description: section.Description,
		}

		result.Modules = append(result.Modules, module)

		// Write module file
		fullPath := filepath.Join(outputDir, fileName)
		moduleContent := s.generateModuleContent(module, section.Content)

		if err := s.writer.WriteFile(fullPath, moduleContent); err != nil {
			return nil, fmt.Errorf("failed to write module file %s: %w", fullPath, err)
		}

		modulePaths = append(modulePaths, fullPath)
	}

	// Generate and write manifest
	result.GenerateManifest()
	manifestContent := s.generateManifestYAML(result.Manifest)

	if err := s.writer.WriteFile(manifestPath, manifestContent); err != nil {
		return nil, fmt.Errorf("failed to write manifest file: %w", err)
	}

	migrateResult := &MigrateResult{
		SourceFile:       rcFilePath,
		Sections:         sections,
		ModulesCreated:   len(result.Modules),
		ManifestPath:     manifestPath,
		Warnings:         result.Warnings,
		ModuleFilesPaths: modulePaths,
	}

	return migrateResult, nil
}

// extractSections parses the RC file content and extracts sections
func (s *MigrationService) extractSections(content string) []domain.Section {
	sections := make([]domain.Section, 0)
	lines := strings.Split(content, "\n")

	// First, try to find explicit section headers
	matches := domain.SectionPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		// No section headers found, try ALL CAPS pattern
		matches = domain.AllCapsPattern.FindAllStringIndex(content, -1)
	}

	if len(matches) == 0 {
		// No sections detected, treat entire file as one section
		section := domain.Section{
			Name:      "main-config",
			Content:   content,
			Category:  domain.CategorizeSection("main", content),
			LineStart: 1,
			LineEnd:   len(lines),
		}
		sections = append(sections, section)
		return sections
	}

	// Extract section names from matches
	sectionHeaders := make([]struct {
		name      string
		lineIndex int
	}, 0)

	for _, match := range matches {
		sectionText := content[match[0]:match[1]]

		// Try SectionPattern first
		submatches := domain.SectionPattern.FindStringSubmatch(sectionText)
		name := ""
		if len(submatches) > 1 && submatches[1] != "" {
			name = strings.TrimSpace(submatches[1])
		} else if len(submatches) > 2 && submatches[2] != "" {
			name = strings.TrimSpace(submatches[2])
		}

		// If SectionPattern didn't match, try AllCapsPattern
		if name == "" {
			submatches = domain.AllCapsPattern.FindStringSubmatch(sectionText)
			if len(submatches) > 1 {
				name = strings.TrimSpace(submatches[1])
			}
		}

		if name != "" {
			// Find line number for this match
			lineIndex := strings.Count(content[:match[0]], "\n")
			sectionHeaders = append(sectionHeaders, struct {
				name      string
				lineIndex int
			}{name: name, lineIndex: lineIndex})
		}
	}

	// Extract content between section headers
	for i, header := range sectionHeaders {
		startLine := header.lineIndex + 1 // Skip the header line
		endLine := len(lines)

		if i+1 < len(sectionHeaders) {
			endLine = sectionHeaders[i+1].lineIndex
		}

		// Extract content
		sectionLines := lines[startLine:endLine]
		sectionContent := strings.Join(sectionLines, "\n")
		sectionContent = strings.TrimSpace(sectionContent)

		if sectionContent == "" {
			continue // Skip empty sections
		}

		section := domain.Section{
			Name:      header.name,
			Content:   sectionContent,
			Category:  domain.CategorizeSection(header.name, sectionContent),
			LineStart: startLine + 1,
			LineEnd:   endLine,
		}

		sections = append(sections, section)
	}

	return sections
}

// generateModuleContent creates the module file content with header
func (s *MigrationService) generateModuleContent(module domain.Module, content string) string {
	var sb strings.Builder

	// Module header
	sb.WriteString("#!/bin/bash\n")
	sb.WriteString(fmt.Sprintf("# Module: %s\n", module.Name))

	if module.Description != "" {
		sb.WriteString(fmt.Sprintf("# Description: %s\n", module.Description))
	}

	if len(module.Requires) > 0 {
		sb.WriteString(fmt.Sprintf("# Requires: %s\n", strings.Join(module.Requires, ", ")))
	}

	if len(module.OS) > 0 {
		sb.WriteString(fmt.Sprintf("# OS: %s\n", strings.Join(module.OS, ", ")))
	}

	sb.WriteString("\n")
	sb.WriteString(content)

	// Ensure file ends with newline
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}

	return sb.String()
}

// generateManifestYAML creates the YAML manifest content
func (s *MigrationService) generateManifestYAML(manifest *domain.Manifest) string {
	var sb strings.Builder

	sb.WriteString("# Generated by gz-shellforge migrate\n")
	sb.WriteString("modules:\n")

	for _, module := range manifest.Modules {
		sb.WriteString(fmt.Sprintf("  - name: %s\n", module.Name))
		sb.WriteString(fmt.Sprintf("    file: %s\n", module.File))

		if len(module.Requires) > 0 {
			sb.WriteString("    requires:\n")
			for _, dep := range module.Requires {
				sb.WriteString(fmt.Sprintf("      - %s\n", dep))
			}
		} else {
			sb.WriteString("    requires: []\n")
		}

		if len(module.OS) > 0 {
			sb.WriteString("    os:\n")
			for _, os := range module.OS {
				sb.WriteString(fmt.Sprintf("      - %s\n", os))
			}
		}

		if module.Description != "" {
			sb.WriteString(fmt.Sprintf("    description: %s\n", module.Description))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
