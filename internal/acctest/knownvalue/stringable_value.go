// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package statecheck

import "github.com/hashicorp/terraform-plugin-testing/knownvalue"

func StringExact[T ~string](value T) knownvalue.Check {
	return knownvalue.StringExact(string(value))
}
