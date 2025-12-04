// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import (
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

// StringLegacyNull can be used to show intention when matching an empty string in Legacy null handling,
// i.e. for resource types implemented with or ported from SDKv2
// Note that not all empty string values have a `null` intent, e.g. enum values
func StringLegacyNull() knownvalue.Check {
	return knownvalue.StringExact("")
}
