package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule_AppliesTo(t *testing.T) {
	tests := []struct {
		name     string
		module   Module
		targetOS string
		expected bool
	}{
		{
			name:     "matches single OS",
			module:   Module{OS: []string{"Mac"}},
			targetOS: "Mac",
			expected: true,
		},
		{
			name:     "matches one of multiple OS",
			module:   Module{OS: []string{"Mac", "Linux"}},
			targetOS: "Linux",
			expected: true,
		},
		{
			name:     "no match",
			module:   Module{OS: []string{"Mac"}},
			targetOS: "Linux",
			expected: false,
		},
		{
			name:     "empty OS list applies to all",
			module:   Module{OS: []string{}},
			targetOS: "Mac",
			expected: true,
		},
		{
			name:     "case insensitive match",
			module:   Module{OS: []string{"Mac"}},
			targetOS: "mac",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.AppliesTo(tt.targetOS)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		module  Module
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid module",
			module: Module{
				Name: "test",
				File: "test.sh",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			module: Module{
				File: "test.sh",
			},
			wantErr: true,
			errMsg:  "missing 'name' field",
		},
		{
			name: "missing file",
			module: Module{
				Name: "test",
			},
			wantErr: true,
			errMsg:  "missing 'file' field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.module.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModule_GetTarget(t *testing.T) {
	tests := []struct {
		name     string
		module   Module
		expected string
	}{
		{
			name:     "returns explicit target",
			module:   Module{Target: "zprofile"},
			expected: "zprofile",
		},
		{
			name:     "defaults to zshrc when empty",
			module:   Module{Target: ""},
			expected: "zshrc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.GetTarget()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestModule_GetPriority(t *testing.T) {
	tests := []struct {
		name     string
		module   Module
		expected int
	}{
		{
			name:     "returns explicit priority",
			module:   Module{Priority: 10},
			expected: 10,
		},
		{
			name:     "defaults to 50 when zero",
			module:   Module{Priority: 0},
			expected: 50,
		},
		{
			name:     "returns negative priority",
			module:   Module{Priority: -5},
			expected: -5,
		},
		{
			name:     "returns high priority",
			module:   Module{Priority: 100},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.module.GetPriority()
			assert.Equal(t, tt.expected, result)
		})
	}
}
