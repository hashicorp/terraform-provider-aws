package set

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// TestBasicSetOps tests the fundamental operations, whose implementations operate
// directly on the underlying data structure. The remaining operations are implemented
// in terms of these.
func TestBasicSetOps(t *testing.T) {
	s := NewSet(testRules{})
	want := map[int][]interface{}{}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("new set has unexpected contents %#v; want %#v", s.vals, want)
	}
	s.Add(1)
	want[1] = []interface{}{1}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Add(1) set has unexpected contents %#v; want %#v", s.vals, want)
	}
	if !s.Has(1) {
		t.Fatalf("s.Has(1) returned false; want true")
	}
	s.Add(2)
	want[2] = []interface{}{2}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Add(2) set has unexpected contents %#v; want %#v", s.vals, want)
	}
	if !s.Has(2) {
		t.Fatalf("s.Has(2) returned false; want true")
	}

	// Our testRules cause 17 and 33 to return the same hash value as 1, so we can use this
	// to test the situation where multiple values are in a bucket.
	if s.Has(17) {
		t.Fatalf("s.Has(17) returned true; want false")
	}
	s.Add(17)
	s.Add(33)
	want[1] = append(want[1], 17, 33)
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Add(17) and s.Add(33) set has unexpected contents %#v; want %#v", s.vals, want)
	}
	if !s.Has(17) {
		t.Fatalf("s.Has(17) returned false; want true")
	}
	if !s.Has(33) {
		t.Fatalf("s.Has(33) returned false; want true")
	}

	vals := make([]int, 0)
	s.EachValue(func(v interface{}) {
		vals = append(vals, v.(int))
	})
	sort.Ints(vals)
	if want := []int{1, 2, 17, 33}; !reflect.DeepEqual(vals, want) {
		t.Fatalf("wrong values from EachValue %#v; want %#v", vals, want)
	}

	s.Remove(2)
	delete(want, 2)
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(2) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	s.Remove(17)
	want[1] = []interface{}{1, 33}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(17) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	s.Remove(1)
	want[1] = []interface{}{33}
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(1) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	s.Remove(33)
	delete(want, 1)
	if !reflect.DeepEqual(s.vals, want) {
		t.Fatalf("after s.Remove(33) set has unexpected contents %#v; want %#v", s.vals, want)
	}

	vals = make([]int, 0)
	s.EachValue(func(v interface{}) {
		vals = append(vals, v.(int))
	})
	if len(vals) > 0 {
		t.Fatalf("s.EachValue produced values %#v; want no calls", vals)
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		s1         Set
		s2         Set
		wantValues []int
	}{
		{
			NewSet(testRules{}),
			NewSet(testRules{}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSet(testRules{}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{2}),
			[]int{1, 2},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			[]int{1, 17, 33},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{2, 1}),
			[]int{1, 2, 17, 33},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := test.s1.Union(test.s2)
			var gotValues []int
			got.EachValue(func(v interface{}) {
				gotValues = append(gotValues, v.(int))
			})
			sort.Ints(gotValues)
			sort.Ints(test.wantValues)
			if !reflect.DeepEqual(gotValues, test.wantValues) {
				s1Values := test.s1.Values()
				s2Values := test.s2.Values()
				t.Errorf(
					"wrong result %#v for %#v union %#v; want %#v",
					gotValues,
					s1Values,
					s2Values,
					test.wantValues,
				)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		s1         Set
		s2         Set
		wantValues []int
	}{
		{
			NewSet(testRules{}),
			NewSet(testRules{}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSet(testRules{}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{2}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1, 17}),
			NewSetFromSlice(testRules{}, []interface{}{1, 2, 3}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{3, 2, 1}),
			NewSetFromSlice(testRules{}, []interface{}{1, 2, 3}),
			[]int{1, 2, 3},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{2, 1}),
			nil,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := test.s1.Intersection(test.s2)
			var gotValues []int
			got.EachValue(func(v interface{}) {
				gotValues = append(gotValues, v.(int))
			})
			sort.Ints(gotValues)
			sort.Ints(test.wantValues)
			if !reflect.DeepEqual(gotValues, test.wantValues) {
				s1Values := test.s1.Values()
				s2Values := test.s2.Values()
				t.Errorf(
					"wrong result %#v for %#v intersection %#v; want %#v",
					gotValues,
					s1Values,
					s2Values,
					test.wantValues,
				)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		s1         Set
		s2         Set
		wantValues []int
	}{
		{
			NewSet(testRules{}),
			NewSet(testRules{}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSet(testRules{}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{2}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1, 17}),
			NewSetFromSlice(testRules{}, []interface{}{1, 2, 3}),
			[]int{17},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{3, 2, 1}),
			NewSetFromSlice(testRules{}, []interface{}{1, 2, 3}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			[]int{17, 33},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{2, 1}),
			[]int{17, 33},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := test.s1.Subtract(test.s2)
			var gotValues []int
			got.EachValue(func(v interface{}) {
				gotValues = append(gotValues, v.(int))
			})
			sort.Ints(gotValues)
			sort.Ints(test.wantValues)
			if !reflect.DeepEqual(gotValues, test.wantValues) {
				s1Values := test.s1.Values()
				s2Values := test.s2.Values()
				t.Errorf(
					"wrong result %#v for %#v subtract %#v; want %#v",
					gotValues,
					s1Values,
					s2Values,
					test.wantValues,
				)
			}
		})
	}
}

func TestSymmetricDifference(t *testing.T) {
	tests := []struct {
		s1         Set
		s2         Set
		wantValues []int
	}{
		{
			NewSet(testRules{}),
			NewSet(testRules{}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSet(testRules{}),
			[]int{1},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{2}),
			[]int{1, 2},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{1, 17}),
			NewSetFromSlice(testRules{}, []interface{}{1, 2, 3}),
			[]int{2, 3, 17},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{3, 2, 1}),
			NewSetFromSlice(testRules{}, []interface{}{1, 2, 3}),
			nil,
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{1}),
			[]int{1, 17, 33},
		},
		{
			NewSetFromSlice(testRules{}, []interface{}{17, 33}),
			NewSetFromSlice(testRules{}, []interface{}{2, 1}),
			[]int{1, 2, 17, 33},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			got := test.s1.SymmetricDifference(test.s2)
			var gotValues []int
			got.EachValue(func(v interface{}) {
				gotValues = append(gotValues, v.(int))
			})
			sort.Ints(gotValues)
			sort.Ints(test.wantValues)
			if !reflect.DeepEqual(gotValues, test.wantValues) {
				s1Values := test.s1.Values()
				s2Values := test.s2.Values()
				t.Errorf(
					"wrong result %#v for %#v symmetric difference %#v; want %#v",
					gotValues,
					s1Values,
					s2Values,
					test.wantValues,
				)
			}
		})
	}
}
