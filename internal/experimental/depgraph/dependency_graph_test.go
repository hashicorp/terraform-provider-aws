package depgraph

import (
	"reflect"
	"testing"
)

func TestDependencyGraphAddAndRemoveNodes(t *testing.T) {
	g := New()

	if got, expected := g.Len(), 0; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	g.AddNode("a")
	if !g.HasNode("a") {
		t.Fatalf("expected graph to contain a")
	}
	if got, expected := g.Len(), 1; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	g.AddNode("b")
	if !g.HasNode("b") {
		t.Fatalf("expected graph to contain b")
	}
	if got, expected := g.Len(), 2; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	// Add node that's already present.
	g.AddNode("b")
	if got, expected := g.Len(), 2; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	if g.HasNode("c") {
		t.Fatalf("expected graph not to contain c")
	}

	g.RemoveNode("b")
	if g.HasNode("b") {
		t.Fatalf("expected graph not to contain b")
	}
	if got, expected := g.Len(), 1; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}

	// Remove node that's not present.
	g.RemoveNode("b")
	if got, expected := g.Len(), 1; got != expected {
		t.Fatalf("incorrect length. Expected: %d, got: %d", expected, got)
	}
}

func TestDependencyGraphDirectDependenciesAndDependents(t *testing.T) {
	g := New()

	if err := g.AddDependency("a", "d"); err == nil {
		t.Fatalf("expected error")
	}

	g.AddNode("a")

	if err := g.AddDependency("a", "d"); err == nil {
		t.Fatalf("expected error")
	}

	g.AddNode("b")
	g.AddNode("c")
	g.AddNode("d")

	if err := g.AddDependency("a", "d"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := g.AddDependency("a", "b"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := g.AddDependency("b", "c"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if err := g.AddDependency("d", "b"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	got, err := g.DirectDependenciesOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"d", "b"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependenciesOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"c"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependenciesOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependenciesOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"b"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependentsOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependents. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependentsOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a", "d"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependents. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependentsOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"b"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependents. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DirectDependentsOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect direct dependents. Expected: %v, got: %v", expected, got)
	}
}

func TestDependencyGraphDependenciesAndDependents(t *testing.T) {
	g := New()

	g.AddNode("a")
	g.AddNode("b")
	g.AddNode("c")
	g.AddNode("d")

	g.AddDependency("a", "d")
	g.AddDependency("a", "b")
	g.AddDependency("b", "c")
	g.AddDependency("d", "b")

	got, err := g.DependenciesOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"c", "b", "d"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependenciesOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"c"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependenciesOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependenciesOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"c", "b"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependencies. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependentsOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependents. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependentsOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a", "d"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependents. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependentsOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a", "d", "b"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependents. Expected: %v, got: %v", expected, got)
	}

	got, err = g.DependentsOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect dependents. Expected: %v, got: %v", expected, got)
	}
}

func TestDependencyGraphDetectCycles(t *testing.T) {
	// Detect cycle when all nodes have dependents (incoming edges).
	g := New()

	g.AddNode("a")
	g.AddNode("b")
	g.AddNode("c")
	g.AddNode("d")

	g.AddDependency("a", "b")
	g.AddDependency("b", "c")
	g.AddDependency("c", "a")
	g.AddDependency("d", "a")

	_, err := g.OverallOrder()
	if err == nil {
		t.Fatalf("expected error")
	}

	// Add a node with no dependent.
	g.AddDependency("d", "a")

	_, err = g.OverallOrder()
	if err == nil {
		t.Fatalf("expected error")
	}

	// Detect cycle when there are several disconnected subgraphs (including one that does not have a cycle).
	g = New()

	g.AddNode("a_1")
	g.AddNode("a_2")
	g.AddNode("b_1")
	g.AddNode("b_2")
	g.AddNode("b_3")

	g.AddDependency("a_1", "a_2")
	g.AddDependency("b_1", "b_2")
	g.AddDependency("b_2", "b_3")
	g.AddDependency("b_3", "b_1")

	_, err = g.OverallOrder()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestDependencyGraphOverallOrder(t *testing.T) {
	g := New()

	got, err := g.OverallOrder()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect overall order. Expected: %v, got: %v", expected, got)
	}

	g.AddNode("a")
	g.AddNode("b")
	g.AddNode("c")
	g.AddNode("d")
	g.AddNode("e")

	g.AddDependency("a", "b")
	g.AddDependency("a", "c")
	g.AddDependency("b", "c")
	g.AddDependency("c", "d")

	got, err = g.OverallOrder()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"d", "c", "b", "a", "e"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect overall order. Expected: %v, got: %v", expected, got)
	}

	// Several disconnected subgraphs.
	g = New()

	g.AddNode("a_1")
	g.AddNode("a_2")
	g.AddNode("b_1")
	g.AddNode("b_2")
	g.AddNode("b_3")

	g.AddDependency("a_1", "a_2")
	g.AddDependency("b_1", "b_2")
	g.AddDependency("b_2", "b_3")

	got, err = g.OverallOrder()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a_2", "a_1", "b_3", "b_2", "b_1"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect overall order. Expected: %v, got: %v", expected, got)
	}
}
