// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"slices"
	"strings"
)

const (
	endpointSeparator = ","
)

// sortEndpointsString sorts a comma-separated list of endpoints.
func sortEndpointsString(s string) string {
	parts := strings.Split(s, endpointSeparator)
	slices.Sort(parts)
	return strings.Join(parts, endpointSeparator)
}
