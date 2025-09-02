// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resiliencehub

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehub/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resiliencePolicySchemaV0(ctx context.Context) schema.Schema {
	requiredObjAttrs := map[string]schema.Attribute{
		"rto": schema.StringAttribute{
			CustomType: timetypes.GoDurationType{},
			Required:   true,
		},
		"rpo": schema.StringAttribute{
			CustomType: timetypes.GoDurationType{},
			Required:   true,
		},
	}

	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"data_location_constraint": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataLocationConstraint](),
				Computed:   true,
				Optional:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"tier": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ResiliencyPolicyTier](),
				Required:   true,
			},
			"estimated_cost_tier": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EstimatedCostTier](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrPolicy: schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				CustomType: fwtypes.NewObjectTypeOf[policyDataV0](ctx),
				Blocks: map[string]schema.Block{
					"az": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[resiliencyObjectiveData](ctx),
						Attributes: requiredObjAttrs,
					},
					"hardware": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[resiliencyObjectiveData](ctx),
						Attributes: requiredObjAttrs,
					},
					"software": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[resiliencyObjectiveData](ctx),
						Attributes: requiredObjAttrs,
					},
					names.AttrRegion: schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[resiliencyObjectiveData](ctx),
						Attributes: map[string]schema.Attribute{
							"rto": schema.StringAttribute{
								CustomType: timetypes.GoDurationType{},
								Optional:   true,
							},
							"rpo": schema.StringAttribute{
								CustomType: timetypes.GoDurationType{},
								Optional:   true,
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

type resourceResiliencyPolicyDataV0 struct {
	DataLocationConstraint fwtypes.StringEnum[awstypes.DataLocationConstraint] `tfsdk:"data_location_constraint"`
	EstimatedCostTier      fwtypes.StringEnum[awstypes.EstimatedCostTier]      `tfsdk:"estimated_cost_tier"`
	Policy                 fwtypes.ObjectValueOf[policyDataV0]                 `tfsdk:"policy"`
	PolicyARN              types.String                                        `tfsdk:"arn"`
	PolicyDescription      types.String                                        `tfsdk:"description"`
	PolicyName             types.String                                        `tfsdk:"name"`
	Tier                   fwtypes.StringEnum[awstypes.ResiliencyPolicyTier]   `tfsdk:"tier"`
	Tags                   tftags.Map                                          `tfsdk:"tags"`
	TagsAll                tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts               timeouts.Value                                      `tfsdk:"timeouts"`
}

type policyDataV0 struct {
	AZ       fwtypes.ObjectValueOf[resiliencyObjectiveData] `tfsdk:"az"`
	Hardware fwtypes.ObjectValueOf[resiliencyObjectiveData] `tfsdk:"hardware"`
	Software fwtypes.ObjectValueOf[resiliencyObjectiveData] `tfsdk:"software"`
	Region   fwtypes.ObjectValueOf[resiliencyObjectiveData] `tfsdk:"region"`
}

func upgradeResiliencyPolicyStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var resiliencyPolicyDataV0 resourceResiliencyPolicyDataV0
	response.Diagnostics.Append(request.State.Get(ctx, &resiliencyPolicyDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	resiliencyPolicyDataV1 := resiliencyPolicyResourceModel{
		DataLocationConstraint: resiliencyPolicyDataV0.DataLocationConstraint,
		EstimatedCostTier:      resiliencyPolicyDataV0.EstimatedCostTier,
		Policy:                 upgradePolicyStateFromV0(ctx, resiliencyPolicyDataV0.Policy, &response.Diagnostics),
		PolicyARN:              resiliencyPolicyDataV0.PolicyARN,
		PolicyDescription:      resiliencyPolicyDataV0.PolicyDescription,
		PolicyName:             resiliencyPolicyDataV0.PolicyName,
		Tier:                   resiliencyPolicyDataV0.Tier,
		Tags:                   resiliencyPolicyDataV0.Tags,
		TagsAll:                resiliencyPolicyDataV0.TagsAll,
		Timeouts:               resiliencyPolicyDataV0.Timeouts,
	}

	response.Diagnostics.Append(response.State.Set(ctx, resiliencyPolicyDataV1)...)
}

func upgradePolicyStateFromV0(ctx context.Context, old fwtypes.ObjectValueOf[policyDataV0], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[policyData] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[policyData](ctx)
	}

	var oldObj policyDataV0
	diags.Append(old.As(ctx, &oldObj, basetypes.ObjectAsOptions{})...)

	newList := []policyData{
		{
			AZ:       upgradeResiliencyObjectiveFromV0(ctx, oldObj.AZ, diags),
			Hardware: upgradeResiliencyObjectiveFromV0(ctx, oldObj.Hardware, diags),
			Software: upgradeResiliencyObjectiveFromV0(ctx, oldObj.Software, diags),
			Region:   upgradeResiliencyObjectiveFromV0(ctx, oldObj.Region, diags),
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}

func upgradeResiliencyObjectiveFromV0(ctx context.Context, old fwtypes.ObjectValueOf[resiliencyObjectiveData], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[resiliencyObjectiveData] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[resiliencyObjectiveData](ctx)
	}

	var oldObj resiliencyObjectiveData
	diags.Append(old.As(ctx, &oldObj, basetypes.ObjectAsOptions{})...)

	newList := []resiliencyObjectiveData{
		{
			Rpo: oldObj.Rpo,
			Rto: oldObj.Rto,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}
