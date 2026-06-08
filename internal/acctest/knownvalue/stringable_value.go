// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package knownvalue

import "github.com/hashicorp/terraform-plugin-testing/knownvalue"

func StringExact[T ~string](value T) knownvalue.Check {
	return knownvalue.StringExact(string(value))
}
