package domain

// Graph represents a directed dependency graph.
type Graph struct {
	nodes map[string]*Node
	edges map[string][]string // node -> list of dependents
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

// GetAllNodes returns all node names in the graph.
func (g *Graph) GetAllNodes() []string {
	names := make([]string, 0, len(g.nodes))
	for name := range g.nodes {
		names = append(names, name)
	}
	return names
}
