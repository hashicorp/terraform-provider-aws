// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sync

import (
	"sync"
)

// Map is a type-safe wrapper around sync.Map from the Go standard library
type Map[K comparable, V any] struct {
	m sync.Map
}

// Load returns the value stored in the map for a key, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (m *Map[K, V]) Load(k K) (V, bool) {
	if a, b := m.m.Load(k); b {
		return a.(V), true
	} else {
		var zero V
		return zero, false
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *Map[K, V]) LoadOrStore(k K, v V) (V, bool) {
	a, b := m.m.LoadOrStore(k, v)
	return a.(V), b
}
