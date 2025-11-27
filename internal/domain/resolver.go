package domain

import "strings"

// Resolver handles dependency resolution and topological sorting.
type Resolver struct{}

// NewResolver creates a new resolver.
func NewResolver() *Resolver {
	return &Resolver{}
}

// BuildGraph creates a dependency graph from a manifest.
func (r *Resolver) BuildGraph(manifest *Manifest) (*Graph, error) {
	graph := NewGraph()

	// Add all modules as nodes
	for i := range manifest.Modules {
		graph.AddNode(&manifest.Modules[i])
	}

	// Add edges for dependencies
	for _, module := range manifest.Modules {
		for _, dep := range module.Requires {
			if err := graph.AddEdge(dep, module.Name); err != nil {
				return nil, err
			}
		}
	}

	return graph, nil
}

// TopologicalSort performs Kahn's algorithm to sort modules by dependencies.
// Only includes modules that apply to the target OS.
func (r *Resolver) TopologicalSort(graph *Graph, targetOS string) ([]Module, error) {
	// Create working copy of in-degrees, filtering by OS
	inDegree := make(map[string]int)
	for _, name := range graph.GetAllNodes() {
		node, _ := graph.GetNode(name)
		if node.Module.AppliesTo(targetOS) {
			inDegree[name] = node.InDegree
		}
	}

	// Find all nodes with in-degree 0
	queue := []string{}
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Process queue (Kahn's algorithm)
	var result []Module
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		node, _ := graph.GetNode(current)
		result = append(result, *node.Module)

		// Process dependents
		for _, dependent := range graph.GetDependents(current) {
			if _, ok := inDegree[dependent]; !ok {
				continue // Skip filtered modules
			}

			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(result) != len(inDegree) {
		return nil, r.detectCycle(graph, inDegree)
	}

	return result, nil
}

// detectCycle finds and reports a circular dependency.
func (r *Resolver) detectCycle(graph *Graph, inDegree map[string]int) error {
	// Find nodes still in graph (part of cycle)
	var cycleNodes []string
	for name, degree := range inDegree {
		if degree > 0 {
			cycleNodes = append(cycleNodes, name)
		}
	}

	// Build cycle path (simplified - just show nodes in cycle)
	cyclePath := strings.Join(cycleNodes, " → ")
	return NewCircularDependencyError("circular dependency detected: %s", cyclePath)
}
