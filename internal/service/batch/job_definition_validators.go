// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	"context"
	"fmt"

	awsTypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type JobDefinitionTypeValidator struct{}

func (v JobDefinitionTypeValidator) Description(ctx context.Context) string {
	return "validates based on the type that certain properties are set"
}

func (v JobDefinitionTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// When the type is set to multi, check to see if another attribute is set. If so, this is an invalid configuration.
func multiTypeIsSetError[T any](ctx context.Context, config tfsdk.Config, elementType string, attribute string) (diags diag.Diagnostics) {
	var listValue fwtypes.ListNestedObjectValueOf[T]
	diags.Append(config.GetAttribute(ctx, path.Root(attribute), &listValue)...)
	if diags.HasError() {
		return diags
	}

	if !listValue.IsNull() && !listValue.IsUnknown() && len(listValue.Elements()) > 0 {
		diags.AddAttributeError(
			path.Root(attribute),
			"Invalid configuration",
			fmt.Sprintf("No `%s` can be specified when `type` is \"%s\"", attribute, elementType),
		)
	}

	return diags
}

func (v JobDefinitionTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var typeValue types.String
	diags := req.Config.GetAttribute(ctx, path.Root(names.AttrType), &typeValue)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	switch typeValue.ValueString() {
	case string(awsTypes.JobDefinitionTypeContainer):
		resp.Diagnostics.Append(multiTypeIsSetError[nodePropertiesModel](ctx, req.Config, string(awsTypes.JobDefinitionTypeContainer), "node_properties")...)
		if resp.Diagnostics.HasError() {
			return
		}
	case string(awsTypes.JobDefinitionTypeMultinode):
		resp.Diagnostics.Append(multiTypeIsSetError[containerPropertiesModel](ctx, req.Config, string(awsTypes.JobDefinitionTypeMultinode), "container_properties")...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
