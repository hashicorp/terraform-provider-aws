package include

import (
	"fmt"

	"github.com/hashicorp/hcl2/hcl"
)

// MapResolver returns a Resolver that consults the given map for preloaded
// bodies (the values) associated with static include paths (the keys).
//
// An error diagnostic is returned if a path is requested that does not appear
// as a key in the given map.
func MapResolver(m map[string]hcl.Body) Resolver {
	return ResolverFunc(func(path string, refRange hcl.Range) (hcl.Body, hcl.Diagnostics) {
		if body, ok := m[path]; ok {
			return body, nil
		}

		return nil, hcl.Diagnostics{
			{
				Severity: hcl.DiagError,
				Summary:  "Invalid include path",
				Detail:   fmt.Sprintf("The include path %q is not recognized.", path),
				Subject:  &refRange,
			},
		}
	})
}
