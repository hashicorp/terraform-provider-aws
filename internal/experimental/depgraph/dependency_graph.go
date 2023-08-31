// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package depgraph

import (
	"fmt"
	"strings"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"golang.org/x/exp/slices"
)

// Graph implements a simple dependency graph.
type Graph struct {
	nodes         []string
	outgoingEdges map[string][]string
	incomingEdges map[string][]string
}

// New returns a new, empty dependency graph.
func New() *Graph {
	return &Graph{
		nodes:         make([]string, 0),
		outgoingEdges: make(map[string][]string),
		incomingEdges: make(map[string][]string),
	}
}

// Len returns the number of nodes in the graph.
func (g *Graph) Len() int {
	return len(g.nodes)
}

// AddNode adds the specified string to the graph.
func (g *Graph) AddNode(s string) {
	if !g.HasNode(s) {
		g.nodes = append(g.nodes, s)
		g.outgoingEdges[s] = make([]string, 0)
		g.incomingEdges[s] = make([]string, 0)
	}
}

// RemoveNode removes the specified string from the graph if it is present.
func (g *Graph) RemoveNode(s string) {
	if g.HasNode(s) {
		for n, edges := range g.outgoingEdges {
			g.outgoingEdges[n] = tfslices.RemoveAll(edges, s)
		}

		for n, edges := range g.incomingEdges {
			g.incomingEdges[n] = tfslices.RemoveAll(edges, s)
		}

		g.nodes = tfslices.RemoveAll(g.nodes, s)
		delete(g.outgoingEdges, s)
		delete(g.incomingEdges, s)
	}
}

// HasNode returns whether the specified string is in the graph.
func (g *Graph) HasNode(s string) bool {
	return slices.Contains(g.nodes, s)
}

// AddDependency adds a dependency between two nodes.
// If either node doesn't exist an error is returned.
func (g *Graph) AddDependency(from, to string) error {
	if !g.HasNode(from) {
		return nonExistentNodeError(from)
	}
	if !g.HasNode(to) {
		return nonExistentNodeError(to)
	}

	if !slices.Contains(g.outgoingEdges[from], to) {
		g.outgoingEdges[from] = append(g.outgoingEdges[from], to)
	}

	if !slices.Contains(g.incomingEdges[to], from) {
		g.incomingEdges[to] = append(g.incomingEdges[to], from)
	}

	return nil
}

// RemoveDependency removes a dependency between two nodes.
// If either node doesn't exist no error is returned.
func (g *Graph) RemoveDependency(from, to string) {
	if g.HasNode(from) {
		g.outgoingEdges[from] = tfslices.RemoveAll(g.outgoingEdges[from], to)
	}

	if g.HasNode(to) {
		g.incomingEdges[to] = tfslices.RemoveAll(g.incomingEdges[to], from)
	}
}

// DirectDependenciesOf returns the nodes that are the direct dependencies of the specified node.
// Returns an error if the specified node doesn't exist.
func (g *Graph) DirectDependenciesOf(s string) ([]string, error) {
	if !g.HasNode(s) {
		return nil, nonExistentNodeError(s)
	}

	return g.outgoingEdges[s], nil
}

// DirectDependentsOf returns the nodes that directly depend on the specified node.
// Returns an error if the specified node doesn't exist.
func (g *Graph) DirectDependentsOf(s string) ([]string, error) {
	if !g.HasNode(s) {
		return nil, nonExistentNodeError(s)
	}

	return g.incomingEdges[s], nil
}

// DependenciesOf returns the nodes that the specified node depends on (transitively).
// Returns an error if the specified node doesn't exist or a dependency cycle is detected.
func (g *Graph) DependenciesOf(s string) ([]string, error) {
	if !g.HasNode(s) {
		return nil, nonExistentNodeError(s)
	}

	dfs := depthFirstSearch(g.outgoingEdges)
	result, err := dfs(s)

	if err != nil {
		return nil, err
	}

	return tfslices.RemoveAll(result, s), nil
}

// DependentsOf returns the nodes that depend on the specified node (transitively).
// Returns an error if the specified node doesn't exist or a dependency cycle is detected.
func (g *Graph) DependentsOf(s string) ([]string, error) {
	if !g.HasNode(s) {
		return nil, nonExistentNodeError(s)
	}

	dfs := depthFirstSearch(g.incomingEdges)
	result, err := dfs(s)

	if err != nil {
		return nil, err
	}

	return tfslices.RemoveAll(result, s), nil
}

// OverallOrder returns the overall processing order for the dependency graph.
// Returns an error if a dependency cycle is detected.
func (g *Graph) OverallOrder() ([]string, error) {
	// Look for cycles.
	cycleDfs := depthFirstSearch(g.outgoingEdges)

	for _, node := range g.nodes {
		if _, err := cycleDfs(node); err != nil {
			return nil, err
		}
	}

	order := make([]string, 0)

	if g.Len() != 0 {
		dfs := depthFirstSearch(g.outgoingEdges)

		// Find all potential starting points (nodes with nothing depending on them)
		// and run the DFS starting at each of these points.
		for _, node := range g.nodes {
			if len(g.incomingEdges[node]) == 0 {
				result, err := dfs(node)

				if err != nil {
					return nil, err
				}

				order = append(order, result...)
			}
		}
	}

	return order, nil
}

// depthFirstSearch returns a Topological Sort using Depth-First-Search on a set of edges.
// Returns an error if a dependency cycle is detected.
func depthFirstSearch(edges map[string][]string) func(s string) ([]string, error) {
	type todoValue struct {
		node      string
		processed bool
	}

	visited := make([]string, 0)

	return func(s string) ([]string, error) {
		results := make([]string, 0)

		if slices.Contains(visited, s) {
			return results, nil
		}

		inCurrentPath := make(map[string]struct{})
		currentPath := make([]string, 0)
		todo := newStack()

		todo.push(&todoValue{
			node: s,
		})

		for todo.len() > 0 {
			current := todo.peek().(*todoValue)
			node := current.node

			if !current.processed {
				// Visit edges.
				if slices.Contains(visited, node) {
					todo.pop()

					continue
				}

				if _, ok := inCurrentPath[node]; ok {
					return nil, dependencyCycleError(append(currentPath, node))
				}

				inCurrentPath[node] = struct{}{}
				currentPath = append(currentPath, node)

				nodeEdges := edges[node]

				for i := len(nodeEdges) - 1; i >= 0; i-- {
					todo.push(&todoValue{
						node: nodeEdges[i],
					})
				}

				current.processed = true
			} else {
				// Edges have been visited.
				// Unroll the stack.
				todo.pop()
				if n := len(currentPath); n > 0 {
					currentPath = currentPath[:n-1]
				}
				delete(inCurrentPath, node)
				visited = append(visited, node)
				results = append(results, node)
			}
		}

		return results, nil
	}
}

func dependencyCycleError(path []string) error {
	return fmt.Errorf("dependency cycle: %s", strings.Join(path, " -> "))
}

func nonExistentNodeError(s string) error {
	return fmt.Errorf("node does not exist: %s", s)
}
