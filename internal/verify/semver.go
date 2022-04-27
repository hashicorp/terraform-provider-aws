package verify

import (
	gversion "github.com/hashicorp/go-version"
)

// SemVerLessThan returns whether or not the first version string is less than the second
// according to Semantic Versioning rules (https://semver.org/).
func SemVerLessThan(s1, s2 string) bool {
	v1, err := gversion.NewVersion(s1)

	if err != nil {
		return false
	}

	v2, err := gversion.NewVersion(s2)

	if err != nil {
		return false
	}

	return v1.LessThan(v2)
}
