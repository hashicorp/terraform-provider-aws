// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package querycheck

import (
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func KnownValueCheck(path tfjsonpath.Path, check knownvalue.Check) querycheck.KnownValueCheck {
	return querycheck.KnownValueCheck{
		Path:       path,
		KnownValue: check,
	}
}
