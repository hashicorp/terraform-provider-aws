// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_resiliencehubv2_policy", name="Policy")
// @Tags(identifierAttribute="arn")
func newPolicyDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &policyDataSource{}, nil
}

type policyDataSource struct {
	framework.DataSourceWithModel[policyDataSourceModel]
}

func (d *policyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: fwschema.StringAttribute{
				Required: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"availability_slo": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[availabilitySloModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"target": fwschema.Float64Attribute{
							Computed: true,
						},
					},
				},
			},
			"data_recovery": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataRecoveryModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"time_between_backups_in_minutes": fwschema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
			"multi_az": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiAzModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"disaster_recovery_approach": fwschema.StringAttribute{
							Computed: true,
						},
						"rpo_in_minutes": fwschema.Int32Attribute{
							Computed: true,
						},
						"rto_in_minutes": fwschema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
			"multi_region": fwschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[multiRegionModel](ctx),
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"disaster_recovery_approach": fwschema.StringAttribute{
							Computed: true,
						},
						"rpo_in_minutes": fwschema.Int32Attribute{
							Computed: true,
						},
						"rto_in_minutes": fwschema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *policyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data policyDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ResilienceHubV2Client(ctx)

	policy, err := findPolicyByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, policy, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	data.ARN = types.StringValue(aws.ToString(policy.PolicyArn))

	tags, err := listTags(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type policyDataSourceModel struct {
	framework.WithRegionModel
	ARN             types.String                                          `tfsdk:"arn"`
	AvailabilitySlo fwtypes.ListNestedObjectValueOf[availabilitySloModel] `tfsdk:"availability_slo"`
	DataRecovery    fwtypes.ListNestedObjectValueOf[dataRecoveryModel]    `tfsdk:"data_recovery"`
	Description     types.String                                          `tfsdk:"description"`
	MultiAz         fwtypes.ListNestedObjectValueOf[multiAzModel]         `tfsdk:"multi_az"`
	MultiRegion     fwtypes.ListNestedObjectValueOf[multiRegionModel]     `tfsdk:"multi_region"`
	Name            types.String                                          `tfsdk:"name"`
	Tags            tftags.Map                                            `tfsdk:"tags"`
}
