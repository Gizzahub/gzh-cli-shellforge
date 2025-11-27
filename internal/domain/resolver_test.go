package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_TopologicalSort(t *testing.T) {
	tests := []struct {
		name     string
		modules  []Module
		targetOS string
		expected []string // module names in expected order
		wantErr  bool
	}{
		{
			name: "linear dependencies",
			modules: []Module{
				{Name: "a", File: "a.sh", Requires: []string{}, OS: []string{"Mac"}},
				{Name: "b", File: "b.sh", Requires: []string{"a"}, OS: []string{"Mac"}},
				{Name: "c", File: "c.sh", Requires: []string{"b"}, OS: []string{"Mac"}},
			},
			targetOS: "Mac",
			expected: []string{"a", "b", "c"},
			wantErr:  false,
		},
		{
			name: "complex DAG",
			modules: []Module{
				{Name: "base", File: "base.sh", Requires: []string{}, OS: []string{"Mac"}},
				{Name: "tool1", File: "tool1.sh", Requires: []string{"base"}, OS: []string{"Mac"}},
				{Name: "tool2", File: "tool2.sh", Requires: []string{"base"}, OS: []string{"Mac"}},
				{Name: "config", File: "config.sh", Requires: []string{"tool1", "tool2"}, OS: []string{"Mac"}},
			},
			targetOS: "Mac",
			expected: []string{"base", "tool1", "tool2", "config"}, // tool1/tool2 order may vary
			wantErr:  false,
		},
		{
			name: "OS filtering",
			modules: []Module{
				{Name: "base", File: "base.sh", Requires: []string{}, OS: []string{"Mac", "Linux"}},
				{Name: "brew", File: "brew.sh", Requires: []string{"base"}, OS: []string{"Mac"}},
				{Name: "pacman", File: "pacman.sh", Requires: []string{"base"}, OS: []string{"Linux"}},
			},
			targetOS: "Mac",
			expected: []string{"base", "brew"},
			wantErr:  false,
		},
		{
			name: "circular dependency",
			modules: []Module{
				{Name: "a", File: "a.sh", Requires: []string{"b"}, OS: []string{"Mac"}},
				{Name: "b", File: "b.sh", Requires: []string{"c"}, OS: []string{"Mac"}},
				{Name: "c", File: "c.sh", Requires: []string{"a"}, OS: []string{"Mac"}},
			},
			targetOS: "Mac",
			wantErr:  true,
		},
		{
			name: "no dependencies",
			modules: []Module{
				{Name: "a", File: "a.sh", Requires: []string{}, OS: []string{"Mac"}},
				{Name: "b", File: "b.sh", Requires: []string{}, OS: []string{"Mac"}},
			},
			targetOS: "Mac",
			expected: []string{"a", "b"}, // order may vary
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := &Manifest{Modules: tt.modules}
			resolver := NewResolver()

			graph, err := resolver.BuildGraph(manifest)
			require.NoError(t, err)

			result, err := resolver.TopologicalSort(graph, tt.targetOS)

			if tt.wantErr {
				assert.Error(t, err)
				_, ok := err.(*CircularDependencyError)
				assert.True(t, ok, "expected CircularDependencyError")
			} else {
				require.NoError(t, err)
				assert.Len(t, result, len(tt.expected))

				// Check that all expected modules are present
				resultNames := make([]string, len(result))
				for i, mod := range result {
					resultNames[i] = mod.Name
				}

				for _, expectedName := range tt.expected {
					assert.Contains(t, resultNames, expectedName)
				}

				// Verify dependency order (if a depends on b, b must come before a)
				moduleIndex := make(map[string]int)
				for i, mod := range result {
					moduleIndex[mod.Name] = i
				}

				for _, mod := range result {
					for _, dep := range mod.Requires {
						if depIndex, exists := moduleIndex[dep]; exists {
							assert.Less(t, depIndex, moduleIndex[mod.Name],
								"dependency %s must come before %s", dep, mod.Name)
						}
					}
				}
			}
		})
	}
}

func TestResolver_BuildGraph(t *testing.T) {
	tests := []struct {
		name    string
		modules []Module
		wantErr bool
	}{
		{
			name: "valid graph",
			modules: []Module{
				{Name: "a", File: "a.sh", Requires: []string{}},
				{Name: "b", File: "b.sh", Requires: []string{"a"}},
			},
			wantErr: false,
		},
		{
			name: "non-existent dependency",
			modules: []Module{
				{Name: "a", File: "a.sh", Requires: []string{"missing"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := &Manifest{Modules: tt.modules}
			resolver := NewResolver()

			graph, err := resolver.BuildGraph(manifest)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, graph)
				assert.Equal(t, len(tt.modules), graph.Size())
			}
		})
	}
}
