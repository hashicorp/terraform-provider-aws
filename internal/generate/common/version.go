// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"

	"github.com/hashicorp/go-version"
)

func VersionDecrementMinor(v *version.Version) (*version.Version, error) {
	segments := v.Segments()
	if segments[1] == 0 {
		return nil, fmt.Errorf("minor version is zero, cannot decrement: %s", v.String())
	}

	newSegments := []int{segments[0], segments[1] - 1}
	newVersionStr := fmt.Sprintf("%d.%d", newSegments[0], newSegments[1])

	return version.NewVersion(newVersionStr)
}
