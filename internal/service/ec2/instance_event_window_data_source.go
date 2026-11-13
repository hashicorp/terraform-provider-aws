// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ec2_instance_event_window", name="Instance Event Window")
// @Tags
// @Testing(tagsTest=false)
func newInstanceEventWindowDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &instanceEventWindowDataSource{}, nil
}

type instanceEventWindowDataSource struct {
	framework.DataSourceWithModel[instanceEventWindowDataSourceModel]
}

func (d *instanceEventWindowDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cron_expression": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrFilter: customFiltersBlock(ctx),
			"time_ranges": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[instanceEventWindowTimeRangeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"start_week_day": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WeekDay](),
							Computed:   true,
						},
						"start_hour": schema.Int32Attribute{
							Computed: true,
						},
						"end_week_day": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.WeekDay](),
							Computed:   true,
						},
						"end_hour": schema.Int32Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *instanceEventWindowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data instanceEventWindowDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().EC2Client(ctx)

	input := ec2.DescribeInstanceEventWindowsInput{
		Filters: newCustomFilterListFramework(ctx, data.Filters),
	}

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		input.InstanceEventWindowIds = []string{data.ID.ValueString()}
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	out, err := findInstanceEventWindow(ctx, conn, &input)

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading EC2 Instance Event Window (%s)", data.ID.ValueString()), tfresource.SingularDataSourceFindError("EC2 Instance Event Window", err).Error())
		return
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("InstanceEventWindow"))...)
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *instanceEventWindowDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.AtLeastOneOf(
			path.MatchRoot(names.AttrID),
			path.MatchRoot(names.AttrFilter),
		),
	}
}

type instanceEventWindowDataSourceModel struct {
	framework.WithRegionModel
	CronExpression types.String                                                       `tfsdk:"cron_expression"`
	Filters        customFilters                                                      `tfsdk:"filter"`
	ID             types.String                                                       `tfsdk:"id"`
	Name           types.String                                                       `tfsdk:"name"`
	Tags           tftags.Map                                                         `tfsdk:"tags"`
	TimeRanges     fwtypes.ListNestedObjectValueOf[instanceEventWindowTimeRangeModel] `tfsdk:"time_ranges"`
}
