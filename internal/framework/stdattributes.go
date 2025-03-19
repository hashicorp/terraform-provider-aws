// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func IDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func IDAttributeDeprecatedWithAlternate(altPath path.Path) schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		DeprecationMessage: fmt.Sprintf("Use %s instead. This attribute will be removed in a future verion of the provider.", altPath.String()),
	}
}

func IDAttributeDeprecatedNoReplacement() schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
		DeprecationMessage: "This attribute will be removed in a future verion of the provider.",
	}
}

func ARNAttributeComputedOnly() schema.StringAttribute {
	return schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}
