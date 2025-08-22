// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package depgraph

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDependencyGraphAddAndRemoveNodes(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	var expected []string
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
	expected = []string{"d", "b"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependenciesOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"c"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependenciesOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependenciesOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"b"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependentsOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependents. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependentsOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"a", "d"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependents. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependentsOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"b"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependents. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DirectDependentsOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"a"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect direct dependents. Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestDependencyGraphDependenciesAndDependents(t *testing.T) {
	t.Parallel()

	var expected []string
	g := New()

	g.AddNode("a")
	g.AddNode("b")
	g.AddNode("c")
	g.AddNode("d")

	err := g.AddDependency("a", "d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("d", "b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	got, err := g.DependenciesOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"c", "b", "d"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependenciesOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"c"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependenciesOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependenciesOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"c", "b"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependencies. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependentsOf("a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependents. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependentsOf("b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"a", "d"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependents. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependentsOf("c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"a", "d", "b"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependents. Unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = g.DependentsOf("d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected = []string{"a"}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("incorrect dependents. Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestDependencyGraphDetectCycles(t *testing.T) {
	t.Parallel()

	// Detect cycle when all nodes have dependents (incoming edges).
	g := New()

	g.AddNode("a")
	g.AddNode("b")
	g.AddNode("c")
	g.AddNode("d")

	err := g.AddDependency("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("c", "a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("d", "a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = g.OverallOrder()
	if err == nil {
		t.Fatalf("expected error")
	}

	// Add a node with no dependent.
	err = g.AddDependency("d", "a")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

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

	err = g.AddDependency("a_1", "a_2")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b_1", "b_2")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b_2", "b_3")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b_3", "b_1")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	_, err = g.OverallOrder()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestDependencyGraphOverallOrder(t *testing.T) {
	t.Parallel()

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

	err = g.AddDependency("a", "b")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("a", "c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b", "c")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("c", "d")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

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

	err = g.AddDependency("a_1", "a_2")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b_1", "b_2")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	err = g.AddDependency("b_2", "b_3")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	got, err = g.OverallOrder()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if expected := []string{"a_2", "a_1", "b_3", "b_2", "b_1"}; !reflect.DeepEqual(got, expected) {
		t.Fatalf("incorrect overall order. Expected: %v, got: %v", expected, got)
	}
}
