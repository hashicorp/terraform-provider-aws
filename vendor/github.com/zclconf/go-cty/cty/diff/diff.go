package diff

import (
	"errors"

	"github.com/zclconf/go-cty/cty"
)

// Diff represents a sequence of changes to transform one Value into another
// Value.
//
// Diff has convenience methods for appending changes one by one, but each
// of these allocates a fresh list and so they may create memory pressure.
// Callers can and should construct and convert []Change values directly
// where appropriate.
type Diff []Change

// NewDiff produces a new diff by comparing the two given values. the
// returned diff represents a sequence of changes required to transform
// a value equal to source into a value equal to target.
//
// Since cty unknown values are never equal, any unknown values will
// *always* produce a ReplaceChange, even if the source and target values
// are both unknown. This represents the fact that we don't know yet whether
// the value will actually change.
//
// For best results the source value should be free of unknowns, since an
// entirely-deterministic diff cannot be created when the source contains
// nested unknown values. However, this function will
// still attempt to construct such a diff since it may still be useful to
// display to a user.
func NewDiff(source, target cty.Value) Diff {
	panic("NewDiff is not yet implemented")
}

// Apply produces a new value by applying the receiving Diff to the given
// source value. If any one change fails then the entire operation is
// considered to have failed.
func (d Diff) Apply(source cty.Value) (cty.Value, error) {
	return cty.NullVal(source.Type()), errors.New("not yet implemented")
}

// Replace returns a copy of the receiver with a ReplaceChange appended.
func (d Diff) Replace(path cty.Path, old, new cty.Value) Diff {
	return d.append(ReplaceChange{
		Path:     path,
		OldValue: old,
		NewValue: new,
	})
}

// Delete returns a copy of the receiver with a DeleteChange appended.
func (d Diff) Delete(path cty.Path, old cty.Value) Diff {
	return d.append(DeleteChange{
		Path:     path,
		OldValue: old,
	})
}

// Insert returns a copy of the receiver with a DeleteChange appended.
func (d Diff) Insert(path cty.Path, new, before cty.Value) Diff {
	return d.append(InsertChange{
		Path:        path,
		NewValue:    new,
		BeforeValue: before,
	})
}

// Add returns a copy of the receiver with an AddChange appended.
func (d Diff) Add(path cty.Path, new cty.Value) Diff {
	return d.append(AddChange{
		Path:     path,
		NewValue: new,
	})
}

// Remove returns a copy of the receiver with a RemoveChange appended.
func (d Diff) Remove(path cty.Path, old cty.Value) Diff {
	return d.append(RemoveChange{
		Path:     path,
		OldValue: old,
	})
}

func (d Diff) append(change Change) Diff {
	ret := make(Diff, len(d)+1)
	copy(ret, d)
	ret[len(d)] = change
	return ret
}
