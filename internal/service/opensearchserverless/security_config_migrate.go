// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func securityConfigSchemaV0() schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"config_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SecurityConfigType](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"saml_options": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				Attributes: map[string]schema.Attribute{
					"group_attribute": schema.StringAttribute{
						Optional: true,
					},
					"metadata": schema.StringAttribute{
						Required: true,
					},
					"session_timeout": schema.Int64Attribute{
						Optional: true,
						Computed: true,
					},
					"user_attribute": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

type resourceSecurityConfigDataV0 struct {
	ID            types.String                                    `tfsdk:"id"`
	ConfigVersion types.String                                    `tfsdk:"config_version"`
	Description   types.String                                    `tfsdk:"description"`
	Name          types.String                                    `tfsdk:"name"`
	SamlOptions   types.Object                                    `tfsdk:"saml_options"`
	Type          fwtypes.StringEnum[awstypes.SecurityConfigType] `tfsdk:"type"`
}

func upgradeSecurityConfigStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var securityConfigDataV0 resourceSecurityConfigDataV0
	response.Diagnostics.Append(request.State.Get(ctx, &securityConfigDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityConfigDataV1 := securityConfigResourceModel{
		ID:            securityConfigDataV0.ID,
		ConfigVersion: securityConfigDataV0.ConfigVersion,
		Description:   securityConfigDataV0.Description,
		Name:          securityConfigDataV0.Name,
		SamlOptions:   upgradeSAMLOptionsStateFromV0(ctx, securityConfigDataV0.SamlOptions, &response.Diagnostics),
		Type:          securityConfigDataV0.Type,
	}

	response.Diagnostics.Append(response.State.Set(ctx, securityConfigDataV1)...)
}

func upgradeSAMLOptionsStateFromV0(ctx context.Context, old types.Object, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[samlOptionsData] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[samlOptionsData](ctx)
	}

	var oldObj samlOptionsData
	diags.Append(old.As(ctx, &oldObj, basetypes.ObjectAsOptions{})...)

	newList := []samlOptionsData{
		{
			GroupAttribute: oldObj.GroupAttribute,
			Metadata:       oldObj.Metadata,
			SessionTimeout: oldObj.SessionTimeout,
			UserAttribute:  oldObj.UserAttribute,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}
