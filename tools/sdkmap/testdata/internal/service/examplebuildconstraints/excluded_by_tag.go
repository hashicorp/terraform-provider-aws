// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package examplebuildconstraints

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// This file is only built with the generate build tag, so sdkmap must not report this resource.

// @SDKResource("aws_examplebuildconstraints_excluded_by_tag", name="Excluded By Tag")
func resourceExcludedByTag() *schema.Resource {
	return &schema.Resource{}
}
