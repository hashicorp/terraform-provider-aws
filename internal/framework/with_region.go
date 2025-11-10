// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WithRegionModel struct {
	Region types.String `tfsdk:"region"`
}
