package output

import (
	"fmt"
)

// SuccessResult prints a success message with optional details
func SuccessResult(message string) {
	fmt.Printf("✓ %s\n", message)
}

// DryRunNotice prints a dry run notice
func DryRunNotice() {
	fmt.Printf("🔍 Dry run - no changes made\n\n")
}

// SnapshotInfo prints snapshot information in a consistent format
type SnapshotInfo struct {
	Timestamp    string
	Size         string
	Location     string
	Target       string
	GitStatus    string
	ShowGit      bool
	GitCommitted bool
}

// Print outputs the snapshot information
func (s *SnapshotInfo) Print() {
	fmt.Printf("Snapshot:\n")
	fmt.Printf("  Timestamp: %s\n", s.Timestamp)
	fmt.Printf("  Size:      %s\n", s.Size)
	if s.Location != "" {
		fmt.Printf("  Location:  %s\n", s.Location)
	}
	if s.Target != "" {
		fmt.Printf("  Target:    %s\n", s.Target)
	}
	if s.ShowGit {
		if s.GitCommitted {
			fmt.Printf("  Git:       Committed\n")
		} else {
			fmt.Printf("  Git:       Not committed (see details below)\n")
		}
	}
}

// PrintDetails prints verbose details section
func PrintDetails(verbose bool, message string) {
	if !verbose || message == "" {
		return
	}
	fmt.Printf("\nDetails:\n")
	fmt.Printf("  %s\n", message)
}

// PrintNextSteps prints next steps or commands to run
func PrintNextSteps(steps []string) {
	if len(steps) == 0 {
		return
	}
	fmt.Println("\nNext steps:")
	for _, step := range steps {
		fmt.Printf("  %s\n", step)
	}
}

// PrintRestoreHint prints the restore command hint
func PrintRestoreHint(filePath, timestamp string) {
	fmt.Printf("\nTo restore this backup:\n")
	fmt.Printf("  gz-shellforge restore --file %s --snapshot %s\n", filePath, timestamp)
}

// PrintApplyHint prints command to apply changes (for dry-run mode)
func PrintApplyHint(command string) {
	fmt.Printf("\nTo apply this change:\n")
	fmt.Printf("  %s\n", command)
}

// Summary prints a summary section
type Summary struct {
	fields []summaryField
}

type summaryField struct {
	label string
	value interface{}
}

// NewSummary creates a new summary printer
func NewSummary() *Summary {
	return &Summary{fields: make([]summaryField, 0)}
}

// Add adds a field to the summary
func (s *Summary) Add(label string, value interface{}) *Summary {
	s.fields = append(s.fields, summaryField{label: label, value: value})
	return s
}

// Print outputs the summary
func (s *Summary) Print() {
	if len(s.fields) == 0 {
		return
	}
	fmt.Printf("\nSummary:\n")
	for _, f := range s.fields {
		fmt.Printf("  %s: %v\n", f.label, f.value)
	}
}
