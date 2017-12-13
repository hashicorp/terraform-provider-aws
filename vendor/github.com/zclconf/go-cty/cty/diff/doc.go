// package diff contains utilities for computing and applying structural
// diffs against data structures within the cty type system.
//
// This can be useful for showing a user how a cty-based configuration
// structure has changed over time, using concepts that are congruent with the
// way the values themselves are expressed.
//
// This package does not contain any functions for rendering a diff for
// display to a user, since it is expected that a language building on cty
// will want to use its own familiar syntax for presenting diffs and will thus
// provide its own diff-presentation functionality.
package diff
