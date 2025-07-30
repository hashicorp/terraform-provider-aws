// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func AWSAccountIDAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Computed: true,
		Validators: []validator.String{
			fwvalidators.AWSAccountID(),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplace(),
		},
	}
}
