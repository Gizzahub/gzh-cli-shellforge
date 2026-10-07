// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package domain

// Graph represents a directed dependency graph.
type Graph struct {
	nodes map[string]*Node
	edges map[string][]string // node -> list of dependents
	order []string            // node names in insertion (manifest) order
}

// Node represents a node in the dependency graph.
type Node struct {
	Module   *Module
	InDegree int
}

// NewGraph creates a new empty graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make(map[string][]string),
	}
}

// AddNode adds a module as a node in the graph.
func (g *Graph) AddNode(module *Module) {
	if _, exists := g.nodes[module.Name]; !exists {
		g.order = append(g.order, module.Name)
	}
	g.nodes[module.Name] = &Node{
		Module:   module,
		InDegree: 0,
	}
}

// AddEdge adds a directed edge from dependency to dependent.
// Returns an error if either node doesn't exist.
func (g *Graph) AddEdge(from, to string) error {
	if _, exists := g.nodes[from]; !exists {
		return NewValidationError("dependency '%s' not found", from)
	}
	if _, exists := g.nodes[to]; !exists {
		return NewValidationError("module '%s' not found", to)
	}

	g.edges[from] = append(g.edges[from], to)
	g.nodes[to].InDegree++
	return nil
}

// Size returns the number of nodes in the graph.
func (g *Graph) Size() int {
	return len(g.nodes)
}

// GetNode returns a node by name.
func (g *Graph) GetNode(name string) (*Node, bool) {
	node, exists := g.nodes[name]
	return node, exists
}

// GetDependents returns the list of modules that depend on the given module.
func (g *Graph) GetDependents(name string) []string {
	return g.edges[name]
}

// GetAllNodes returns all node names in insertion order, so callers that walk
// the graph produce the same output on every run (map order is randomized).
func (g *Graph) GetAllNodes() []string {
	return append([]string(nil), g.order...)
}
