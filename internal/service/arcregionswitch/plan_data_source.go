// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_arcregionswitch_plan", name="Plan")
// @Tags(identifierAttribute="arn")
// @Region(overrideDeprecated=true)
// @Testing(altRegionTfVars=true)
func newPlanDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &planDataSource{}, nil
}

type planDataSource struct {
	framework.DataSourceWithModel[planDataSourceModel]
}

func (d *planDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Computed: true,
			},
			"execution_role": fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrName: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrOwner: fwschema.StringAttribute{
				Computed: true,
			},
			"primary_region": fwschema.StringAttribute{
				Computed: true,
			},
			"recovery_approach": fwschema.StringAttribute{
				Computed: true,
			},
			"recovery_time_objective_minutes": fwschema.Int32Attribute{
				Computed: true,
			},
			"regions": fwschema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			"updated_at": fwschema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrVersion: fwschema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{},
	}
}

func (d *planDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data planDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ARCRegionSwitchClient(ctx)

	plan, err := findPlanByARN(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, plan, &data), smerr.ID, data.ARN.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, conn, data.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ARN.ValueString())
		return
	}
	setTagsOut(ctx, tags.Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type planDataSourceModel struct {
	framework.WithRegionModel
	planModel
	Tags tftags.Map `tfsdk:"tags"`
}

type planModel struct {
	ARN                          types.String         `tfsdk:"arn"`
	Description                  types.String         `tfsdk:"description"`
	ExecutionRole                types.String         `tfsdk:"execution_role"`
	Name                         types.String         `tfsdk:"name"`
	Owner                        types.String         `tfsdk:"owner"`
	PrimaryRegion                types.String         `tfsdk:"primary_region"`
	RecoveryApproach             types.String         `tfsdk:"recovery_approach"`
	RecoveryTimeObjectiveMinutes types.Int32          `tfsdk:"recovery_time_objective_minutes"`
	Regions                      fwtypes.ListOfString `tfsdk:"regions"`
	UpdatedAt                    timetypes.RFC3339    `tfsdk:"updated_at"`
	Version                      types.String         `tfsdk:"version"`
}
