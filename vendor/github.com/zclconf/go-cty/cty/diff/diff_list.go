package diff

import "github.com/zclconf/go-cty/cty"

// diffListsShallow takes two values that must be known, non-null lists of the
// same element type and returns a shallow diff that adds and removes only
// direct list members, even if they are themselves collection- or
// structural-typed values.
func diffListsShallow(old cty.Value, new cty.Value, path cty.Path) Diff {
	var diff Diff

	oldEls := make([]cty.Value, 0, old.LengthInt())
	newEls := make([]cty.Value, 0, old.LengthInt())
	it := old.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		oldEls = append(oldEls, v)
	}
	it = new.ElementIterator()
	for it.Next() {
		_, v := it.Element()
		newEls = append(newEls, v)
	}

	lcs := longestCommonSubsequence(oldEls, newEls)
	op := 0        // position in "old"
	np := 0        // position in "new"
	cp := 0        // position in "lcs"
	ip := int64(0) // current index for diff changes

	path = append(path, nil)
	step := &path[len(path)-1]

	for op < len(oldEls) || np < len(newEls) {

		// We want to produce blocks of removes, adds, and contexts
		// until we run out of items.

		// Elements unique to old are deleted
		for op < len(oldEls) {
			if cp < len(lcs) {
				eq := oldEls[op].Equals(lcs[cp])
				if eq.IsKnown() && eq.True() {
					break
				}
			}
			*step = cty.IndexStep{
				Key: cty.NumberIntVal(ip),
			}
			diff = append(diff, DeleteChange{
				Path:     path.Copy(),
				OldValue: oldEls[op],
			})
			op++
		}

		// Elements unique to new are inserted
		for np < len(newEls) {
			if cp < len(lcs) {
				eq := newEls[np].Equals(lcs[cp])
				if eq.IsKnown() && eq.True() {
					break
				}
			}
			*step = cty.IndexStep{
				Key: cty.NumberIntVal(ip),
			}
			var beforeVal cty.Value
			if op < len(oldEls) {
				beforeVal = oldEls[op]
			} else {
				beforeVal = cty.NullVal(old.Type().ElementType())
			}
			diff = append(diff, InsertChange{
				Path:        path.Copy(),
				NewValue:    newEls[np],
				BeforeValue: beforeVal,
			})
			np++
			ip++
		}

		// Elements common to both are context.
		// For this loop we'll advance all three pointers at once because
		// we expect to be walking through the same elements in all three.
		for cp < len(lcs) && op < len(oldEls) && np < len(newEls) {
			oeq := oldEls[op].Equals(lcs[cp])
			if !(oeq.IsKnown() && oeq.True()) {
				break
			}
			neq := newEls[np].Equals(lcs[cp])
			if !(neq.IsKnown() && neq.True()) {
				break
			}

			*step = cty.IndexStep{
				Key: cty.NumberIntVal(ip),
			}
			diff = append(diff, Context{
				Path:      path.Copy(),
				WantValue: lcs[cp],
			})

			cp++
			op++
			np++
			ip++
		}

	}

	return diff
}
