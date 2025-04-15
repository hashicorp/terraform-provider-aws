// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dns

import (
	"slices"
	"strings"
)

// Reverse switches a DNS hostname to reverse DNS and vice-versa.
func Reverse(hostname string) string {
	parts := strings.Split(hostname, ".")
	slices.Reverse(parts)

	return strings.Join(parts, ".")
}
