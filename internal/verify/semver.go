// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	gversion "github.com/hashicorp/go-version"
)

// SemVerLessThan returns whether or not the first version string is less than the second
// according to Semantic Versioning rules (https://semver.org/).
func SemVerLessThan(s1, s2 string) bool {
	v1, v2, err := parseVersions(s1, s2)

	if err != nil {
		return false
	}

	return v1.LessThan(v2)
}

// SemVerGreaterThanOrEqual returns whether or not the first version string is greater than or equal
// to the second according to Semantic Versioning rules (https://semver.org/).
func SemVerGreaterThanOrEqual(s1, s2 string) bool {
	v1, v2, err := parseVersions(s1, s2)

	if err != nil {
		return false
	}

	return v1.GreaterThanOrEqual(v2)
}

func parseVersions(s1, s2 string) (*gversion.Version, *gversion.Version, error) {
	v1, err := gversion.NewVersion(s1)

	if err != nil {
		return nil, nil, err
	}

	v2, err := gversion.NewVersion(s2)

	if err != nil {
		return nil, nil, err
	}
	return v1, v2, nil
}
