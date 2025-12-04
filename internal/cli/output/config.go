package output

import (
	"fmt"
	"strings"
)

// ConfigPrinter handles verbose configuration output
type ConfigPrinter struct {
	title  string
	fields []configField
}

type configField struct {
	label string
	value interface{}
}

// NewConfigPrinter creates a new configuration printer
func NewConfigPrinter(title string) *ConfigPrinter {
	return &ConfigPrinter{
		title:  title,
		fields: make([]configField, 0),
	}
}

// Add adds a field to the configuration output
func (p *ConfigPrinter) Add(label string, value interface{}) *ConfigPrinter {
	p.fields = append(p.fields, configField{label: label, value: value})
	return p
}

// AddIf adds a field only if condition is true
func (p *ConfigPrinter) AddIf(condition bool, label string, value interface{}) *ConfigPrinter {
	if condition {
		p.fields = append(p.fields, configField{label: label, value: value})
	}
	return p
}

// Print outputs the configuration if verbose is true
func (p *ConfigPrinter) Print(verbose bool) {
	if !verbose {
		return
	}

	fmt.Printf("%s:\n", p.title)
	for _, f := range p.fields {
		fmt.Printf("  %s: %v\n", f.label, f.value)
	}
	fmt.Println()
}

// PrintVerboseHeader prints a header line if verbose mode is enabled
func PrintVerboseHeader(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// PrintKeyValue prints a key-value pair with consistent formatting
func PrintKeyValue(key string, value interface{}) {
	fmt.Printf("  %s: %v\n", key, value)
}

// PrintKeyValues prints multiple key-value pairs
func PrintKeyValues(pairs map[string]interface{}) {
	// Find the longest key for alignment
	maxLen := 0
	for k := range pairs {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}

	for k, v := range pairs {
		padding := strings.Repeat(" ", maxLen-len(k))
		fmt.Printf("  %s%s %v\n", k+":", padding, v)
	}
}
