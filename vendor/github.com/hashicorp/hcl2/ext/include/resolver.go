package include

import (
	"github.com/hashicorp/hcl2/hcl"
)

// A Resolver maps an include path (an arbitrary string, but usually something
// filepath-like) to a hcl.Body.
//
// The parameter "refRange" is the source range of the expression in the calling
// body that provided the given path, for use in generating "invalid path"-type
// diagnostics.
//
// If the returned body is nil, it will be ignored.
//
// Any returned diagnostics will be emitted when content is requested from the
// final composed body (after all includes have been dealt with).
type Resolver interface {
	ResolveBodyPath(path string, refRange hcl.Range) (hcl.Body, hcl.Diagnostics)
}

// ResolverFunc is a function type that implements Resolver.
type ResolverFunc func(path string, refRange hcl.Range) (hcl.Body, hcl.Diagnostics)

// ResolveBodyPath is an implementation of Resolver.ResolveBodyPath.
func (f ResolverFunc) ResolveBodyPath(path string, refRange hcl.Range) (hcl.Body, hcl.Diagnostics) {
	return f(path, refRange)
}
