// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
)

// suppressEqualCIDRBlockDiffs provides custom difference suppression for CIDR blocks
// that have different string values but represent the same CIDR.
func suppressEqualCIDRBlockDiffs(k, old, new string, d *schema.ResourceData) bool {
	return types.CIDRBlocksEqual(old, new)
}
