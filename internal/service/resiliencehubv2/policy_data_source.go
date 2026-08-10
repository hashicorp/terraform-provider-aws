// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
			"availability_slo": framework.DataSourceComputedListOfObjectAttribute[availabilitySLOModel](ctx),
			"data_recovery":    framework.DataSourceComputedListOfObjectAttribute[dataRecoveryTargetsModel](ctx),
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: fwschema.StringAttribute{
				Computed: true,
			},
			"multi_az":     framework.DataSourceComputedListOfObjectAttribute[multiAZTargetsModel](ctx),
			"multi_region": framework.DataSourceComputedListOfObjectAttribute[multiRegionTargetsModel](ctx),
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
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

	arn := fwflex.StringValueFromFramework(ctx, data.PolicyARN)
	policy, err := findPolicyByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, policy, data))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, policy.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type policyDataSourceModel struct {
	framework.WithRegionModel
	AvailabilitySlo fwtypes.ListNestedObjectValueOf[availabilitySLOModel]     `tfsdk:"availability_slo"`
	DataRecovery    fwtypes.ListNestedObjectValueOf[dataRecoveryTargetsModel] `tfsdk:"data_recovery"`
	Description     types.String                                              `tfsdk:"description"`
	KMSKeyID        fwtypes.ARN                                               `tfsdk:"kms_key_id"`
	MultiAz         fwtypes.ListNestedObjectValueOf[multiAZTargetsModel]      `tfsdk:"multi_az"`
	MultiRegion     fwtypes.ListNestedObjectValueOf[multiRegionTargetsModel]  `tfsdk:"multi_region"`
	Name            types.String                                              `tfsdk:"name"`
	PolicyARN       types.String                                              `tfsdk:"arn"`
	Tags            tftags.Map                                                `tfsdk:"tags"`
}
