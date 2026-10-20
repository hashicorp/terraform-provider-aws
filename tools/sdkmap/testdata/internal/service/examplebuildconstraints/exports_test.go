// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplebuildconstraints

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Exports for use in tests only.
var (
	ResourceIncluded = resourceIncluded
)

// This file is a test file, so sdkmap must not report this resource.

// @SDKResource("aws_examplebuildconstraints_excluded_by_test_file", name="Excluded By Test File")
func resourceExcludedByTestFile() *schema.Resource {
	return &schema.Resource{}
}
