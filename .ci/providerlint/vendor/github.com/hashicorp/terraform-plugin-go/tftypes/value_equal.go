// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import (
	"errors"
	"fmt"
	"math/big"
)

// deepEqual walks both Value to ensure any underlying Value are equal. This
// logic is essentially a duplicate of Diff, however it is intended to return
// early on any inequality and avoids memory allocations where possible.
//
// There might be ways to better share the internal logic of this method with
// Diff, however that effort is reserved for a time when the effort is justified
// over resolving the inherent compute and memory performance issues with Diff
// when only checking for inequality.
func (val1 Value) deepEqual(val2 Value) (bool, error) {
	if val1.Type() == nil && val2.Type() == nil && val1.value == nil && val2.value == nil {
		return false, nil
	}

	if (val1.Type() == nil && val2.Type() != nil) || (val1.Type() != nil && val2.Type() == nil) {
		return false, errors.New("cannot diff value missing type")
	}

	if !val1.Type().Is(val2.Type()) {
		return false, errors.New("Can't diff values of different types")
	}

	// Capture walk differences for returning early
	var hasDiff bool

	// make sure everything in val2 is also in val1
	err := Walk(val2, func(path *AttributePath, _ Value) (bool, error) {
		_, _, err := val1.walkAttributePath(path)

		if err != nil && err != ErrInvalidStep {
			return false, fmt.Errorf("Error walking %q: %w", path, err)
		} else if err == ErrInvalidStep {
			hasDiff = true

			return false, stopWalkError
		}

		return true, nil
	})

	if err != nil {
		return false, err
	}

	if hasDiff {
		return false, nil
	}

	// make sure everything in val1 is also in val2 and also that it all matches
	err = Walk(val1, func(path *AttributePath, value1 Value) (bool, error) {
		// pull out the Value at the same path in val2
		value2, _, err := val2.walkAttributePath(path)

		if err != nil && err != ErrInvalidStep {
			return false, fmt.Errorf("Error walking %q: %w", path, err)
		} else if err == ErrInvalidStep {
			hasDiff = true

			return false, stopWalkError
		}

		// if they're both unknown, no need to continue
		if !value1.IsKnown() && !value2.IsKnown() {
			return false, nil
		}

		if value1.IsKnown() != value2.IsKnown() {
			hasDiff = true

			return false, stopWalkError
		}

		// if they're both null, no need to continue
		if value1.IsNull() && value2.IsNull() {
			return false, nil
		}

		if value1.IsNull() != value2.IsNull() {
			hasDiff = true

			return false, stopWalkError
		}

		// We know there are known, non-null values, time to compare them.
		// Since this logic is very hot path, it is optimized to use type and
		// value implementation details rather than Equal() and As()
		// respectively, since both result in memory allocations.
		switch typ := value1.Type().(type) {
		case primitive:
			switch typ.name {
			case String.name:
				s1, ok := value1.value.(string)

				if !ok {
					return false, fmt.Errorf("cannot convert %T into string", value1.value)
				}

				s2, ok := value2.value.(string)

				if !ok {
					return false, fmt.Errorf("cannot convert %T into string", value2.value)
				}

				if s1 != s2 {
					hasDiff = true

					return false, stopWalkError
				}
			case Number.name:
				n1, ok := value1.value.(*big.Float)

				if !ok {
					return false, fmt.Errorf("cannot convert %T into *big.Float", value1.value)
				}

				n2, ok := value2.value.(*big.Float)

				if !ok {
					return false, fmt.Errorf("cannot convert %T into *big.Float", value2.value)
				}

				if n1.Cmp(n2) != 0 {
					hasDiff = true

					return false, stopWalkError
				}
			case Bool.name:
				b1, ok := value1.value.(bool)

				if !ok {
					return false, fmt.Errorf("cannot convert %T into bool", value1.value)
				}

				b2, ok := value2.value.(bool)

				if !ok {
					return false, fmt.Errorf("cannot convert %T into bool", value2.value)
				}

				if b1 != b2 {
					hasDiff = true

					return false, stopWalkError
				}
			case DynamicPseudoType.name:
				// Let recursion from the walk check the sub-values match
				return true, nil
			}

			return false, nil
		case List, Set, Tuple:
			s1, ok := value1.value.([]Value)

			if !ok {
				return false, fmt.Errorf("cannot convert %T into []tftypes.Value", value1.value)
			}

			s2, ok := value2.value.([]Value)

			if !ok {
				return false, fmt.Errorf("cannot convert %T into []tftypes.Value", value2.value)
			}

			// we only care about if the lengths match for lists,
			// sets, and tuples. If any of the elements differ,
			// the recursion of the walk will find them for us.
			if len(s1) != len(s2) {
				hasDiff = true

				return false, stopWalkError
			}

			return true, nil
		case Map, Object:
			m1, ok := value1.value.(map[string]Value)

			if !ok {
				return false, fmt.Errorf("cannot convert %T into map[string]tftypes.Value", value1.value)
			}

			m2, ok := value2.value.(map[string]Value)

			if !ok {
				return false, fmt.Errorf("cannot convert %T into map[string]tftypes.Value", value2.value)
			}

			// we only care about if the number of keys match for maps and
			// objects. If any of the elements differ, the recursion of the walk
			// will find them for us.
			if len(m1) != len(m2) {
				hasDiff = true

				return false, stopWalkError
			}

			return true, nil
		}

		return false, fmt.Errorf("unexpected type %v in Diff at %s", value1.Type(), path)
	})

	return !hasDiff, err
}
